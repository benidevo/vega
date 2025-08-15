package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseCache_GetLatestRelease(t *testing.T) {
	var apiCalls int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		apiCalls++
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"tag_name": "v1.0.0",
			"name": "Release 1.0.0",
			"html_url": "https://github.com/owner/repo/releases/tag/v1.0.0",
			"assets": [
				{
					"name": "extension.zip",
					"browser_download_url": "https://github.com/owner/repo/releases/download/v1.0.0/extension.zip"
				}
			]
		}`))
	}))
	defer server.Close()

	cache := &releaseCache{
		client: &client{
			httpClient: &http.Client{Timeout: 10 * time.Second},
			baseURL:    server.URL,
		},
		ttl:         100 * time.Millisecond,
		fallbackURL: "https://fallback.url",
	}

	ctx := context.Background()

	release1 := cache.getLatestRelease(ctx, "owner", "repo")
	require.NotNil(t, release1, "Expected non-nil release on first call")
	assert.Equal(t, 1, apiCalls, "Expected 1 API call")

	release2 := cache.getLatestRelease(ctx, "owner", "repo")
	require.NotNil(t, release2, "Expected non-nil release on second call")
	assert.Equal(t, 1, apiCalls, "Expected still 1 API call (cached)")

	time.Sleep(150 * time.Millisecond)

	release3 := cache.getLatestRelease(ctx, "owner", "repo")
	require.NotNil(t, release3, "Expected non-nil release on third call")
	assert.Equal(t, 2, apiCalls, "Expected 2 API calls (cache expired)")
}

func TestReleaseCache_GetDownloadURL(t *testing.T) {
	tests := []struct {
		name         string
		responseCode int
		responseBody string
		expectedURL  string
	}{
		{
			name:         "successful with zip asset",
			responseCode: http.StatusOK,
			responseBody: `{
				"tag_name": "v1.0.0",
				"assets": [
					{
						"name": "extension.zip",
						"browser_download_url": "https://github.com/download/extension.zip"
					}
				]
			}`,
			expectedURL: "https://github.com/download/extension.zip",
		},
		{
			name:         "successful without zip asset",
			responseCode: http.StatusOK,
			responseBody: `{
				"tag_name": "v1.0.0",
				"assets": [
					{
						"name": "readme.txt",
						"browser_download_url": "https://github.com/download/readme.txt"
					}
				]
			}`,
			expectedURL: "https://fallback.url",
		},
		{
			name:         "no releases",
			responseCode: http.StatusNotFound,
			responseBody: `{"message": "Not Found"}`,
			expectedURL:  "https://fallback.url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			cache := &releaseCache{
				client: &client{
					httpClient: &http.Client{Timeout: 10 * time.Second},
					baseURL:    server.URL,
				},
				ttl:         1 * time.Hour,
				fallbackURL: "https://fallback.url",
			}

			ctx := context.Background()
			url := cache.getDownloadURL(ctx, "owner", "repo")

			assert.Equal(t, tt.expectedURL, url, "URL mismatch")
		})
	}
}

func TestReleaseCache_ClearCache(t *testing.T) {
	var apiCalls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name": "v1.0.0"}`))
	}))
	defer server.Close()

	cache := &releaseCache{
		client: &client{
			httpClient: &http.Client{Timeout: 10 * time.Second},
			baseURL:    server.URL,
		},
		ttl:         1 * time.Hour,
		fallbackURL: "https://fallback.url",
	}

	ctx := context.Background()

	cache.getLatestRelease(ctx, "owner", "repo")
	assert.Equal(t, 1, apiCalls, "Expected 1 API call")

	cache.getLatestRelease(ctx, "owner", "repo")
	assert.Equal(t, 1, apiCalls, "Expected still 1 API call (cached)")

	cache.clearCache()

	cache.getLatestRelease(ctx, "owner", "repo")
	assert.Equal(t, 2, apiCalls, "Expected 2 API calls (cache cleared)")
}

func TestReleaseCache_ConcurrentAccess(t *testing.T) {
	var apiCalls int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		apiCalls++
		callNum := apiCalls
		mu.Unlock()

		time.Sleep(50 * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"tag_name": "v` + string(rune('0'+callNum)) + `.0.0",
			"assets": [
				{
					"name": "extension.zip",
					"browser_download_url": "https://github.com/download/v` + string(rune('0'+callNum)) + `.zip"
				}
			]
		}`))
	}))
	defer server.Close()

	cache := &releaseCache{
		client: &client{
			httpClient: &http.Client{Timeout: 10 * time.Second},
			baseURL:    server.URL,
		},
		ttl:         50 * time.Millisecond,
		fallbackURL: "https://fallback.url",
	}

	ctx := context.Background()

	cache.getLatestRelease(ctx, "owner", "repo")
	assert.Equal(t, 1, apiCalls, "Expected 1 API call for initial fetch")

	time.Sleep(60 * time.Millisecond)

	var wg sync.WaitGroup
	results := make([]*release, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = cache.getLatestRelease(ctx, "owner", "repo")
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 2, apiCalls, "Expected only 2 total API calls despite concurrent access")

	for i := 0; i < 10; i++ {
		assert.NotNil(t, results[i], "All concurrent requests should receive a release")
	}
	v1Count := 0
	v2Count := 0
	for i := 0; i < 10; i++ {
		if results[i] != nil && len(results[i].Assets) > 0 {
			url := results[i].Assets[0].BrowserDownloadURL
			if url == "https://github.com/download/v1.zip" {
				v1Count++
			} else if url == "https://github.com/download/v2.zip" {
				v2Count++
			}
		}
	}

	assert.Greater(t, v2Count, 0, "At least one goroutine should get the new version")
	t.Logf("Results: %d got stale cache (v1), %d got new cache (v2)", v1Count, v2Count)
}

func TestReleaseCache_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name": "v1.0.0"}`))
	}))
	defer server.Close()

	cache := &releaseCache{
		client: &client{
			httpClient: &http.Client{Timeout: 10 * time.Second},
			baseURL:    server.URL,
		},
		ttl:         1 * time.Hour,
		fallbackURL: "https://fallback.url",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	go func() {
		bgCtx := context.Background()
		cache.getLatestRelease(bgCtx, "owner", "repo")
	}()

	time.Sleep(10 * time.Millisecond)

	result := cache.getLatestRelease(ctx, "owner", "repo")

	if ctx.Err() != nil {
		assert.Nil(t, result, "Expected nil result when context is cancelled")
		t.Log("Context cancelled as expected during concurrent refresh wait")
	}
}
