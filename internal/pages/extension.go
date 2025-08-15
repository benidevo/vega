package pages

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/benidevo/vega/internal/common/github"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetExtensionDownload redirects to the latest extension download URL
func (h *Handler) GetExtensionDownload(c *gin.Context) {
	downloadURL := github.GetExtensionDownloadURL(c.Request.Context())

	if downloadURL != "" && downloadURL != github.ExtensionFallbackURL {
		// Validate the URL is from GitHub to prevent open redirect
		u, err := url.Parse(downloadURL)
		if err != nil {
			log.Warn().Err(err).Str("url", downloadURL).Msg("Invalid download URL")
			c.Redirect(http.StatusTemporaryRedirect, github.ExtensionFallbackURL)
			return
		}

		// Ensure the host is github.com or a GitHub subdomain
		if !strings.HasSuffix(u.Host, "github.com") && !strings.HasSuffix(u.Host, "githubusercontent.com") {
			log.Warn().Str("url", downloadURL).Str("host", u.Host).Msg("Download URL not from GitHub")
			c.Redirect(http.StatusTemporaryRedirect, github.ExtensionFallbackURL)
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, downloadURL)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, github.ExtensionFallbackURL)
}
