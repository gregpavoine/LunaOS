package discovery

import "slices"

// Policy is the owner's allowlist and rights policy. Task 4 completes it; until
// then Admit fails closed on anything it cannot vouch for.
type Policy struct {
	Civitai CivitaiPolicy
	Reddit  RedditPolicy
}

type CivitaiPolicy struct {
	Tags, Creators, ModelFamilies []string
	MaxContentLevel               string
	AllowedLicenses               []string
}

type RedditPolicy struct {
	Subreddits []string
}

// Admit decides whether a candidate may be downloaded or stored. reason is a
// short code, never content.
func (p Policy) Admit(c Candidate) (ok bool, reason string) {
	if c.MinorFlag {
		return false, "minor_flag"
	}
	if c.MediaURL == "" {
		return true, ""
	}
	if c.Provider != "civitai" || c.License == "" {
		return false, "no_licence"
	}
	if !slices.Contains(p.Civitai.AllowedLicenses, c.License) {
		return false, "licence_not_allowed"
	}
	return true, ""
}
