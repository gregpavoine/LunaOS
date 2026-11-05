package intimacy

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// The owner's private adult media library. LUNA indexes folders the owner
// registers (a local directory or a mounted NAS share); it never fetches
// anything. Each file is identified by content: one SHA-256 per distinct image
// or video, however many copies exist. Everything that could reveal the
// collection (folder paths, file names, tags) is sealed with the vault key;
// path lookups use a keyed hash, so the index finds an unchanged file without
// storing its name in clear. A source that goes offline keeps its index and
// marks its items unavailable instead of deleting them. Reading and writing
// the library need their own owner consent scopes.

const (
	ScopeLibraryRead  Scope = "library.read"
	ScopeLibraryWrite Scope = "library.write"

	maxLibraryFileBytes = 64 << 30
	maxLibraryTags      = 32
	maxLibraryResults   = 200
)

var (
	ErrLibrarySource   = errors.New("INVALID_LIBRARY_SOURCE")
	ErrLibraryNotFound = errors.New("LIBRARY_ITEM_NOT_FOUND")
	ErrInvalidLibrary  = errors.New("INVALID_LIBRARY_REQUEST")
)

type LibrarySource struct {
	ID      string `json:"source_id"`
	Path    string `json:"path"`
	Label   string `json:"label"`
	Online  bool   `json:"online"`
	Scanned string `json:"last_scan_utc,omitempty"`
}

type ScanReport struct {
	SourceID    string `json:"source_id"`
	Online      bool   `json:"online"`
	Indexed     int    `json:"indexed"`
	Unchanged   int    `json:"unchanged"`
	Duplicates  int    `json:"duplicates"`
	Skipped     int    `json:"skipped"`
	Unavailable int    `json:"unavailable"`
}

type LibraryItem struct {
	ID        string   `json:"item_id"`
	SourceID  string   `json:"source_id"`
	SHA256    string   `json:"sha256"`
	MediaType string   `json:"media_type"`
	Bytes     int64    `json:"bytes"`
	Name      string   `json:"name"`
	Tags      []string `json:"tags"`
	Available bool     `json:"available"`
	Copies    int      `json:"copies"`
}

type LibraryQuery struct {
	Text      string   `json:"text,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	MediaType string   `json:"media_type,omitempty"`
	Limit     int      `json:"limit,omitempty"`
}

type libraryLocation struct {
	RelativePath string `json:"relative_path"`
	ModTimeNanos int64  `json:"mtime_ns"`
}

// AddLibrarySource registers a folder. It must be an existing absolute
// directory that is not a symlink.
func (v *Vault) AddLibrarySource(ctx context.Context, consentID, path, label string) (LibrarySource, error) {
	if err := v.authorize(ctx, consentID, "owner", ScopeLibraryWrite); err != nil {
		return LibrarySource{}, err
	}
	clean := filepath.Clean(path)
	if !filepath.IsAbs(path) || clean != path || len(label) > 128 {
		return LibrarySource{}, ErrLibrarySource
	}
	info, err := os.Lstat(clean)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return LibrarySource{}, ErrLibrarySource
	}
	id, err := randomID()
	if err != nil {
		return LibrarySource{}, err
	}
	sealed, err := sealWithAEAD(v.aead, mustJSON(map[string]string{"path": clean, "label": label}), librarySourceAAD(id))
	if err != nil {
		return LibrarySource{}, err
	}
	if _, err = v.db.ExecContext(ctx, `INSERT INTO library_source(id,ciphertext,online,created_at_utc) VALUES(?,?,1,?)`, id, sealed, utcNow()); err != nil {
		return LibrarySource{}, err
	}
	return LibrarySource{ID: id, Path: clean, Label: label, Online: true}, nil
}

// ListLibrarySources returns the registered folders.
func (v *Vault) ListLibrarySources(ctx context.Context, consentID string) ([]LibrarySource, error) {
	if err := v.authorize(ctx, consentID, "owner", ScopeLibraryRead); err != nil {
		return nil, err
	}
	rows, err := v.db.QueryContext(ctx, `SELECT id,ciphertext,online,COALESCE(last_scan_utc,'') FROM library_source ORDER BY created_at_utc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sources := []LibrarySource{}
	for rows.Next() {
		var source LibrarySource
		var sealed []byte
		var online int
		if err = rows.Scan(&source.ID, &sealed, &online, &source.Scanned); err != nil {
			return nil, err
		}
		if source.Path, source.Label, err = v.openSource(source.ID, sealed); err != nil {
			return nil, err
		}
		source.Online = online == 1
		sources = append(sources, source)
	}
	return sources, rows.Err()
}

// ScanLibrary indexes new and changed images and videos under a source. Files
// are recognised by their content, not their extension; symlinks are not
// followed, so a link cannot pull files from outside the source into the index.
func (v *Vault) ScanLibrary(ctx context.Context, consentID, sourceID string) (ScanReport, error) {
	if err := v.authorize(ctx, consentID, "owner", ScopeLibraryWrite); err != nil {
		return ScanReport{}, err
	}
	var sealed []byte
	err := v.db.QueryRowContext(ctx, `SELECT ciphertext FROM library_source WHERE id=?`, sourceID).Scan(&sealed)
	if errors.Is(err, sql.ErrNoRows) {
		return ScanReport{}, ErrLibraryNotFound
	}
	if err != nil {
		return ScanReport{}, err
	}
	root, _, err := v.openSource(sourceID, sealed)
	if err != nil {
		return ScanReport{}, err
	}
	report := ScanReport{SourceID: sourceID}
	indexKey, err := v.libraryIndexKey(ctx)
	if err != nil {
		return ScanReport{}, err
	}
	if info, statErr := os.Lstat(root); statErr != nil || !info.IsDir() {
		// An unmounted NAS share or a removed drive: keep the index, say so.
		result, err := v.db.ExecContext(ctx, `UPDATE library_location SET available=0 WHERE source_id=?`, sourceID)
		if err != nil {
			return ScanReport{}, err
		}
		unavailable, _ := result.RowsAffected()
		report.Unavailable = int(unavailable)
		_, err = v.db.ExecContext(ctx, `UPDATE library_source SET online=0,last_scan_utc=? WHERE id=?`, utcNow(), sourceID)
		return report, err
	}
	report.Online = true
	seen := map[string]bool{}
	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			report.Skipped++
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if entry.IsDir() || !entry.Type().IsRegular() {
			return nil
		}
		info, err := entry.Info()
		if err != nil || info.Size() == 0 || info.Size() > maxLibraryFileBytes {
			report.Skipped++
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			report.Skipped++
			return nil
		}
		pathKey := libraryPathKey(indexKey, sourceID, relative)
		seen[pathKey] = true
		var knownSize, knownMod int64
		err = v.db.QueryRowContext(ctx, `SELECT size_bytes,mtime_ns FROM library_location WHERE path_key=?`, pathKey).Scan(&knownSize, &knownMod)
		if err == nil && knownSize == info.Size() && knownMod == info.ModTime().UnixNano() {
			report.Unchanged++
			_, err = v.db.ExecContext(ctx, `UPDATE library_location SET available=1 WHERE path_key=?`, pathKey)
			return err
		}
		mediaType, digest, err := sniffAndHash(path)
		if err != nil || mediaType == "" {
			report.Skipped++
			return nil
		}
		duplicate, err := v.indexFile(ctx, sourceID, pathKey, digest, mediaType, relative, info)
		if err != nil {
			return err
		}
		if duplicate {
			report.Duplicates++
		} else {
			report.Indexed++
		}
		return nil
	})
	if walkErr != nil {
		return ScanReport{}, walkErr
	}
	rows, err := v.db.QueryContext(ctx, `SELECT path_key FROM library_location WHERE source_id=?`, sourceID)
	if err != nil {
		return ScanReport{}, err
	}
	var gone []string
	for rows.Next() {
		var key string
		if err = rows.Scan(&key); err != nil {
			rows.Close()
			return ScanReport{}, err
		}
		if !seen[key] {
			gone = append(gone, key)
		}
	}
	if err = rows.Close(); err != nil {
		return ScanReport{}, err
	}
	for _, key := range gone {
		if _, err = v.db.ExecContext(ctx, `DELETE FROM library_location WHERE path_key=?`, key); err != nil {
			return ScanReport{}, err
		}
	}
	if _, err = v.db.ExecContext(ctx, `DELETE FROM library_item WHERE id NOT IN (SELECT item_id FROM library_location)`); err != nil {
		return ScanReport{}, err
	}
	_, err = v.db.ExecContext(ctx, `UPDATE library_source SET online=1,last_scan_utc=? WHERE id=?`, utcNow(), sourceID)
	return report, err
}

func (v *Vault) indexFile(ctx context.Context, sourceID, pathKey, digest, mediaType, relative string, info fs.FileInfo) (bool, error) {
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var itemID string
	duplicate := true
	err = tx.QueryRowContext(ctx, `SELECT id FROM library_item WHERE sha256=?`, digest).Scan(&itemID)
	if errors.Is(err, sql.ErrNoRows) {
		duplicate = false
		if itemID, err = randomID(); err != nil {
			return false, err
		}
		tags, err := sealWithAEAD(v.aead, mustJSON([]string{}), libraryTagsAAD(itemID))
		if err != nil {
			return false, err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO library_item(id,sha256,media_type,size_bytes,tags_ciphertext,created_at_utc) VALUES(?,?,?,?,?,?)`,
			itemID, digest, mediaType, info.Size(), tags, utcNow()); err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	}
	location, err := sealWithAEAD(v.aead, mustJSON(libraryLocation{RelativePath: relative, ModTimeNanos: info.ModTime().UnixNano()}), libraryLocationAAD(pathKey))
	if err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO library_location(path_key,item_id,source_id,ciphertext,size_bytes,mtime_ns,available) VALUES(?,?,?,?,?,?,1)
		ON CONFLICT(path_key) DO UPDATE SET item_id=excluded.item_id,ciphertext=excluded.ciphertext,size_bytes=excluded.size_bytes,mtime_ns=excluded.mtime_ns,available=1`,
		pathKey, itemID, sourceID, location, info.Size(), info.ModTime().UnixNano()); err != nil {
		return false, err
	}
	return duplicate, tx.Commit()
}

// TagLibraryItem replaces an item's owner-declared tags.
func (v *Vault) TagLibraryItem(ctx context.Context, consentID, itemID string, tags []string) error {
	if err := v.authorize(ctx, consentID, "owner", ScopeLibraryWrite); err != nil {
		return err
	}
	normalized, err := normalizeTags(tags)
	if err != nil {
		return err
	}
	sealed, err := sealWithAEAD(v.aead, mustJSON(normalized), libraryTagsAAD(itemID))
	if err != nil {
		return err
	}
	result, err := v.db.ExecContext(ctx, `UPDATE library_item SET tags_ciphertext=? WHERE id=?`, sealed, itemID)
	if err != nil {
		return err
	}
	if changed, _ := result.RowsAffected(); changed != 1 {
		return ErrLibraryNotFound
	}
	return nil
}

// RemoveLibraryItem drops an item and every location of it from the index, for
// content the owner does not want indexed. The files themselves are untouched.
func (v *Vault) RemoveLibraryItem(ctx context.Context, consentID, itemID string) error {
	if err := v.authorize(ctx, consentID, "owner", ScopeLibraryWrite); err != nil {
		return err
	}
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM library_location WHERE item_id=?`, itemID); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM library_item WHERE id=?`, itemID)
	if err != nil {
		return err
	}
	if changed, _ := result.RowsAffected(); changed != 1 {
		return ErrLibraryNotFound
	}
	return tx.Commit()
}

// SearchLibrary matches decrypted names and tags in memory; nothing about the
// query is written anywhere.
func (v *Vault) SearchLibrary(ctx context.Context, consentID string, query LibraryQuery) ([]LibraryItem, error) {
	if err := v.authorize(ctx, consentID, "owner", ScopeLibraryRead); err != nil {
		return nil, err
	}
	wantTags, err := normalizeTags(query.Tags)
	if err != nil {
		return nil, err
	}
	if query.MediaType != "" && query.MediaType != "image" && query.MediaType != "video" {
		return nil, ErrInvalidLibrary
	}
	limit := query.Limit
	if limit <= 0 || limit > maxLibraryResults {
		limit = maxLibraryResults
	}
	text := strings.ToLower(strings.TrimSpace(query.Text))
	rows, err := v.db.QueryContext(ctx, `SELECT i.id,i.sha256,i.media_type,i.size_bytes,i.tags_ciphertext,l.path_key,l.source_id,l.ciphertext,l.available,
		(SELECT count(*) FROM library_location c WHERE c.item_id=i.id)
		FROM library_item i JOIN library_location l ON l.path_key=(SELECT path_key FROM library_location WHERE item_id=i.id ORDER BY available DESC,path_key LIMIT 1)
		ORDER BY i.created_at_utc,i.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := []LibraryItem{}
	for rows.Next() {
		var item LibraryItem
		var tagsSealed, locationSealed []byte
		var pathKey string
		var available int
		if err = rows.Scan(&item.ID, &item.SHA256, &item.MediaType, &item.Bytes, &tagsSealed, &pathKey, &item.SourceID, &locationSealed, &available, &item.Copies); err != nil {
			return nil, err
		}
		if query.MediaType != "" && !strings.HasPrefix(item.MediaType, query.MediaType+"/") {
			continue
		}
		tagsPlain, err := openWithAEAD(v.aead, tagsSealed, libraryTagsAAD(item.ID))
		if err != nil || json.Unmarshal(tagsPlain, &item.Tags) != nil {
			return nil, fmt.Errorf("library tags cannot be opened")
		}
		locationPlain, err := openWithAEAD(v.aead, locationSealed, libraryLocationAAD(pathKey))
		var location libraryLocation
		if err != nil || json.Unmarshal(locationPlain, &location) != nil {
			return nil, fmt.Errorf("library location cannot be opened")
		}
		item.Name = filepath.Base(location.RelativePath)
		item.Available = available == 1
		if !hasAllTags(item.Tags, wantTags) {
			continue
		}
		if text != "" && !strings.Contains(strings.ToLower(location.RelativePath), text) && !containsTag(item.Tags, text) {
			continue
		}
		results = append(results, item)
		if len(results) == limit {
			break
		}
	}
	return results, rows.Err()
}

func (v *Vault) openSource(id string, sealed []byte) (string, string, error) {
	plain, err := openWithAEAD(v.aead, sealed, librarySourceAAD(id))
	if err != nil {
		return "", "", fmt.Errorf("library source cannot be opened")
	}
	var source map[string]string
	if err = json.Unmarshal(plain, &source); err != nil {
		return "", "", err
	}
	return source["path"], source["label"], nil
}

// libraryIndexKey is a random key, sealed by the vault key, for the keyed path
// hashes. Key rotation rewraps it, so the index survives rotation.
func (v *Vault) libraryIndexKey(ctx context.Context) ([]byte, error) {
	var wrapped []byte
	err := v.db.QueryRowContext(ctx, `SELECT wrapped_key FROM library_key WHERE singleton=1`).Scan(&wrapped)
	if errors.Is(err, sql.ErrNoRows) {
		key := make([]byte, 32)
		if _, err = rand.Read(key); err != nil {
			return nil, err
		}
		if wrapped, err = sealWithAEAD(v.aead, key, []byte("lunaos-library-index-key")); err != nil {
			return nil, err
		}
		if _, err = v.db.ExecContext(ctx, `INSERT INTO library_key(singleton,wrapped_key) VALUES(1,?)`, wrapped); err != nil {
			return nil, err
		}
		return key, nil
	}
	if err != nil {
		return nil, err
	}
	return openWithAEAD(v.aead, wrapped, []byte("lunaos-library-index-key"))
}

// sniffAndHash reads the file once: the first bytes decide whether it is an
// image or a video, and the whole stream is hashed.
func sniffAndHash(path string) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()
	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", "", err
	}
	head = head[:n]
	mediaType := detectMedia(head)
	if mediaType == "" {
		return "", "", nil
	}
	hasher := sha256.New()
	hasher.Write(head)
	if _, err = io.Copy(hasher, file); err != nil {
		return "", "", err
	}
	return mediaType, hex.EncodeToString(hasher.Sum(nil)), nil
}

func detectMedia(head []byte) string {
	switch detected := http.DetectContentType(head); {
	case strings.HasPrefix(detected, "image/") && detected != "image/svg+xml":
		return detected
	case strings.HasPrefix(detected, "video/"):
		return detected
	}
	switch {
	case len(head) >= 12 && bytes.Equal(head[4:8], []byte("ftyp")):
		brand := string(head[8:12])
		if strings.HasPrefix(brand, "heic") || strings.HasPrefix(brand, "heix") || strings.HasPrefix(brand, "mif1") {
			return "image/heic"
		}
		if brand == "qt  " {
			return "video/quicktime"
		}
		return "video/mp4"
	case len(head) >= 4 && bytes.Equal(head[:4], []byte{0x1a, 0x45, 0xdf, 0xa3}):
		return "video/x-matroska"
	}
	return ""
}

func normalizeTags(tags []string) ([]string, error) {
	if len(tags) > maxLibraryTags {
		return nil, ErrInvalidLibrary
	}
	seen := map[string]bool{}
	normalized := []string{}
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" || len(tag) > 64 {
			return nil, ErrInvalidLibrary
		}
		if !seen[tag] {
			seen[tag] = true
			normalized = append(normalized, tag)
		}
	}
	sort.Strings(normalized)
	return normalized, nil
}

func hasAllTags(have, want []string) bool {
	for _, tag := range want {
		if !containsTag(have, tag) {
			return false
		}
	}
	return true
}

func containsTag(tags []string, wanted string) bool {
	for _, tag := range tags {
		if tag == wanted {
			return true
		}
	}
	return false
}

func libraryPathKey(key []byte, sourceID, relative string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(sourceID + "\x00" + relative))
	return hex.EncodeToString(mac.Sum(nil))
}

func librarySourceAAD(id string) []byte        { return []byte("lunaos-library-source:" + id) }
func libraryTagsAAD(id string) []byte          { return []byte("lunaos-library-tags:" + id) }
func libraryLocationAAD(pathKey string) []byte { return []byte("lunaos-library-location:" + pathKey) }

func mustJSON(value any) []byte {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return encoded
}
