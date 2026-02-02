package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	githubAPIURL = "https://api.github.com/repos/nogo/gitree/releases/latest"
	userAgent    = "gitree-update-checker"
	timeout      = 5 * time.Second
)

// ReleaseAsset contains download information for a release artifact.
type ReleaseAsset struct {
	Name        string
	DownloadURL string
	Size        int64
}

// ReleaseInfo contains information about a GitHub release.
type ReleaseInfo struct {
	Version     string
	URL         string
	PublishedAt time.Time
	Assets      []ReleaseAsset
}

// githubAsset represents an asset in the GitHub API response.
type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// githubRelease represents the GitHub API response.
type githubRelease struct {
	TagName     string        `json:"tag_name"`
	HTMLURL     string        `json:"html_url"`
	PublishedAt string        `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

// CheckLatestRelease fetches the latest release from GitHub.
func CheckLatestRelease(ctx context.Context) (*ReleaseInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found")
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("rate limited by GitHub API")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	publishedAt, _ := time.Parse(time.RFC3339, release.PublishedAt)

	// Convert assets
	var assets []ReleaseAsset
	for _, a := range release.Assets {
		assets = append(assets, ReleaseAsset{
			Name:        a.Name,
			DownloadURL: a.BrowserDownloadURL,
			Size:        a.Size,
		})
	}

	return &ReleaseInfo{
		Version:     release.TagName,
		URL:         release.HTMLURL,
		PublishedAt: publishedAt,
		Assets:      assets,
	}, nil
}

// IsNewer returns true if latest version is newer than current.
// Compares semantic versions (v1.2.3 format).
func IsNewer(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Handle "dev" version - always consider releases newer
	if current == "dev" || current == "" {
		return latest != ""
	}

	currentParts := parseVersion(current)
	latestParts := parseVersion(latest)

	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

// parseVersion parses "1.2.3" into [1, 2, 3].
func parseVersion(v string) [3]int {
	var parts [3]int
	// Strip any suffix like -beta, -rc1, etc.
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	fmt.Sscanf(v, "%d.%d.%d", &parts[0], &parts[1], &parts[2])
	return parts
}
