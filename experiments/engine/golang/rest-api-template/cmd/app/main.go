package main

import (
	"log"

	"github.com/agusheryanto182/go-rest-api-template/config"
	_ "github.com/agusheryanto182/go-rest-api-template/docs"
	"github.com/agusheryanto182/go-rest-api-template/internal/app"
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
