package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/version"
)

const (
	githubOwner = "meta222888"
	githubRepo  = "RepayPane"
)

type Release struct {
	TagName string
	HTMLURL string
	Version string
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func FetchLatestRelease() (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", githubOwner, githubRepo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "RelayPane/"+version.Version)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no release found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: %s", strings.TrimSpace(string(body)))
	}

	var raw githubRelease
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	ver := normalizeVersion(raw.TagName)
	if ver == "" {
		return nil, fmt.Errorf("invalid release tag: %q", raw.TagName)
	}
	htmlURL := raw.HTMLURL
	if htmlURL == "" {
		htmlURL = fmt.Sprintf("https://github.com/%s/%s/releases/latest", githubOwner, githubRepo)
	}
	return &Release{TagName: raw.TagName, HTMLURL: htmlURL, Version: ver}, nil
}

func normalizeVersion(tag string) string {
	tag = strings.TrimSpace(tag)
	tag = strings.TrimPrefix(tag, "v")
	tag = strings.TrimPrefix(tag, "V")
	return tag
}

// IsNewer reports whether remote is newer than local (semver-like x.y.z).
func IsNewer(local, remote string) bool {
	return compareVersions(normalizeVersion(local), normalizeVersion(remote)) < 0
}

func compareVersions(a, b string) int {
	ap := parseParts(a)
	bp := parseParts(b)
	n := len(ap)
	if len(bp) > n {
		n = len(bp)
	}
	for i := 0; i < n; i++ {
		av, bv := 0, 0
		if i < len(ap) {
			av = ap[i]
		}
		if i < len(bp) {
			bv = bp[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

func parseParts(v string) []int {
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			out = append(out, 0)
			continue
		}
		n := 0
		for _, ch := range p {
			if ch < '0' || ch > '9' {
				break
			}
			n = n*10 + int(ch-'0')
		}
		out = append(out, n)
	}
	return out
}
