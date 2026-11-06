package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePolicy(t *testing.T, body string, mode os.FileMode) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "policy.json")
	if err := os.WriteFile(p, []byte(body), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, mode); err != nil {
		t.Fatal(err)
	}
	return p
}

const validPolicy = `{"civitai":{"tags":["4"],"creators":["Ajuro"],"model_families":["SDXL 1.0"],"max_content_level":"Mature","allowed_licenses":["provider-terms:civitai"]},"reddit":{"subreddits":["x"]}}`

func TestPolicyFileModeEnforced(t *testing.T) {
	for _, mode := range []os.FileMode{0644, 0640, 0604, 0700, 0666} {
		if _, err := LoadPolicy(writePolicy(t, validPolicy, mode)); err == nil {
			t.Errorf("mode %o must be refused", mode)
		}
	}
	dir := t.TempDir()
	if _, err := LoadPolicy(dir); err == nil {
		t.Error("a directory must be refused")
	}
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(writePolicy(t, validPolicy, 0600), link); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPolicy(link); err == nil {
		t.Error("a symlink must be refused")
	}
	for _, mode := range []os.FileMode{0600, 0400} {
		p, err := LoadPolicy(writePolicy(t, validPolicy, mode))
		if err != nil {
			t.Fatalf("mode %o: %v", mode, err)
		}
		if p.Civitai.MaxContentLevel != "Mature" || p.Civitai.Creators[0] != "Ajuro" || p.Civitai.Tags[0] != "4" ||
			p.Civitai.ModelFamilies[0] != "SDXL 1.0" || p.Civitai.AllowedLicenses[0] != civitaiLicense || p.Reddit.Subreddits[0] != "x" {
			t.Fatalf("decoded %+v", p)
		}
	}
}

func TestLoadPolicyRejectsInvalid(t *testing.T) {
	for name, body := range map[string]string{
		"unknown field":   `{"civitai":{"tags":["4"],"max_content_level":"X","nsfw":true}}`,
		"unknown top":     `{"civitai":{},"pornhub":{}}`,
		"no level":        `{"civitai":{"tags":["4"]}}`,
		"bad level":       `{"civitai":{"creators":["Ajuro"],"max_content_level":"XXX"}}`,
		"tag not an id":   `{"civitai":{"tags":["anime"],"max_content_level":"X"}}`,
		"empty entry":     `{"civitai":{"creators":[""],"max_content_level":"X"}}`,
		"empty subreddit": `{"reddit":{"subreddits":[" "]}}`,
		"trailing data":   validPolicy + `{}`,
		"not json":        `civitai`,
	} {
		_, err := LoadPolicy(writePolicy(t, body, 0600))
		if err == nil {
			t.Errorf("%s: must be refused", name)
			continue
		}
		if !strings.HasPrefix(err.Error(), "DISCOVERY_POLICY_") || strings.Contains(err.Error(), "anime") || strings.Contains(err.Error(), "Ajuro") {
			t.Errorf("%s: error must be a bare code, got %q", name, err)
		}
	}
	// An empty allowlist needs no level: nothing will be queried.
	if _, err := LoadPolicy(writePolicy(t, `{"civitai":{},"reddit":{"subreddits":[]}}`, 0600)); err != nil {
		t.Fatalf("empty policy: %v", err)
	}
}

func civitaiCand(level string) Candidate {
	return Candidate{
		Item:     Item{Provider: "civitai", ExternalID: "1", Creator: "Ajuro", License: civitaiLicense, ContentLevel: level, Tags: []string{baseModelTag + "SDXL 1.0"}},
		MediaURL: "https://image.civitai.test/1.jpeg",
	}
}

func testPolicy(max string) Policy {
	return Policy{
		Civitai: CivitaiPolicy{Creators: []string{"ajuro"}, Tags: []string{"4"}, ModelFamilies: []string{"SDXL 1.0"}, MaxContentLevel: max, AllowedLicenses: []string{civitaiLicense}},
		Reddit:  RedditPolicy{Subreddits: []string{"Pics"}},
	}
}

func TestMinorFlagRejected(t *testing.T) {
	p := testPolicy("X")
	c := civitaiCand("None")
	if ok, _ := p.Admit(c); !ok {
		t.Fatal("control candidate must be admitted")
	}
	c.MinorFlag = true
	if ok, reason := p.Admit(c); ok || reason != "minor_flag" {
		t.Fatalf("civitai minor: %v %q", ok, reason)
	}
	r := Candidate{Item: Item{Provider: "reddit", ExternalID: "t3_a", Tags: []string{"r/pics"}}}
	if ok, reason := p.Admit(r); !ok {
		t.Fatalf("reddit control: %q", reason)
	}
	r.MinorFlag = true
	if ok, reason := p.Admit(r); ok || reason != "minor_flag" {
		t.Fatalf("reddit minor: %v %q", ok, reason)
	}
}

func TestMissingLicenseRejected(t *testing.T) {
	p := testPolicy("X")
	c := civitaiCand("Soft")
	c.License = ""
	if ok, reason := p.Admit(c); ok || reason != "no_licence" {
		t.Fatalf("no licence: %v %q", ok, reason)
	}
	c.License = "cc-by"
	if ok, reason := p.Admit(c); ok || reason != "licence_not_allowed" {
		t.Fatalf("unlisted licence: %v %q", ok, reason)
	}
	// The provider-terms token is only accepted when the owner lists it.
	p.Civitai.AllowedLicenses = nil
	if ok, _ := p.Admit(civitaiCand("Soft")); ok {
		t.Fatal("empty AllowedLicenses must admit nothing")
	}
	p.Civitai.AllowedLicenses = []string{"Provider-Terms:Civitai"}
	if ok, _ := p.Admit(civitaiCand("Soft")); ok {
		t.Fatal("the licence token must match exactly")
	}

	// Reddit is metadata-only and needs no licence, but never media and only listed subreddits.
	p = testPolicy("X")
	r := Candidate{Item: Item{Provider: "reddit", ExternalID: "t3_a", Tags: []string{"r/PICS"}}}
	if ok, reason := p.Admit(r); !ok {
		t.Fatalf("reddit metadata: %q", reason)
	}
	r.MediaURL = "https://i.redd.it/a.jpg"
	r.License = civitaiLicense
	if ok, reason := p.Admit(r); ok || reason != "media_not_allowed" {
		t.Fatalf("reddit media: %v %q", ok, reason)
	}
	r.MediaURL = ""
	r.Tags = []string{"r/other", "pics"}
	if ok, reason := p.Admit(r); ok || reason != "not_allowlisted" {
		t.Fatalf("reddit unlisted: %v %q", ok, reason)
	}
	if ok, reason := p.Admit(Candidate{Item: Item{Provider: "pornhub", License: civitaiLicense}}); ok || reason != "unknown_provider" {
		t.Fatalf("unknown provider: %v %q", ok, reason)
	}
}

func TestContentLevelAboveMaxRejected(t *testing.T) {
	p := testPolicy("Soft")
	for level, want := range map[string]bool{"None": true, "Soft": true, "Mature": false, "X": false, "": false, "soft": false, "XXX": false} {
		if ok, _ := p.Admit(civitaiCand(level)); ok != want {
			t.Errorf("level %q admitted=%v, want %v", level, ok, want)
		}
	}
	p.Civitai.MaxContentLevel = ""
	if ok, _ := p.Admit(civitaiCand("None")); ok {
		t.Error("an unset maximum must admit nothing")
	}

	// The allowlist binds as well: creator (any case) or tag id, and the base
	// model must be listed when ModelFamilies narrows the allowlist.
	p = testPolicy("X")
	c := civitaiCand("X")
	c.Creator = "someone"
	if ok, reason := p.Admit(c); ok || reason != "not_allowlisted" {
		t.Fatalf("a listed base model alone must not admit: %v %q", ok, reason)
	}
	c.Tags = []string{"4", baseModelTag + "sdxl 1.0"}
	if ok, _ := p.Admit(c); !ok {
		t.Fatal("listed tag and base model must be admitted")
	}
	for _, tags := range [][]string{{"4", baseModelTag + "Pony"}, {"4"}} {
		c.Tags = tags
		if ok, reason := p.Admit(c); ok || reason != "model_family" {
			t.Fatalf("tags %q: %v %q", tags, ok, reason)
		}
	}
	p.Civitai.ModelFamilies = nil
	if ok, _ := p.Admit(c); !ok {
		t.Fatal("without ModelFamilies any base model is admitted")
	}
}
