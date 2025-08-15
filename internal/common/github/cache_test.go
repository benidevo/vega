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
