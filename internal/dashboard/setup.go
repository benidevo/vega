package dashboard

import "github.com/benidevo/prospector/internal/config"

// Setup initializes and returns a new dashboard Handler.
func Setup(cfg *config.Settings) *Handler {
	return NewHandler(cfg)
}
