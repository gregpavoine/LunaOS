package discovery

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

// maxPolicyBytes bounds the owner's policy file read.
const maxPolicyBytes = 1 << 20

// Policy is the owner's allowlist and rights policy, loaded from a 0600 JSON
// file. Admit fails closed on anything it cannot vouch for.
type Policy struct {
	Civitai CivitaiPolicy `json:"civitai"`
	Reddit  RedditPolicy  `json:"reddit"`
}

// CivitaiPolicy lists the Civitai tag IDs and uploader usernames to query, the
// base models that narrow them, the highest content level to accept and the
// licence tokens the owner accepts. Tags are numeric tag IDs, as the images endpoint's `tags`
// parameter requires.
type CivitaiPolicy struct {
	Tags            []string `json:"tags"`
	Creators        []string `json:"creators"`
	ModelFamilies   []string `json:"model_families"`
	MaxContentLevel string   `json:"max_content_level"`
	AllowedLicenses []string `json:"allowed_licenses"`
}

type RedditPolicy struct {
	Subreddits []string `json:"subreddits"`
}

// contentLevels orders Civitai's documented nsfwLevel strings.
var contentLevels = map[string]int{"None": 0, "Soft": 1, "Mature": 2, "X": 3}

var (
	errPolicyFile    = errors.New("DISCOVERY_POLICY_FILE")
	errPolicyMode    = errors.New("DISCOVERY_POLICY_MODE")
	errPolicyDecode  = errors.New("DISCOVERY_POLICY_DECODE")
	errPolicyLevel   = errors.New("DISCOVERY_POLICY_LEVEL")
	errPolicyTag     = errors.New("DISCOVERY_POLICY_TAG")
	errPolicyEntry   = errors.New("DISCOVERY_POLICY_EMPTY_ENTRY")
	errPolicyTooLong = errors.New("DISCOVERY_POLICY_TOO_LARGE")
)

// LoadPolicy reads the owner's policy. The file must be a regular file with no
// permission bit outside 0600; unknown fields are refused. Errors are bare
// codes and never quote policy values.
func LoadPolicy(path string) (Policy, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Policy{}, errPolicyFile
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&^0600 != 0 {
		return Policy{}, errPolicyMode
	}
	f, err := os.Open(path)
	if err != nil {
		return Policy{}, errPolicyFile
	}
	defer f.Close()
	// Re-check the opened file: the path may have been swapped after Lstat.
	if opened, err := f.Stat(); err != nil || !os.SameFile(info, opened) {
		return Policy{}, errPolicyMode
	}
	data, err := io.ReadAll(io.LimitReader(f, maxPolicyBytes+1))
	if err != nil {
		return Policy{}, errPolicyFile
	}
	if len(data) > maxPolicyBytes {
		return Policy{}, errPolicyTooLong
	}
	var p Policy
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return Policy{}, errPolicyDecode
	}
	if _, err := dec.Token(); err != io.EOF {
		return Policy{}, errPolicyDecode
	}
	if err := p.validate(); err != nil {
		return Policy{}, err
	}
	return p, nil
}

func (p Policy) validate() error {
	c := p.Civitai
	for _, list := range [][]string{c.Tags, c.Creators, c.ModelFamilies, c.AllowedLicenses, p.Reddit.Subreddits} {
		for _, v := range list {
			if strings.TrimSpace(v) == "" {
				return errPolicyEntry
			}
		}
	}
	for _, tag := range c.Tags {
		if _, err := strconv.ParseUint(tag, 10, 64); err != nil {
			return errPolicyTag
		}
	}
	if len(c.Tags)+len(c.Creators) > 0 {
		if _, ok := contentLevels[c.MaxContentLevel]; !ok {
			return errPolicyLevel
		}
	}
	return nil
}

// Admit decides whether a candidate may be downloaded or stored. reason is a
// short code, never content.
func (p Policy) Admit(c Candidate) (ok bool, reason string) {
	if c.MinorFlag {
		return false, "minor_flag"
	}
	switch c.Provider {
	case "civitai":
		return p.Civitai.admit(c)
	case "reddit":
		return p.Reddit.admit(c)
	}
	return false, "unknown_provider"
}

func (p CivitaiPolicy) admit(c Candidate) (bool, string) {
	if c.License == "" {
		return false, "no_licence"
	}
	if !slices.Contains(p.AllowedLicenses, c.License) {
		return false, "licence_not_allowed"
	}
	level, known := contentLevels[c.ContentLevel]
	max, maxKnown := contentLevels[p.MaxContentLevel]
	if !known || !maxKnown || level > max {
		return false, "content_level"
	}
	listed, familyOK := containsFold(p.Creators, c.Creator), len(p.ModelFamilies) == 0
	for _, tag := range c.Tags {
		if family, isFamily := strings.CutPrefix(tag, baseModelTag); isFamily {
			familyOK = familyOK || containsFold(p.ModelFamilies, family)
		} else if containsFold(p.Tags, tag) {
			listed = true
		}
	}
	// ModelFamilies only narrows the creator and tag allowlist.
	if !listed {
		return false, "not_allowlisted"
	}
	if !familyOK {
		return false, "model_family"
	}
	return true, ""
}

// admit accepts Reddit link metadata only: nothing is downloaded, so no
// licence is needed, but the subreddit (carried as the tag "r/<name>") must be
// listed.
func (p RedditPolicy) admit(c Candidate) (bool, string) {
	if c.MediaURL != "" {
		return false, "media_not_allowed"
	}
	for _, tag := range c.Tags {
		if name, ok := strings.CutPrefix(tag, "r/"); ok && containsFold(p.Subreddits, name) {
			return true, ""
		}
	}
	return false, "not_allowlisted"
}

func containsFold(list []string, v string) bool {
	if v == "" {
		return false
	}
	return slices.ContainsFunc(list, func(e string) bool { return strings.EqualFold(e, v) })
}
