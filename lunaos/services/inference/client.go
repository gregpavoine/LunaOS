package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// clientTimeout is a ceiling above the server's start wait plus generation.
const clientTimeout = 10 * time.Minute

// Client calls luna-inference over its Unix socket.
type Client struct {
	http *http.Client
}

// Result is one generation. Text is model output: untrusted data that can
// never authorize an action, lift BLIND or widen a scope. Untrusted is what
// the server sent; Generate fails rather than return a Result without it.
type Result struct {
	Text      string
	Model     string
	Untrusted bool
}

// APIError is a non-200 reply. Code is the server's JSON error code, or ""
// when the reply carried none.
type APIError struct {
	Status int
	Code   string
}

func (e *APIError) Error() string {
	if e.Code == "" {
		return fmt.Sprintf("luna-inference: HTTP %d", e.Status)
	}
	return fmt.Sprintf("luna-inference: HTTP %d %s", e.Status, e.Code)
}

// NewClient returns a client for the socket path. It never uses a proxy.
func NewClient(socket string) *Client {
	var d net.Dialer
	return &Client{http: &http.Client{
		Timeout: clientTimeout,
		Transport: &http.Transport{
			Proxy: nil,
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return d.DialContext(ctx, "unix", socket)
			},
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// Generate asks the local model for at most maxTokens tokens. The server
// validates prompt and maxTokens; its refusals come back as *APIError.
func (c *Client) Generate(ctx context.Context, prompt string, maxTokens int) (Result, error) {
	var body bytes.Buffer
	enc := json.NewEncoder(&body)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(generateRequest{Prompt: prompt, MaxTokens: maxTokens}); err != nil {
		return Result{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://luna-inference/v1/generate", &body)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxUpstreamBytes))
	if err != nil {
		return Result{}, err
	}
	if resp.StatusCode != http.StatusOK {
		var e errorReply
		_ = json.Unmarshal(raw, &e)
		return Result{}, &APIError{Status: resp.StatusCode, Code: e.Code}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var out struct {
		Text      string `json:"text"`
		Model     string `json:"model"`
		Untrusted *bool  `json:"untrusted"`
	}
	if err := dec.Decode(&out); err != nil {
		return Result{}, fmt.Errorf("luna-inference reply: %w", err)
	}
	if dec.Decode(new(any)) != io.EOF {
		return Result{}, errors.New("luna-inference reply: trailing data")
	}
	// A reply not flagged untrusted is refused, so it can never pass as trusted.
	if out.Untrusted == nil || !*out.Untrusted {
		return Result{}, errors.New("luna-inference reply is not marked untrusted")
	}
	return Result{Text: out.Text, Model: out.Model, Untrusted: true}, nil
}
