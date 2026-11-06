package discovery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DefaultMaxBytes caps one downloaded media file when the caller sets no limit.
const DefaultMaxBytes int64 = 512 << 20

const (
	stalePartAge           = 24 * time.Hour
	downloadDefaultBackoff = 60 * time.Second
)

var (
	// ErrTooLarge means the media exceeds the size cap; it is never retried.
	ErrTooLarge = errors.New("DISCOVERY_TOO_LARGE")
	// ErrDownloadGone means the media URL is gone (404/410); never retried.
	ErrDownloadGone = errors.New("DISCOVERY_DOWNLOAD_GONE")
	// ErrDownloadRefused means the media could not or may not be fetched
	// (insecure URL or redirect, refused status); never retried.
	ErrDownloadRefused = errors.New("DISCOVERY_DOWNLOAD_REFUSED")
)

// NewDownloader returns an Engine.Download backed by Download.
func NewDownloader(hc *http.Client, maxBytes int64, tmpDir string) func(context.Context, Candidate) (string, string, error) {
	return func(ctx context.Context, c Candidate) (string, string, error) {
		return Download(ctx, hc, c, maxBytes, tmpDir)
	}
}

// Download fetches c.MediaURL into tmpDir (a private directory) and returns the
// final file's path and SHA-256. The partial file is named after the candidate's
// identity hash, never its URL or title, and survives a failed attempt so the
// next one resumes it with a Range request. Errors carry codes only.
func Download(ctx context.Context, hc *http.Client, c Candidate, maxBytes int64, tmpDir string) (string, string, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	u, err := url.Parse(c.MediaURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", "", ErrDownloadRefused
	}
	if err := privateDir(tmpDir); err != nil {
		return "", "", ErrDownloadRefused
	}
	sweepStaleParts(tmpDir)
	if hc == nil {
		hc = http.DefaultClient
	}
	client := *hc
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if req.URL.Scheme != "https" || len(via) >= 10 {
			return ErrDownloadRefused
		}
		return nil
	}

	id := sha256.Sum256([]byte(c.Provider + "\x00" + c.ExternalID))
	base := filepath.Join(tmpDir, hex.EncodeToString(id[:]))
	part, final := base+".part", base+".media"

	// A second pass runs only after a resume the server could not honour.
	for pass := 0; pass < 2; pass++ {
		restart, err := fetchPart(ctx, &client, u.String(), part, maxBytes, pass == 1)
		if err != nil {
			return "", "", err
		}
		if restart {
			continue
		}
		sha, err := hashFile(part)
		if err != nil {
			return "", "", ErrTransient
		}
		if c.DeclaredSHA256 != "" && !strings.EqualFold(c.DeclaredSHA256, sha) {
			_ = os.Remove(part)
			return "", "", ErrHashMismatch
		}
		if err := ctx.Err(); err != nil {
			return "", "", err
		}
		if err := os.Rename(part, final); err != nil {
			return "", "", ErrTransient
		}
		return final, sha, nil
	}
	return "", "", ErrTransient
}

// fetchPart appends the media to part. restart reports that the part was
// discarded and the request must be repeated from byte zero.
func fetchPart(ctx context.Context, hc *http.Client, rawURL, part string, maxBytes int64, fromZero bool) (restart bool, err error) {
	flags := os.O_RDWR | os.O_CREATE
	if fromZero {
		flags |= os.O_TRUNC
	}
	f, err := os.OpenFile(part, flags, 0600)
	if err != nil {
		return false, ErrTransient
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return false, ErrTransient
	}
	have := info.Size()
	if have > maxBytes {
		have = 0
		if err = f.Truncate(0); err != nil {
			return false, ErrTransient
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return false, ErrDownloadRefused
	}
	if have > 0 {
		req.Header.Set("Range", "bytes="+strconv.FormatInt(have, 10)+"-")
	}
	res, err := hc.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		if errors.Is(err, ErrDownloadRefused) {
			return false, ErrDownloadRefused
		}
		return false, ErrTransient
	}
	defer res.Body.Close()

	switch {
	case res.StatusCode == http.StatusOK:
		if have > 0 {
			if err = f.Truncate(0); err != nil {
				return false, ErrTransient
			}
			have = 0
		}
		if res.ContentLength > maxBytes {
			return false, tooLarge(f, part)
		}
	case res.StatusCode == http.StatusPartialContent && have > 0:
		start, total, ok := parseContentRange(res.Header.Get("Content-Range"))
		if !ok || start != have {
			return true, resetPart(f)
		}
		if total > maxBytes || (res.ContentLength >= 0 && have+res.ContentLength > maxBytes) {
			return false, tooLarge(f, part)
		}
	case res.StatusCode == http.StatusPartialContent || res.StatusCode == http.StatusRequestedRangeNotSatisfiable:
		if have > 0 {
			return true, resetPart(f)
		}
		return false, ErrTransient
	case res.StatusCode == http.StatusTooManyRequests:
		return false, ErrRateLimited{RetryAfter: downloadRetryAfter(res.Header.Get("Retry-After"))}
	case res.StatusCode >= 500:
		return false, ErrTransient
	case res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusGone:
		return false, ErrDownloadGone
	default:
		return false, ErrDownloadRefused
	}

	if _, err = f.Seek(have, io.SeekStart); err != nil {
		return false, ErrTransient
	}
	n, err := io.Copy(f, io.LimitReader(res.Body, maxBytes-have+1))
	if have+n > maxBytes {
		return false, tooLarge(f, part)
	}
	if err != nil {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		return false, ErrTransient
	}
	// A body shorter than declared surfaces as an error from the reader; a
	// context cancelled right at the end must not yield a final file.
	return false, ctx.Err()
}

func resetPart(f *os.File) error {
	if err := f.Truncate(0); err != nil {
		return ErrTransient
	}
	return nil
}

func tooLarge(f *os.File, part string) error {
	_ = f.Close()
	_ = os.Remove(part)
	return ErrTooLarge
}

// parseContentRange reads "bytes start-end/total"; total is -1 when unknown.
func parseContentRange(v string) (start, total int64, ok bool) {
	rest, found := strings.CutPrefix(strings.TrimSpace(v), "bytes ")
	if !found {
		return 0, 0, false
	}
	span, size, found := strings.Cut(rest, "/")
	if !found {
		return 0, 0, false
	}
	from, _, found := strings.Cut(span, "-")
	if !found {
		return 0, 0, false
	}
	start, err := strconv.ParseInt(from, 10, 64)
	if err != nil || start < 0 {
		return 0, 0, false
	}
	total = -1
	if size != "*" {
		if total, err = strconv.ParseInt(size, 10, 64); err != nil {
			return 0, 0, false
		}
	}
	return start, total, true
}

func downloadRetryAfter(v string) time.Duration {
	if s, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && s >= 0 {
		return time.Duration(s) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		return max(time.Until(t), 0)
	}
	return downloadDefaultBackoff
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// sweepStaleParts removes abandoned partial downloads older than a day.
func sweepStaleParts(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".part") {
			continue
		}
		if info, err := e.Info(); err == nil && time.Since(info.ModTime()) > stalePartAge {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}
