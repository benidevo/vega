package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_GetLatestRelease(t *testing.T) {
	tests := []struct {
		name           string
		responseCode   int
		responseBody   string
		expectedError  bool
		expectedNil    bool
		expectedAssets int
	}{
		{
			name:         "successful response",
			responseCode: http.StatusOK,
			responseBody: `{
				"tag_name": "v1.0.0",
				"name": "Release 1.0.0",
				"html_url": "https://github.com/owner/repo/releases/tag/v1.0.0",
				"assets": [
					{
						"name": "extension.zip",
						"browser_download_url": "https://github.com/owner/repo/releases/download/v1.0.0/extension.zip"
					}
				]
			}`,
			expectedError:  false,
			expectedNil:    false,
			expectedAssets: 1,
		},
		{
			name:          "no releases found",
			responseCode:  http.StatusNotFound,
			responseBody:  `{"message": "Not Found"}`,
			expectedError: false,
			expectedNil:   true,
		},
		{
			name:          "server error",
			responseCode:  http.StatusInternalServerError,
			responseBody:  `{"message": "Internal Server Error"}`,
			expectedError: true,
			expectedNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method, "Expected GET request")
				assert.NotEmpty(t, r.Header.Get("User-Agent"), "User-Agent header not set")

				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &client{
				httpClient: &http.Client{Timeout: 10 * time.Second},
				baseURL:    server.URL,
			}

			ctx := context.Background()
			release, err := client.getLatestRelease(ctx, "owner", "repo")

			if tt.expectedError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")
			}

			if tt.expectedNil {
				assert.Nil(t, release, "Expected nil release")
			} else {
				assert.NotNil(t, release, "Expected non-nil release")
			}

			if release != nil {
				assert.Equal(t, tt.expectedAssets, len(release.Assets), "Assets count mismatch")
			}
		})
	}
}

func TestRelease_GetZipAssetURL(t *testing.T) {
	tests := []struct {
		name        string
		release     *release
		expectedURL string
	}{
		{
			name:        "nil release",
			release:     nil,
			expectedURL: "",
		},
		{
			name: "release with zip asset",
			release: &release{
				Assets: []releaseAsset{
					{Name: "file.txt", BrowserDownloadURL: "https://example.com/file.txt"},
					{Name: "extension.zip", BrowserDownloadURL: "https://example.com/extension.zip"},
					{Name: "readme.md", BrowserDownloadURL: "https://example.com/readme.md"},
				},
			},
			expectedURL: "https://example.com/extension.zip",
		},
		{
			name: "release without zip asset",
			release: &release{
				Assets: []releaseAsset{
					{Name: "file.txt", BrowserDownloadURL: "https://example.com/file.txt"},
					{Name: "readme.md", BrowserDownloadURL: "https://example.com/readme.md"},
				},
			},
			expectedURL: "",
		},
		{
			name:        "release with no assets",
			release:     &release{Assets: []releaseAsset{}},
			expectedURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.release.getZipAssetURL()
			assert.Equal(t, tt.expectedURL, url, "URL mismatch")
		})
	}
}
