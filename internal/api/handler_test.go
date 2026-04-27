package api

import (
	"auto-deploy-agent/internal/pipeline"
	"context"
	"testing"
	"time"
)

func TestHandleWithPipelineCancelable_Cancel(t *testing.T) {
	cmd := Command{Action: "deploy", Payload: map[string]string{"deploy_cmd": "sleep 2"}}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	var err error
	go func() {
		// Use pipeline.ExecuteWithContext directly for isolation
		_, err = pipeline.ExecuteWithContext(ctx, cmd.Action, cmd.Payload, nil)
		close(done)
	}()
	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done
	if err == nil {
		t.Errorf("expected cancellation error, got %v", err)
	}
}
