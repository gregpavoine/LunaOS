package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

// The Civitai connector follows civitai-developer-docs site/reference/images.md
// (GET /api/v1/images) and site/guide/authentication.md (bearer token).
const (
	// civitaiLicense marks rights granted by Civitai's terms: image items
	// declare no licence of their own. The owner must list it explicitly.
	civitaiLicense = "provider-terms:civitai"
	// baseModelTag prefixes the documented item field `baseModel` in Tags.
	baseModelTag = "base_model:"

	// The documented 400 messages for a cursor the API will not continue.
	cursorExpiredMsg = "This cursor can no longer be continued"
	cursorFiltersMsg = "This cursor cannot be continued with these filters"

	// civitaiMaxLevel is the highest documented `nsfw` value.
	civitaiMaxLevel = "X"

	civitaiMaxBody          = 8 << 20
	civitaiDefaultRateLimit = 60 * time.Second

	scopeTag  = "tag:"
	scopeUser = "user:"
)

var (
	errCivitaiScope    = errors.New("DISCOVERY_CIVITAI_SCOPE")
	errCivitaiResponse = errors.New("DISCOVERY_CIVITAI_BAD_RESPONSE")
	errCivitaiTooLarge = errors.New("DISCOVERY_CIVITAI_RESPONSE_TOO_LARGE")
	errCivitaiRequest  = errors.New("DISCOVERY_CIVITAI_REQUEST")
	// errCivitaiBadCursor is a 400 refusing the cursor; Fetch turns it into
	// ErrCursorExpired when a cursor was sent.
	errCivitaiBadCursor  = errors.New("DISCOVERY_CIVITAI_BAD_CURSOR")
	errCivitaiBadRequest = errors.New("DISCOVERY_CIVITAI_STATUS_400")
)

type civitai struct {
	base   string
	apiKey string
	policy CivitaiPolicy
	hc     *http.Client
	now    func() time.Time
}

// NewCivitai returns the allowlisted Civitai connector. Every listing request
// carries exactly one allowlist filter (tag ID or uploader), the owner's base
// models when listed (they only narrow, never form a scope) and the
// content-level cap; an empty allowlist yields no scopes and no request.
func NewCivitai(baseURL, apiKey string, p CivitaiPolicy, hc *http.Client) Connector {
	if hc == nil {
		hc = http.DefaultClient
	}
	return &civitai{base: strings.TrimRight(baseURL, "/"), apiKey: apiKey, policy: p, hc: hc, now: time.Now}
}

func (c *civitai) Name() string { return "civitai" }

func (c *civitai) Scopes() []string {
	var scopes []string
	for _, t := range c.policy.Tags {
		scopes = append(scopes, scopeTag+t)
	}
	for _, u := range c.policy.Creators {
		scopes = append(scopes, scopeUser+u)
	}
	return scopes
}

type civitaiImage struct {
	ID        json.Number `json:"id"`
	URL       string      `json:"url"`
	NSFWLevel string      `json:"nsfwLevel"`
	Username  string      `json:"username"`
	BaseModel string      `json:"baseModel"`
}

type civitaiResponse struct {
	Items    []civitaiImage `json:"items"`
	Metadata struct {
		NextCursor string `json:"nextCursor"`
	} `json:"metadata"`
}

// Fetch issues one GET /api/v1/images for an allowlisted scope.
func (c *civitai) Fetch(ctx context.Context, scope, cursor string) (Page, error) {
	if scope == "" || !slices.Contains(c.Scopes(), scope) {
		return Page{}, errCivitaiScope
	}
	q := url.Values{}
	var tag string
	switch {
	case strings.HasPrefix(scope, scopeTag):
		tag = strings.TrimPrefix(scope, scopeTag)
		q.Set("tags", tag)
	case strings.HasPrefix(scope, scopeUser):
		q.Set("username", strings.TrimPrefix(scope, scopeUser))
	default:
		return Page{}, errCivitaiScope
	}
	if len(c.policy.ModelFamilies) > 0 {
		q.Set("baseModels", strings.Join(c.policy.ModelFamilies, ","))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	resp, err := c.get(ctx, q)
	if errors.Is(err, errCivitaiBadCursor) {
		if cursor != "" {
			return Page{}, ErrCursorExpired
		}
		return Page{}, errCivitaiBadRequest
	}
	if err != nil {
		return Page{}, err
	}
	page := Page{NextCursor: resp.Metadata.NextCursor}
	for _, img := range resp.Items {
		cand := c.candidate(img)
		if tag != "" {
			cand.Tags = append(cand.Tags, tag)
		}
		// The item's own base model, so a result outside ModelFamilies is rejected.
		if img.BaseModel != "" {
			cand.Tags = append(cand.Tags, baseModelTag+img.BaseModel)
		}
		page.Items = append(page.Items, cand)
	}
	return page, nil
}

func (c *civitai) candidate(img civitaiImage) Candidate {
	return Candidate{
		Item: Item{
			Provider:     "civitai",
			ExternalID:   img.ID.String(),
			Creator:      img.Username,
			License:      civitaiLicense,
			ContentLevel: img.NSFWLevel,
		},
		MediaURL: img.URL,
	}
}

// Removed looks each stored ID up with the documented single-image filter
// `imageId` at the documented maximum level `nsfw=X`, so the owner's cap never
// drives deletions; an ID the provider no longer returns is reported removed.
// The engine's takedown guard bounds the damage of a wrong absence.
func (c *civitai) Removed(ctx context.Context, externalIDs []string) ([]string, error) {
	var gone []string
	for _, id := range externalIDs {
		if _, err := strconv.ParseUint(id, 10, 64); err != nil {
			continue
		}
		resp, err := c.get(ctx, url.Values{"imageId": {id}, "nsfw": {civitaiMaxLevel}})
		if err != nil {
			return nil, err
		}
		if !slices.ContainsFunc(resp.Items, func(img civitaiImage) bool { return img.ID.String() == id }) {
			gone = append(gone, id)
		}
	}
	return gone, nil
}

// get sends one authenticated request, with the owner's content-level cap
// unless the caller set `nsfw`. Errors are codes only: no URL, query, header
// or body is ever included.
func (c *civitai) get(ctx context.Context, q url.Values) (civitaiResponse, error) {
	if !q.Has("nsfw") {
		q.Set("nsfw", c.policy.MaxContentLevel)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/api/v1/images?"+q.Encode(), nil)
	if err != nil {
		return civitaiResponse{}, errCivitaiRequest
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	res, err := c.hc.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return civitaiResponse{}, ctx.Err()
		}
		return civitaiResponse{}, ErrTransient
	}
	defer res.Body.Close()
	switch {
	case res.StatusCode == http.StatusTooManyRequests:
		return civitaiResponse{}, ErrRateLimited{RetryAfter: c.retryAfter(res.Header.Get("Retry-After"))}
	case res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden:
		return civitaiResponse{}, ErrUnauthorized
	case res.StatusCode >= 500:
		return civitaiResponse{}, ErrTransient
	case res.StatusCode == http.StatusBadRequest:
		return civitaiResponse{}, badRequest(res.Body)
	case res.StatusCode != http.StatusOK:
		return civitaiResponse{}, fmt.Errorf("DISCOVERY_CIVITAI_STATUS_%d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, civitaiMaxBody+1))
	if err != nil {
		if ctx.Err() != nil {
			return civitaiResponse{}, ctx.Err()
		}
		return civitaiResponse{}, ErrTransient
	}
	if len(body) > civitaiMaxBody {
		return civitaiResponse{}, errCivitaiTooLarge
	}
	var out civitaiResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return civitaiResponse{}, errCivitaiResponse
	}
	for _, img := range out.Items {
		if _, err := strconv.ParseUint(img.ID.String(), 10, 64); err != nil {
			return civitaiResponse{}, errCivitaiResponse
		}
	}
	return out, nil
}

// badRequest classifies a 400 by its documented `message`; the body is never
// surfaced.
func badRequest(body io.Reader) error {
	var msg struct {
		Message string `json:"message"`
	}
	data, _ := io.ReadAll(io.LimitReader(body, 64<<10))
	if json.Unmarshal(data, &msg) == nil &&
		(strings.HasPrefix(msg.Message, cursorExpiredMsg) || strings.HasPrefix(msg.Message, cursorFiltersMsg)) {
		return errCivitaiBadCursor
	}
	return errCivitaiBadRequest
}

// retryAfter reads Retry-After as seconds or an HTTP date, defaulting to 60 s.
func (c *civitai) retryAfter(v string) time.Duration {
	if s, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && s >= 0 {
		return time.Duration(s) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := t.Sub(c.now()); d > 0 {
			return d
		}
		return 0
	}
	return civitaiDefaultRateLimit
}
