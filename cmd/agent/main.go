package agent

import (
	"auto-deploy-agent/internal/api"
	"auto-deploy-agent/internal/config"
	"log"

	"golang.org/x/tools/go/cfg"
)

func main() {
	log.Println("Starting Auto Deploy Agent...")

	//Load config

	cfg := config.Load()

	//Initialize API client and connect
	client := api.NewAgentClient(cfg)
	client.Connect()

	//Keep the connection alive and listen for code changes

	for{
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("Connection lost. Reconnecting...: %v", err)
			break
		}
		log.Printf("Received message: %s", message)
	}
}
