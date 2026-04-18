package api

import (
	"auto-deploy-agent/pkg/models"
	"encoding/json"
	"log"
)

func (c *AgentClient) HandleIncomingMessage(rawMessage []byte) {
	var task models.TaskPayload
	//first parse the json message
	if err := json.Unmarshal(rawMessage, &task); err != nil {
		log.Printf("Error decoding message: %v", err)
		return
	}
	log.Printf("Processing task: [%s] Action: %s", task.ID, task.Action)

	//route the action
	switch task.Action {
	case models.ActionBuild:
		//TODO implement build logic
		log.Println("Triggering Pipeline Build...")
	case models.ActionDeploy:
		//TODO implement deploy logic
		log.Println("Triggering Docker Deployment...")
	case models.ActionRestart:
		//TODO implement restart logic
		log.Println("Triggering Container Restart...")
	case models.ActionStatus:
		//TODO implement status reporting logic
		log.Println("Fetching Deployment Status...")
	default:
		log.Printf("Unknown action: %s", task.Action)

	}
}

func (c *AgentClient) SendStatusReport() {
	//TODO implement status reporting logic
	log.Println("Sending health report to backend")
}
