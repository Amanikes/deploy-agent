package api

import (
	"testing"
	"auto-deploy-agent/internal/config"
)

func TestNewAgentClient(t *testing.T) {
	cfg := &config.Config{BackendUrl: "url", AgentID: "id", AgentToken: "tok"}
	c := NewAgentClient(cfg)
	if c.Config != cfg {
		t.Error("config not set")
	}
	if c.tasks == nil {
		t.Error("tasks map not initialized")
	}
}
