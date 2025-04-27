package main

import (
	"log"

	"github.com/benidevo/prospector/internal/prospector"
)

func main() {
	config := prospector.NewConfig()
	app := prospector.New(config)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to start the application: %v", err)
	}

	app.WaitForShutdown()
}
