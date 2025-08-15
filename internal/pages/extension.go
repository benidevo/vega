package pages

import (
	"net/http"
	"net/url"

	"github.com/benidevo/vega/internal/common/github"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetExtensionDownload redirects to the latest extension download URL
func (h *Handler) GetExtensionDownload(c *gin.Context) {
	downloadURL := github.GetExtensionDownloadURL(c.Request.Context())

	if downloadURL != "" && downloadURL != github.ExtensionFallbackURL {
		u, err := url.Parse(downloadURL)
		if err != nil {
			log.Warn().Err(err).Str("url", downloadURL).Msg("Invalid download URL")
			c.Redirect(http.StatusTemporaryRedirect, github.ExtensionFallbackURL)
			return
		}

		// Whitelist of allowed GitHub hosts/subdomains
		allowedHosts := map[string]bool{
			"github.com":                true,
			"api.github.com":            true,
			"uploads.github.com":        true,
			"githubusercontent.com":     true,
			"raw.githubusercontent.com": true,
		}

		if !allowedHosts[u.Host] {
			log.Warn().Str("url", downloadURL).Str("host", u.Host).Msg("Download URL not from allowed GitHub domain")
			c.Redirect(http.StatusTemporaryRedirect, github.ExtensionFallbackURL)
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, downloadURL)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, github.ExtensionFallbackURL)
}
