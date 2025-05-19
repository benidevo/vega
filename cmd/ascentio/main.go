package main

import (
	"log"

	"github.com/benidevo/ascentio/internal/ascentio"
	"github.com/benidevo/ascentio/internal/config"
)

func main() {
	cfg := config.NewSettings()
	app := ascentio.New(cfg)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to start the application: %v", err)
	}

	app.WaitForShutdown()
}
