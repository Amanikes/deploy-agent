package main

import (
	"auto-deploy-agent/internal/api"
	"auto-deploy-agent/internal/config"
	"log"
)

func main() {
	log.Println("Starting Auto Deploy Agent...")

	// Load config.
	cfg := config.Load()

	// Initialize API client and start the reconnecting listener.
	client := api.NewAgentClient(cfg)
	client.Listen()
}
