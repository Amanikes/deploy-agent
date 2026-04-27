package api

import "testing"

func TestNormalizeAction(t *testing.T) {
	if normalizeAction("build") != "deploy" {
		t.Error("expected build -> deploy")
	}
	if normalizeAction("get_status") != "status" {
		t.Error("expected get_status -> status")
	}
	if normalizeAction("restart") != "restart" {
		t.Error("expected restart -> restart")
	}
}

func TestNewAck(t *testing.T) {
	cmd := Command{ID: "1", Action: "deploy", Project: "proj"}
	ack := NewAck(cmd, "progress", "msg", map[string]string{"k": "v"})
	if ack.CommandID != "1" || ack.Action != "deploy" || ack.Status != "progress" || ack.Message != "msg" || ack.Project != "proj" || ack.Meta["k"] != "v" {
		t.Error("unexpected ack fields")
	}
}
