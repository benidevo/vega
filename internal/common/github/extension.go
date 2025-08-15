package github

import (
	"context"
	"sync"
	"time"
)

var (
	extensionCache     *releaseCache
	extensionCacheOnce sync.Once
)

const (
	extensionOwner = "benidevo"
	extensionRepo  = "vega-ai-extension"

	extensionCacheTTL = 15 * time.Minute
	// ExtensionFallbackURL is exported as it's used by the pages package for comparison
	ExtensionFallbackURL = "https://github.com/benidevo/vega-ai-extension/releases/latest"
)

// getExtensionReleaseCache returns the singleton extension release cache
func getExtensionReleaseCache() *releaseCache {
	extensionCacheOnce.Do(func() {
		extensionCache = newReleaseCache(extensionCacheTTL, ExtensionFallbackURL)
	})
	return extensionCache
}

// GetExtensionDownloadURL returns the direct download URL for the extension
func GetExtensionDownloadURL(ctx context.Context) string {
	cache := getExtensionReleaseCache()
	return cache.getDownloadURL(ctx, extensionOwner, extensionRepo)
}
