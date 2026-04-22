package api

import (
	"auto-deploy-agent/internal/pipeline"
	"fmt"
	"log"
)

func (c *AgentClient) DispatchCommand(cmd Command) {
	action := normalizeAction(cmd.Action)
	cmd.Action = action
	log.Printf("Dispatching command [%s] action=%s project=%s", cmd.ID, action, cmd.Project)
	_ = c.SendAck(NewAck(cmd, "received", "command received", nil))
	_ = c.SendAck(NewAck(cmd, "started", "worker started", nil))

	var err error
	switch action {
	case "deploy":
		err = c.handleWithPipeline(cmd)
	case "restart":
		err = c.handleWithPipeline(cmd)
	case "status":
		err = c.handleWithPipeline(cmd)
	default:
		err = fmt.Errorf("unsupported action: %s", action)
	}

	if err != nil {
		log.Printf("Command [%s] failed: %v", cmd.ID, err)
		_ = c.SendAck(NewAck(cmd, "failed", err.Error(), nil))
		return
	}

	_ = c.SendAck(NewAck(cmd, "completed", "command completed", nil))
}

func (c *AgentClient) handleWithPipeline(cmd Command) error {
	result, err := pipeline.Execute(cmd.Action, cmd.Payload, func(update pipeline.ProgressUpdate) {
		meta := map[string]string{"step": update.Step}
		for k, v := range update.Meta {
			meta[k] = v
		}
		_ = c.SendAck(NewAck(cmd, "progress", update.Message, meta))
	})
	if err != nil {
		return err
	}
	if len(result) > 0 {
		_ = c.SendAck(NewAck(cmd, "progress", "result collected", result))
	}
	return nil
}
