package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The Reddit connector follows the reddit-archive OAuth2 wiki (refresh-token
// flow, HTTP Basic client auth, "Authorization: bearer" against
// oauth.reddit.com), its API rules page (OAuth required, descriptive
// User-Agent, X-Ratelimit-Used/Remaining/Reset) and www.reddit.com/dev/api
// (GET [/r/subreddit]/new with after/limit ≤ 100, GET /api/info with a
// comma-separated list of fullnames). Using it requires API access approved
// by Reddit under its Responsible Builder Policy.
//
// Reddit posts declare no licence, so the connector feeds the index with link
// metadata only: MediaURL is always empty and nothing is ever downloaded.
const (
	redditDefaultTokenURL = "https://www.reddit.com/api/v1/access_token"
	redditDefaultAPIBase  = "https://oauth.reddit.com"
	redditPermalinkBase   = "https://www.reddit.com"

	redditPageSize  = 100
	redditInfoBatch = 100
	redditMaxBody   = 8 << 20
	redditMaxToken  = 64 << 10
	// All bearer tokens expire after one hour; refresh a minute early.
	redditTokenLife   = time.Hour
	redditTokenMargin = time.Minute
	redditDefaultWait = 60 * time.Second

	scopeSubreddit = "r/"
)

var (
	subredditName = regexp.MustCompile(`^[A-Za-z0-9_]{2,21}$`)
	// A link's fullname: the t3_ kind prefix and a base-36 ID.
	redditLinkName = regexp.MustCompile(`^t3_[a-z0-9]{1,13}$`)

	errRedditScope    = errors.New("DISCOVERY_REDDIT_SCOPE")
	errRedditResponse = errors.New("DISCOVERY_REDDIT_BAD_RESPONSE")
	errRedditTooLarge = errors.New("DISCOVERY_REDDIT_RESPONSE_TOO_LARGE")
	errRedditRequest  = errors.New("DISCOVERY_REDDIT_REQUEST")
	// errRedditRetryAuth is a 401 on an API request: refresh once and retry.
	errRedditRetryAuth = errors.New("DISCOVERY_REDDIT_TOKEN_REFUSED")
)

// RedditConfig holds the owner's OAuth client and refresh token. ClientSecret
// may be empty for an installed app; the others are required.
type RedditConfig struct {
	ClientID, ClientSecret, RefreshToken, UserAgent string
	TokenURL, APIBase                               string
}

type reddit struct {
	cfg  RedditConfig
	subs []string
	hc   *http.Client
	now  func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
	// blocked holds the end of an exhausted X-Ratelimit window.
	blocked time.Time
}

// NewReddit returns the subreddit-allowlisted Reddit connector. Only listed
// subreddits with valid names become scopes; no other subreddit is ever
// requested or suggested.
func NewReddit(cfg RedditConfig, p RedditPolicy, hc *http.Client) Connector {
	if hc == nil {
		hc = http.DefaultClient
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = redditDefaultTokenURL
	}
	if cfg.APIBase == "" {
		cfg.APIBase = redditDefaultAPIBase
	}
	cfg.APIBase = strings.TrimRight(cfg.APIBase, "/")
	var subs []string
	for _, s := range p.Subreddits {
		if subredditName.MatchString(s) && !containsFold(subs, s) {
			subs = append(subs, s)
		}
	}
	return &reddit{cfg: cfg, subs: subs, hc: hc, now: time.Now}
}

func (r *reddit) Name() string { return "reddit" }

func (r *reddit) Scopes() []string {
	var scopes []string
	for _, s := range r.subs {
		scopes = append(scopes, scopeSubreddit+s)
	}
	return scopes
}

func (r *reddit) configured() bool {
	return strings.TrimSpace(r.cfg.UserAgent) != "" && r.cfg.ClientID != "" && r.cfg.RefreshToken != ""
}

type redditPost struct {
	Name              string  `json:"name"`
	Subreddit         string  `json:"subreddit"`
	Title             string  `json:"title"`
	Author            string  `json:"author"`
	Over18            bool    `json:"over_18"`
	PostHint          string  `json:"post_hint"`
	URL               string  `json:"url"`
	Permalink         string  `json:"permalink"`
	RemovedByCategory *string `json:"removed_by_category"`
}

type redditListing struct {
	Kind string `json:"kind"`
	Data struct {
		After    *string `json:"after"`
		Children []struct {
			Kind string     `json:"kind"`
			Data redditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// posts returns the listing's well-formed links.
func (l redditListing) posts() []redditPost {
	var out []redditPost
	for _, c := range l.Data.Children {
		if c.Kind == "t3" && redditLinkName.MatchString(c.Data.Name) {
			out = append(out, c.Data)
		}
	}
	return out
}

func (p redditPost) gone() bool {
	return p.RemovedByCategory != nil || p.Author == "[deleted]"
}

// directMedia reports a post whose link is the media itself, as Reddit
// classifies it or as hosted on Reddit's own media domains.
func (p redditPost) directMedia() bool {
	if p.PostHint == "image" || p.PostHint == "hosted:video" {
		return true
	}
	u, err := url.Parse(p.URL)
	if err != nil || u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return host == "i.redd.it" || host == "v.redd.it"
}

// Fetch issues one GET /r/<sub>/new for a listed subreddit.
func (r *reddit) Fetch(ctx context.Context, scope, cursor string) (Page, error) {
	if !r.configured() {
		return Page{}, ErrConfig
	}
	if scope == "" || !slices.Contains(r.Scopes(), scope) {
		return Page{}, errRedditScope
	}
	sub := strings.TrimPrefix(scope, scopeSubreddit)
	q := url.Values{"limit": {strconv.Itoa(redditPageSize)}, "raw_json": {"1"}}
	if cursor != "" {
		q.Set("after", cursor)
	}
	l, err := r.get(ctx, "/r/"+sub+"/new", q, true)
	if err != nil {
		return Page{}, err
	}
	var page Page
	if l.Data.After != nil {
		page.NextCursor = *l.Data.After
	}
	for _, p := range l.posts() {
		switch {
		case p.gone():
			page.Removed = append(page.Removed, p.Name)
		case p.directMedia():
			page.Items = append(page.Items, p.candidate())
		}
	}
	return page, nil
}

func (p redditPost) candidate() Candidate {
	level := "None"
	if p.Over18 {
		level = "X"
	}
	source := ""
	if strings.HasPrefix(p.Permalink, "/r/") {
		source = redditPermalinkBase + p.Permalink
	}
	return Candidate{Item: Item{
		Provider:     "reddit",
		ExternalID:   p.Name,
		SourceURL:    source,
		LinkURL:      p.URL,
		Title:        p.Title,
		Creator:      p.Author,
		Tags:         []string{scopeSubreddit + p.Subreddit},
		ContentLevel: level,
	}}
}

// Removed asks GET /api/info for the stored links in batches of 100. A link
// absent from the answer, removed or deleted is reported removed; the
// engine's takedown guard bounds the damage of a wrong absence.
func (r *reddit) Removed(ctx context.Context, externalIDs []string) ([]string, error) {
	if !r.configured() {
		return nil, ErrConfig
	}
	var ids []string
	for _, id := range externalIDs {
		if redditLinkName.MatchString(id) {
			ids = append(ids, id)
		}
	}
	var gone []string
	for batch := range slices.Chunk(ids, redditInfoBatch) {
		q := url.Values{"id": {strings.Join(batch, ",")}, "raw_json": {"1"}}
		l, err := r.get(ctx, "/api/info", q, false)
		if err != nil {
			return nil, err
		}
		live := map[string]bool{}
		for _, p := range l.posts() {
			if !p.gone() {
				live[p.Name] = true
			}
		}
		for _, id := range batch {
			if !live[id] {
				gone = append(gone, id)
			}
		}
	}
	return gone, nil
}

// get sends one authenticated API request. A 401 refreshes the token once and
// retries once. A listing's 403/404 is ErrScopeUnavailable. Errors are codes
// only: no URL, subreddit, token or body is ever included.
func (r *reddit) get(ctx context.Context, path string, q url.Values, listing bool) (redditListing, error) {
	if err := r.budget(); err != nil {
		return redditListing{}, err
	}
	l, err := r.getOnce(ctx, path, q, listing)
	if !errors.Is(err, errRedditRetryAuth) {
		return l, err
	}
	r.dropToken()
	if err = r.budget(); err != nil {
		return redditListing{}, err
	}
	l, err = r.getOnce(ctx, path, q, listing)
	if errors.Is(err, errRedditRetryAuth) {
		// The fresh token was refused too: never reuse it.
		r.dropToken()
		return redditListing{}, ErrUnauthorized
	}
	return l, err
}

func (r *reddit) getOnce(ctx context.Context, path string, q url.Values, listing bool) (redditListing, error) {
	tok, err := r.accessToken(ctx)
	if err != nil {
		return redditListing{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.cfg.APIBase+path+"?"+q.Encode(), nil)
	if err != nil {
		return redditListing{}, errRedditRequest
	}
	req.Header.Set("Authorization", "bearer "+tok)
	req.Header.Set("User-Agent", r.cfg.UserAgent)
	req.Header.Set("Accept", "application/json")
	res, err := r.do(ctx, req)
	if err != nil {
		return redditListing{}, err
	}
	defer res.Body.Close()
	switch {
	case res.StatusCode == http.StatusUnauthorized:
		drain(res.Body)
		return redditListing{}, errRedditRetryAuth
	case listing && (res.StatusCode == http.StatusForbidden || res.StatusCode == http.StatusNotFound):
		drain(res.Body)
		return redditListing{}, ErrScopeUnavailable
	}
	if err = r.status(res); err != nil {
		return redditListing{}, err
	}
	var l redditListing
	if err = readJSON(ctx, res.Body, redditMaxBody, &l); err != nil {
		return redditListing{}, err
	}
	if l.Kind != "Listing" {
		return redditListing{}, errRedditResponse
	}
	return l, nil
}

// accessToken returns the cached bearer token, refreshing it near expiry with
// the documented refresh_token grant. A refused grant is ErrUnauthorized.
func (r *reddit) accessToken(ctx context.Context) (string, error) {
	r.mu.Lock()
	tok, expires := r.token, r.expires
	r.mu.Unlock()
	if tok != "" && r.now().Before(expires) {
		return tok, nil
	}
	form := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {r.cfg.RefreshToken}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", errRedditRequest
	}
	req.SetBasicAuth(r.cfg.ClientID, r.cfg.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", r.cfg.UserAgent)
	req.Header.Set("Accept", "application/json")
	res, err := r.do(ctx, req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		drain(res.Body)
		return "", ErrUnauthorized
	}
	if err = r.status(res); err != nil {
		return "", err
	}
	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
		Error       string      `json:"error"`
	}
	if err = readJSON(ctx, res.Body, redditMaxToken, &out); err != nil {
		return "", err
	}
	if out.Error != "" {
		return "", ErrUnauthorized
	}
	if out.AccessToken == "" || !strings.EqualFold(out.TokenType, "bearer") {
		return "", errRedditResponse
	}
	// expires_in is seconds; anything missing or beyond the documented hour
	// is clamped to it.
	life := redditTokenLife
	if s, err := out.ExpiresIn.Int64(); err == nil && s > 0 && time.Duration(s)*time.Second < life {
		life = time.Duration(s) * time.Second
	}
	r.mu.Lock()
	r.token, r.expires = out.AccessToken, r.now().Add(life-redditTokenMargin)
	r.mu.Unlock()
	return out.AccessToken, nil
}

func (r *reddit) dropToken() {
	r.mu.Lock()
	r.token, r.expires = "", time.Time{}
	r.mu.Unlock()
}

// budget refuses to send a request while an exhausted rate-limit window runs.
func (r *reddit) budget() error {
	r.mu.Lock()
	until := r.blocked
	r.mu.Unlock()
	if wait := until.Sub(r.now()); wait > 0 {
		return ErrRateLimited{RetryAfter: wait}
	}
	return nil
}

// do sends a request and records an exhausted X-Ratelimit budget: the
// response is still used, and the next request waits for X-Ratelimit-Reset.
func (r *reddit) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	res, err := r.hc.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, ErrTransient
	}
	if v := res.Header.Get("X-Ratelimit-Remaining"); v != "" {
		if left, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil && left < 1 {
			r.mu.Lock()
			r.blocked = r.now().Add(r.resetIn(res.Header))
			r.mu.Unlock()
		}
	}
	return res, nil
}

// status maps a non-200 status other than the auth and scope cases.
func (r *reddit) status(res *http.Response) error {
	if res.StatusCode == http.StatusOK {
		return nil
	}
	drain(res.Body)
	switch {
	case res.StatusCode == http.StatusTooManyRequests:
		return ErrRateLimited{RetryAfter: r.retryAfter(res.Header)}
	case res.StatusCode >= 500:
		return ErrTransient
	}
	return fmt.Errorf("DISCOVERY_REDDIT_STATUS_%d", res.StatusCode)
}

// retryAfter reads Retry-After (seconds or an HTTP date), then
// X-Ratelimit-Reset, defaulting to 60 s.
func (r *reddit) retryAfter(h http.Header) time.Duration {
	v := strings.TrimSpace(h.Get("Retry-After"))
	if s, err := strconv.Atoi(v); err == nil && s >= 0 {
		return time.Duration(s) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		return max(t.Sub(r.now()), 0)
	}
	return r.resetIn(h)
}

// resetIn reads X-Ratelimit-Reset (seconds to the end of the period),
// defaulting to 60 s.
func (r *reddit) resetIn(h http.Header) time.Duration {
	if s, err := strconv.ParseFloat(strings.TrimSpace(h.Get("X-Ratelimit-Reset")), 64); err == nil && s >= 0 && s < 24*3600 {
		return time.Duration(s * float64(time.Second))
	}
	return redditDefaultWait
}

func readJSON(ctx context.Context, body io.Reader, limit int64, v any) error {
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return ErrTransient
	}
	if int64(len(data)) > limit {
		return errRedditTooLarge
	}
	if json.Unmarshal(data, v) != nil {
		return errRedditResponse
	}
	return nil
}

// drain reads a bounded remainder so the connection can be reused.
func drain(body io.Reader) { _, _ = io.Copy(io.Discard, io.LimitReader(body, 64<<10)) }
