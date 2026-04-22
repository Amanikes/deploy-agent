package api

import (
	"auto-deploy-agent/internal/config"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type AgentClient struct {
	Conn    *websocket.Conn
	Config  *config.Config
	writeMu sync.Mutex
	connMu  sync.RWMutex
}

func NewAgentClient(cfg *config.Config) *AgentClient {
	return &AgentClient{
		Config: cfg,
	}
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
		c.setConn(conn)
		log.Println("Connected to backend")
		return nil
	}

}

func (c *AgentClient) Listen() {
	for {
		if c.getConn() == nil {
			if err := c.Connect(); err != nil {
				log.Printf("Connect returned error: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
		}

		log.Println("Listening for commands...")
		conn := c.getConn()
		if conn == nil {
			continue
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Connection lost: %v. Reconnecting...", err)
			c.closeConn()
			continue
		}

		var cmd Command

		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Printf("Failed to parse command: %v", err)
			_ = c.SendAck(AckMessage{
				Status:  "invalid",
				Message: "invalid command payload",
				At:      time.Now().UTC().Format(time.RFC3339),
			})
			continue
		}

		if cmd.ID == "" {
			cmd.ID = time.Now().UTC().Format("20060102150405.000000000")
		}

		go c.DispatchCommand(cmd)
	}
}

func (c *AgentClient) SendAck(ack AckMessage) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	conn := c.getConn()
	if conn == nil {
		return errors.New("no active websocket connection")
	}

	if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return err
	}

	if err := conn.WriteJSON(ack); err != nil {
		return err
	}

	return nil
}

func (c *AgentClient) setConn(conn *websocket.Conn) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	c.Conn = conn
}

func (c *AgentClient) getConn() *websocket.Conn {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.Conn
}

func (c *AgentClient) closeConn() {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.Conn != nil {
		_ = c.Conn.Close()
		c.Conn = nil
	}
}
