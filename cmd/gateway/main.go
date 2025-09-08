package main

import (
	"log"

	"github.com/mangooer/gamehub-arena/internal/config"
)

func main() {
	config, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Config: %+v", config)
	log.Printf("Starting gateway server on %s:%d", config.Server.Host, config.Server.Port)
}
