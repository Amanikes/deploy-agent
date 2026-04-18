package api

import (
	"auto-deploy-agent/internal/config"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type AgentClient struct {
	Conn   *websocket.Conn
	Config *config.Config
}

func newAgentClient(cfg *config.Config) (*AgentClient, error) {
	return &AgentClient{
		Config: cfg,
	}, nil
}

func (c *AgentClient) Connect() error {
	header := http.Header{}
	header.Add("X-Agent-ID", c.Config.AgentID)
	header.Add("X-Agent-Token", c.Config.AgentToken)

	for {
		log.Printf("Attempting to connect to %s", c.Config.BackendUrl)
		conn, _, err := websocket.DefaultDialer.Dial(c.Config.BackendUrl, header)
		if err != nil {
			log.Printf("Connection failed: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		c.Conn = conn
		log.Println("Connected to backend")
		return nil
	}

}
