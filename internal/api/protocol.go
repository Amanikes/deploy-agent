package api

import (
	"strings"
	"time"
)

type Command struct {
	ID      string            `json:"id"`
	Action  string            `json:"action"`
	Project string            `json:"project"`
	Payload map[string]string `json:"payload"`
}

type AckMessage struct {
	CommandID string            `json:"command_id"`
	Action    string            `json:"action"`
	Status    string            `json:"status"`
	Message   string            `json:"message,omitempty"`
	Project   string            `json:"project,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
	At        string            `json:"at"`
}

func NewAck(cmd Command, status string, message string, meta map[string]string) AckMessage {
	return AckMessage{
		CommandID: cmd.ID,
		Action:    normalizeAction(cmd.Action),
		Status:    status,
		Message:   message,
		Project:   cmd.Project,
		Meta:      meta,
		At:        time.Now().UTC().Format(time.RFC3339),
	}
}

func normalizeAction(action string) string {
	a := strings.ToLower(strings.TrimSpace(action))
	switch a {
	case "build":
		return "deploy"
	case "get_status":
		return "status"
	default:
		return a
	}
}
