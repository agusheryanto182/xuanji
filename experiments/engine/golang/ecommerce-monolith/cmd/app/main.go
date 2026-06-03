package main

import (
	"log"

	"github.com/agusheryanto182/ecommerce-monolith/config"
	"github.com/agusheryanto182/ecommerce-monolith/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
