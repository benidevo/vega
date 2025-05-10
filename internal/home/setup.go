package home

import "github.com/benidevo/prospector/internal/config"

// Setup initializes and returns a new home Handler.
func Setup(cfg *config.Settings) *Handler {
	return NewHandler(cfg)
}
