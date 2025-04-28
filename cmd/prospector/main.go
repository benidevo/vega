package main

import (
	"log"

	"github.com/benidevo/prospector/internal/config"
	"github.com/benidevo/prospector/internal/prospector"
)

func main() {
	cfg := config.NewSettings()
	app := prospector.New(cfg)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to start the application: %v", err)
	}

	app.WaitForShutdown()
}
