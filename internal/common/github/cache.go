package github

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// releaseCache provides caching for GitHub release information
type releaseCache struct {
	client      *client
	mu          sync.RWMutex
	release     *release
	lastFetch   time.Time
	ttl         time.Duration
	fallbackURL string
}

// newReleaseCache creates a new release cache with specified TTL
func newReleaseCache(ttl time.Duration, fallbackURL string) *releaseCache {
	return &releaseCache{
		client:      newClient(),
		ttl:         ttl,
		fallbackURL: fallbackURL,
	}
}

// getLatestRelease returns cached release or fetches new one if cache expired
func (rc *releaseCache) getLatestRelease(ctx context.Context, owner, repo string) *release {
	rc.mu.RLock()
	if rc.release != nil && time.Since(rc.lastFetch) < rc.ttl {
		release := rc.release
		rc.mu.RUnlock()
		return release
	}
	rc.mu.RUnlock()

	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.release != nil && time.Since(rc.lastFetch) < rc.ttl {
		return rc.release
	}

	release, err := rc.client.getLatestRelease(ctx, owner, repo)
	if err != nil {
		log.Error().Err(err).
			Str("owner", owner).
			Str("repo", repo).
			Msg("Failed to fetch latest release from GitHub")

		if rc.release != nil {
			log.Debug().Msg("Returning stale cached release due to fetch error")
			return rc.release
		}

		return nil
	}

	rc.release = release
	rc.lastFetch = time.Now()

	return release
}

// getDownloadURL returns the download URL for the extension ZIP file
func (rc *releaseCache) getDownloadURL(ctx context.Context, owner, repo string) string {
	release := rc.getLatestRelease(ctx, owner, repo)

	if release == nil {
		log.Debug().Msg("No release available, returning fallback URL")
		return rc.fallbackURL
	}

	zipURL := release.getZipAssetURL()
	if zipURL == "" {
		log.Debug().Msg("No ZIP asset found in release, returning fallback URL")
		return rc.fallbackURL
	}

	return zipURL
}

// clearCache forces a cache refresh on next request
func (rc *releaseCache) clearCache() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.release = nil
	rc.lastFetch = time.Time{}
}
