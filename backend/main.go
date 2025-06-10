package main

import (
	"log"
	"site-availability/config"
	"site-availability/logging"
	"site-availability/server"
)

func main() {
	// Attempt to initialize the logger, and fall back to Go's log package if it fails
	if err := logging.Init(); err != nil {
		log.Fatalf("Logger initialization failed: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		logging.Logger.Fatalf("Failed to load configuration: %v", err)
	}

	srv := server.NewServer(cfg)
	if err := srv.Start(); err != nil {
		logging.Logger.Fatalf("Failed to start server: %v", err)
	}
}
