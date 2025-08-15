package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// releaseAsset represents a GitHub release asset
type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// release represents a GitHub release
type release struct {
	Assets []releaseAsset `json:"assets"`
}

// client provides methods to interact with GitHub API
type client struct {
	httpClient *http.Client
	baseURL    string
}

// newClient creates a new GitHub API client
func newClient() *client {
	return &client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.github.com",
	}
}

// getLatestRelease fetches the latest release for a given repository
func (c *client) getLatestRelease(ctx context.Context, owner, repo string) (*release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.baseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Vega-AI-Application")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			log.Debug().Msg("No releases found for repository")
			return nil, nil
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &rel, nil
}

// getZipAssetURL returns the download URL for a .zip asset from a release
func (r *release) getZipAssetURL() string {
	if r == nil {
		return ""
	}

	for _, asset := range r.Assets {
		if len(asset.Name) > 4 && asset.Name[len(asset.Name)-4:] == ".zip" {
			return asset.BrowserDownloadURL
		}
	}

	return ""
}
