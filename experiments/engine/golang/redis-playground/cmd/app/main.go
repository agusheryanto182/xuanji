package main

import (
	"log"

	"github.com/agusheryanto182/redis-playground/config"
	_ "github.com/agusheryanto182/redis-playground/docs"
	"github.com/agusheryanto182/redis-playground/internal/app"
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
