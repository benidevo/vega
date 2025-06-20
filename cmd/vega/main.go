package main

import (
	"log"

	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/vega"
)

func main() {
	cfg := config.NewSettings()
	app := vega.New(cfg)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to start the application: %v", err)
	}

	app.WaitForShutdown()
}
