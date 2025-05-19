package main

import (
	"fmt"
	"os"

	"github.com/benidevo/ascentio/internal/config"
	"github.com/benidevo/ascentio/internal/logger"
	"github.com/benidevo/ascentio/internal/scheduler"
)

func main() {
	cfg := config.NewSettings()
	logger.Initialize(
		cfg.IsDevelopment,
		cfg.LogLevel,
	)
	log := logger.GetLogger("scheduler")

	log.Info().Msg("Starting scheduler application")

	app := scheduler.NewApp(&cfg)

	if err := app.Run(os.Args[1:]); err != nil {
		log.Error().Err(err).Msg("Scheduler failed")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
