package inference

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

// Model is one manifest entry. The file is stored as <id>.gguf.
type Model struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	SHA256    string `json:"sha256"`
	License   string `json:"license"`
	SizeBytes int64  `json:"size_bytes"`
	VRAMMiB   int    `json:"vram_mib"`
}

var (
	// IDs become file names: lowercase alphanumerics, '.', '_', '-', and no leading dot.
	idRE     = regexp.MustCompile(`^[a-z0-9_-][a-z0-9._-]*$`)
	sha256RE = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

func (m Model) validate() error {
	if !idRE.MatchString(m.ID) {
		return fmt.Errorf("model id %q is not a safe file name", m.ID)
	}
	if !sha256RE.MatchString(m.SHA256) {
		return fmt.Errorf("model %s: sha256 must be 64 lowercase hex characters", m.ID)
	}
	u, err := url.Parse(m.URL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("model %s: url must be an https URL", m.ID)
	}
	if m.License == "" {
		return fmt.Errorf("model %s: license is empty", m.ID)
	}
	if m.SizeBytes <= 0 {
		return fmt.Errorf("model %s: size_bytes must be positive", m.ID)
	}
	if m.VRAMMiB < 0 {
		return fmt.Errorf("model %s: vram_mib is negative", m.ID)
	}
	return nil
}

// LoadManifest reads and validates a JSON array of models.
func LoadManifest(path string) ([]Model, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var models []Model
	if err := dec.Decode(&models); err != nil {
		return nil, fmt.Errorf("manifest %s: %w", path, err)
	}
	seen := map[string]bool{}
	for _, m := range models {
		if err := m.validate(); err != nil {
			return nil, err
		}
		if seen[m.ID] {
			return nil, fmt.Errorf("duplicate model id %q", m.ID)
		}
		seen[m.ID] = true
	}
	return models, nil
}

// fileHash returns the sha256 of a file as lowercase hex.
func fileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Fetch returns the path of the verified model file in dir, downloading it
// when absent. An existing file with a wrong hash is deleted first. The
// download goes to dir/.<id>.part, is capped at SizeBytes, hash-checked, then
// renamed to dir/<id>.gguf. The partial file is removed on every failure.
func Fetch(ctx context.Context, m Model, dir string, client *http.Client) (string, error) {
	// Fetch is the last line of defense: check what it relies on itself.
	if !idRE.MatchString(m.ID) || !sha256RE.MatchString(m.SHA256) || m.SizeBytes <= 0 {
		return "", fmt.Errorf("model %q: invalid id, sha256 or size_bytes", m.ID)
	}
	final := filepath.Join(dir, m.ID+".gguf")
	part := filepath.Join(dir, "."+m.ID+".part")

	if got, err := fileHash(final); err == nil {
		if got == m.SHA256 {
			return final, nil
		}
		if err := os.Remove(final); err != nil {
			return "", err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	if err := download(ctx, m, part, client); err != nil {
		os.Remove(part)
		return "", err
	}
	if err := os.Rename(part, final); err != nil {
		os.Remove(part)
		return "", err
	}
	return final, nil
}

func download(ctx context.Context, m Model, part string, client *http.Client) error {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.URL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("model %s: download returned HTTP %d", m.ID, resp.StatusCode)
	}
	if resp.ContentLength > m.SizeBytes {
		return fmt.Errorf("model %s: server announces %d bytes, limit is %d", m.ID, resp.ContentLength, m.SizeBytes)
	}

	f, err := os.OpenFile(part, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	h := sha256.New()
	// Read one byte past the limit to detect oversize bodies.
	n, err := io.Copy(io.MultiWriter(f, h), io.LimitReader(resp.Body, m.SizeBytes+1))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	if n > m.SizeBytes {
		return fmt.Errorf("model %s: download exceeds %d bytes", m.ID, m.SizeBytes)
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != m.SHA256 {
		return fmt.Errorf("model %s: sha256 mismatch (got %s)", m.ID, got)
	}
	return nil
}
