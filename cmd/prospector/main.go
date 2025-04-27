package main

import "github.com/benidevo/prospector/internal/prospector"

func main() {
	config := prospector.NewConfig()
	app := prospector.New(config)
	defer app.Shutdown()

	app.Run()
}
