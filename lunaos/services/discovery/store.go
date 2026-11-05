// Package discovery syncs adult media from official provider APIs into an
// encrypted local library. This file is its storage layer.
//
// Discovery must never import services/intimacy and never reads vault or
// body-profile assets: its key, database and blobs are entirely its own.
package discovery

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var storeSchema string

var (
	ErrInvalidKey   = errors.New("INVALID_DISCOVERY_KEY")
	ErrHashMismatch = errors.New("DISCOVERY_HASH_MISMATCH")
	ErrBlobCorrupt  = errors.New("DISCOVERY_BLOB_CORRUPT")
	ErrNotFound     = errors.New("DISCOVERY_NOT_FOUND")
)

const chunkSize = 64 * 1024

// Item is one library entry. Creator, SourceURL, Tags and SHA256 are sealed at
// rest; the other fields stay queryable in clear columns.
type Item struct {
	Provider, ExternalID, SourceURL, Creator, License, ContentLevel string
	Tags                                                            []string
	SHA256                                                          string
	Bytes                                                           int64
	FetchedAt                                                       time.Time
}

type Store struct {
	db      *sql.DB
	blobDir string
	nameKey []byte
	aead    cipher.AEAD // metadata
	blob    cipher.AEAD // blobs
}

type sealedMeta struct {
	Creator   string   `json:"creator"`
	SourceURL string   `json:"source_url"`
	Tags      []string `json:"tags"`
	SHA256    string   `json:"sha256"`
}

func subKey(key []byte, label string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(label))
	return m.Sum(nil)
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func OpenStore(dbPath, blobDir, keyPath string) (*Store, error) {
	if dbPath == "" || blobDir == "" || keyPath == "" {
		return nil, fmt.Errorf("invalid discovery store paths")
	}
	info, err := os.Lstat(keyPath)
	if err != nil {
		return nil, ErrInvalidKey
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 || info.Size() != 32 {
		return nil, ErrInvalidKey
	}
	key, err := os.ReadFile(keyPath)
	if err != nil || len(key) != 32 {
		return nil, ErrInvalidKey
	}
	s := &Store{blobDir: blobDir, nameKey: subKey(key, "discovery-name-v1")}
	if s.aead, err = newGCM(subKey(key, "discovery-meta-v1")); err != nil {
		return nil, err
	}
	if s.blob, err = newGCM(subKey(key, "discovery-blob-v1")); err != nil {
		return nil, err
	}
	if err = os.MkdirAll(blobDir, 0700); err != nil {
		return nil, err
	}
	if err = os.Chmod(blobDir, 0700); err != nil {
		return nil, err
	}
	if err = privateDir(blobDir); err != nil {
		return nil, err
	}
	absolute, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, err
	}
	if err = notWritableByOthers(filepath.Dir(absolute)); err != nil {
		return nil, err
	}
	if _, statErr := os.Lstat(absolute); os.IsNotExist(statErr) {
		f, createErr := os.OpenFile(absolute, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
		if createErr != nil {
			return nil, createErr
		}
		_ = f.Close()
	}
	uri := (&url.URL{Scheme: "file", Path: filepath.ToSlash(absolute)}).String()
	db, err := sql.Open("sqlite3", uri+"?_busy_timeout=5000&_journal_mode=WAL&_synchronous=FULL&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err = os.Chmod(absolute, 0600); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err = db.Exec(storeSchema); err != nil {
		_ = db.Close()
		return nil, err
	}
	s.db = db
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) blobName(sha string) string {
	m := hmac.New(sha256.New, s.nameKey)
	m.Write([]byte(strings.ToLower(sha)))
	return hex.EncodeToString(m.Sum(nil))
}

func metaAAD(provider, externalID string) []byte {
	return []byte("meta\x00" + provider + "\x00" + externalID)
}

func (s *Store) sealMeta(m sealedMeta, provider, externalID string) ([]byte, error) {
	plain, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, s.aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return s.aead.Seal(nonce, nonce, plain, metaAAD(provider, externalID)), nil
}

func (s *Store) openMeta(blob []byte, provider, externalID string) (sealedMeta, error) {
	var m sealedMeta
	n := s.aead.NonceSize()
	if len(blob) < n {
		return m, ErrBlobCorrupt
	}
	plain, err := s.aead.Open(nil, blob[:n], blob[n:], metaAAD(provider, externalID))
	if err != nil {
		return m, ErrBlobCorrupt
	}
	if err = json.Unmarshal(plain, &m); err != nil {
		return m, ErrBlobCorrupt
	}
	return m, nil
}

// Blob format: 8 random bytes (per-file nonce prefix) followed by sealed
// chunks. Chunk i uses nonce prefix||uint32(i); its AAD carries the final flag,
// so truncation, reordering and chunk swaps all fail authentication.
func chunkNonce(prefix []byte, i uint32) []byte {
	n := make([]byte, 12)
	copy(n, prefix)
	binary.BigEndian.PutUint32(n[8:], i)
	return n
}

// chunkAAD binds a chunk to the format version, the blob's HMAC file name and
// its final flag.
func chunkAAD(name string, final bool) []byte {
	f := byte(0)
	if final {
		f = 1
	}
	return append([]byte("discovery-blob-v1\x00"+name+"\x00"), f)
}

// Put stores the item. A re-Put of an existing (Provider, ExternalID) returns
// stored=false without consuming media. stored is true when new blob bytes (or, with a nil
// media, a new metadata row) were written.
func (s *Store) Put(ctx context.Context, it Item, media io.Reader) (stored bool, err error) {
	if it.Provider == "" || it.ExternalID == "" {
		return false, fmt.Errorf("provider and external id are required")
	}
	var exists int
	if err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM items WHERE provider=? AND external_id=?`, it.Provider, it.ExternalID).Scan(&exists); err != nil {
		return false, err
	}
	if exists > 0 {
		return false, nil
	}
	meta := sealedMeta{Creator: it.Creator, SourceURL: it.SourceURL, Tags: it.Tags, SHA256: strings.ToLower(it.SHA256)}
	name := ""
	size := it.Bytes
	newBlob := false
	if media != nil {
		var sha string
		var tmp string
		sha, size, tmp, err = s.encryptToTemp(media, "")
		if err != nil {
			return false, err
		}
		defer os.Remove(tmp)
		if it.SHA256 != "" && !strings.EqualFold(it.SHA256, sha) {
			return false, ErrHashMismatch
		}
		meta.SHA256 = sha
		name = s.blobName(sha)
		final := filepath.Join(s.blobDir, name)
		if _, statErr := os.Lstat(final); os.IsNotExist(statErr) {
			// Reseal under the name-bound AAD; plaintext never touches disk.
			src, openErr := s.openBlobFile(tmp, "")
			if openErr != nil {
				return false, openErr
			}
			_, _, bound, sealErr := s.encryptToTemp(src, name)
			_ = src.Close()
			if sealErr != nil {
				return false, sealErr
			}
			defer os.Remove(bound)
			if err = os.Rename(bound, final); err != nil {
				return false, err
			}
			newBlob = true
		} else if statErr != nil {
			return false, statErr
		}
	}
	sealed, err := s.sealMeta(meta, it.Provider, it.ExternalID)
	if err != nil {
		return false, err
	}
	fetched := it.FetchedAt
	if fetched.IsZero() {
		fetched = time.Now()
	}
	res, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO items(provider, external_id, blob_name, content_level, license, bytes, fetched_at, meta) VALUES(?,?,?,?,?,?,?,?)`,
		it.Provider, it.ExternalID, name, it.ContentLevel, it.License, size, fetched.UTC().Unix(), sealed)
	if err != nil {
		if newBlob {
			_ = os.Remove(filepath.Join(s.blobDir, name))
		}
		return false, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return false, nil
	}
	if media == nil {
		return true, nil
	}
	return newBlob, nil
}

func (s *Store) encryptToTemp(media io.Reader, name string) (sha string, size int64, path string, err error) {
	f, err := os.CreateTemp(s.blobDir, ".tmp-*")
	if err != nil {
		return "", 0, "", err
	}
	path = f.Name()
	defer func() {
		_ = f.Close()
		if err != nil {
			_ = os.Remove(path)
		}
	}()
	if err = f.Chmod(0600); err != nil {
		return
	}
	prefix := make([]byte, 8)
	if _, err = rand.Read(prefix); err != nil {
		return
	}
	if _, err = f.Write(prefix); err != nil {
		return
	}
	h := sha256.New()
	br := bufio.NewReaderSize(media, chunkSize+1)
	buf := make([]byte, chunkSize)
	for i := uint32(0); ; i++ {
		var n int
		n, err = io.ReadFull(br, buf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return
		}
		err = nil
		_, peekErr := br.Peek(1)
		final := peekErr != nil
		if peekErr != nil && peekErr != io.EOF {
			err = peekErr
			return
		}
		h.Write(buf[:n])
		size += int64(n)
		ct := s.blob.Seal(nil, chunkNonce(prefix, i), buf[:n], chunkAAD(name, final))
		if _, err = f.Write(ct); err != nil {
			return
		}
		if final {
			break
		}
	}
	if err = f.Sync(); err != nil {
		return
	}
	return hex.EncodeToString(h.Sum(nil)), size, path, nil
}

type blobReader struct {
	f      *os.File
	br     *bufio.Reader
	aead   cipher.AEAD
	name   string
	prefix []byte
	idx    uint32
	plain  []byte
	done   bool
	err    error
}

func (r *blobReader) next() error {
	ct := make([]byte, chunkSize+r.aead.Overhead())
	n, err := io.ReadFull(r.br, ct)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return err
	}
	_, peekErr := r.br.Peek(1)
	final := peekErr == io.EOF
	plain, err := r.aead.Open(nil, chunkNonce(r.prefix, r.idx), ct[:n], chunkAAD(r.name, final))
	if err != nil {
		return ErrBlobCorrupt
	}
	r.idx++
	r.plain = plain
	r.done = final
	return nil
}

func (r *blobReader) Read(p []byte) (int, error) {
	for len(r.plain) == 0 {
		if r.err != nil {
			return 0, r.err
		}
		if r.done {
			return 0, io.EOF
		}
		r.err = r.next()
	}
	n := copy(p, r.plain)
	r.plain = r.plain[n:]
	return n, nil
}

func (r *blobReader) Close() error { return r.f.Close() }

func (s *Store) openBlobFile(path, name string) (*blobReader, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	prefix := make([]byte, 8)
	if _, err = io.ReadFull(f, prefix); err != nil {
		_ = f.Close()
		return nil, ErrBlobCorrupt
	}
	r := &blobReader{f: f, br: bufio.NewReaderSize(f, chunkSize+s.blob.Overhead()+1), aead: s.blob, name: name, prefix: prefix}
	if err = r.next(); err != nil {
		_ = f.Close()
		return nil, err
	}
	return r, nil
}

// verifyReader fails at EOF unless the decrypted bytes hash to the requested
// SHA-256.
type verifyReader struct {
	*blobReader
	h    hash.Hash
	want string
}

func (v *verifyReader) Read(p []byte) (int, error) {
	n, err := v.blobReader.Read(p)
	v.h.Write(p[:n])
	if err == io.EOF && !strings.EqualFold(hex.EncodeToString(v.h.Sum(nil)), v.want) {
		return n, ErrBlobCorrupt
	}
	return n, err
}

// Open returns a reader over the decrypted media. The first chunk is verified
// before returning; a wrong key, tampered or swapped file fails here or on
// read (the name-bound AAD and a final SHA-256 check), never yielding
// unauthenticated media as a clean EOF.
func (s *Store) Open(ctx context.Context, sha256hex string) (io.ReadCloser, error) {
	name := s.blobName(sha256hex)
	r, err := s.openBlobFile(filepath.Join(s.blobDir, name), name)
	if err != nil {
		return nil, err
	}
	return &verifyReader{blobReader: r, h: sha256.New(), want: sha256hex}, nil
}

// Delete removes the row and, when no other row shares it, the blob.
func (s *Store) Delete(ctx context.Context, provider, externalID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var name string
	err = tx.QueryRowContext(ctx, `SELECT blob_name FROM items WHERE provider=? AND external_id=?`, provider, externalID).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM items WHERE provider=? AND external_id=?`, provider, externalID); err != nil {
		return err
	}
	remaining := 0
	if name != "" {
		if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM items WHERE blob_name=?`, name).Scan(&remaining); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	if name != "" && remaining == 0 {
		if err = os.Remove(filepath.Join(s.blobDir, name)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (s *Store) SaveCursor(ctx context.Context, provider, scope, cursor string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO cursors(provider, scope, cursor) VALUES(?,?,?) ON CONFLICT(provider, scope) DO UPDATE SET cursor=excluded.cursor`, provider, scope, cursor)
	return err
}

// Cursor returns "" when none is saved.
func (s *Store) Cursor(ctx context.Context, provider, scope string) (string, error) {
	var c string
	err := s.db.QueryRowContext(ctx, `SELECT cursor FROM cursors WHERE provider=? AND scope=?`, provider, scope).Scan(&c)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return c, err
}

func privateDir(path string) error {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("discovery requires a private non-symlink blob directory")
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); !ok || int(stat.Uid) != os.Getuid() {
		return fmt.Errorf("discovery blob directory owner mismatch")
	}
	return nil
}

func notWritableByOthers(path string) error {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0022 != 0 {
		return fmt.Errorf("discovery database directory must exist and not be group/other-writable")
	}
	return nil
}
