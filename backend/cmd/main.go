package main

import (
	"context"
	"runner-manager-backend/internal/app"
	"runner-manager-backend/internal/config"
)

func main() {
	cfg, err := config.LoadConfig("config")
	if err != nil {
		// If an error occurs while loading the configuration, panic with the error.
		panic(err)
	}

	app.NewApp(context.Background(), cfg).Run()
}
