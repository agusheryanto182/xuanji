package main

import (
	"log"

	"github.com/agusheryanto182/ecommerce-monolith/config"
	_ "github.com/agusheryanto182/ecommerce-monolith/docs"
	"github.com/agusheryanto182/ecommerce-monolith/internal/seed"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	seed.Run(cfg)
}
