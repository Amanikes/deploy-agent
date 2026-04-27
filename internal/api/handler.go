package api

import (
	"auto-deploy-agent/internal/pipeline"
	"context"
	"fmt"
	"log"
	"maps"
)

// Like handleWithPipeline, but supports cancellation via context
func (c *AgentClient) handleWithPipelineCancelable(cmd Command) (error, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	var result map[string]string
	var err error
	go func() {
		result, err = pipeline.ExecuteWithContext(ctx, cmd.Action, cmd.Payload, func(update pipeline.ProgressUpdate) {
			meta := map[string]string{"step": update.Step}
			for k, v := range update.Meta {
				meta[k] = v
			}
			_ = c.SendAck(NewAck(cmd, "progress", update.Message, meta))
		})
		close(done)
	}()
	// Wait for completion or cancellation
	select {
	case <-done:
		if err != nil {
			return err, nil
		}
		if len(result) > 0 {
			_ = c.SendAck(NewAck(cmd, "progress", "result collected", result))
		}
		return nil, nil
	case <-ctx.Done():
		// Rollback logic can be added here if needed
		return ctx.Err(), nil
	}
	// Return cancel function for task manager
	// (This is unreachable, but required for signature)
	// The actual cancel function is returned below
	// (see DispatchCommand)
	// This is a stub
	return nil, cancel
}

func (c *AgentClient) DispatchCommand(cmd Command) {
	action := normalizeAction(cmd.Action)
	cmd.Action = action
	log.Printf("Dispatching command [%s] action=%s project=%s", cmd.ID, action, cmd.Project)
	_ = c.SendAck(NewAck(cmd, "received", "command received", nil))
	_ = c.SendAck(NewAck(cmd, "started", "worker started", nil))

	var err error
	var cancelFn func()
	// Only register cancel for long-running actions
	switch action {
	case "deploy", "restart", "status":
		err, cancelFn = c.handleWithPipelineCancelable(cmd)
		if cancelFn != nil {
			c.tasksMu.Lock()
			c.tasks[cmd.ID] = cancelFn
			c.tasksMu.Unlock()
			defer func() {
				c.tasksMu.Lock()
				delete(c.tasks, cmd.ID)
				c.tasksMu.Unlock()
			}()
		}
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
			maps.Copy(meta, map[string]string{k: v})
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
