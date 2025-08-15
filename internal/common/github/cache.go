package github

import (
	"context"
	"sync"
	"sync/atomic"
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
	refreshing  int32 // atomic flag to prevent concurrent refreshes
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

	if !atomic.CompareAndSwapInt32(&rc.refreshing, 0, 1) {
		rc.mu.RLock()
		release := rc.release
		rc.mu.RUnlock()
		if release != nil {
			log.Debug().Msg("Returning stale cache while another goroutine refreshes")
			return release
		}
		for atomic.LoadInt32(&rc.refreshing) == 1 {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(10 * time.Millisecond):
			}
		}
		rc.mu.RLock()
		release = rc.release
		rc.mu.RUnlock()
		return release
	}
	defer atomic.StoreInt32(&rc.refreshing, 0)

	rc.mu.RLock()
	if rc.release != nil && time.Since(rc.lastFetch) < rc.ttl {
		release := rc.release
		rc.mu.RUnlock()
		return release
	}
	rc.mu.RUnlock()

	release, err := rc.client.getLatestRelease(ctx, owner, repo)
	if err != nil {
		log.Error().Err(err).
			Str("owner", owner).
			Str("repo", repo).
			Msg("Failed to fetch latest release from GitHub")

		rc.mu.RLock()
		oldRelease := rc.release
		rc.mu.RUnlock()

		if oldRelease != nil {
			log.Debug().Msg("Returning stale cached release due to fetch error")
			return oldRelease
		}

		return nil
	}

	rc.mu.Lock()
	rc.release = release
	rc.lastFetch = time.Now()
	rc.mu.Unlock()

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
