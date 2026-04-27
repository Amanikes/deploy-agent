package api

import (
	"context"
	"testing"
	"time"
	"auto-deploy-agent/internal/pipeline"
)

type fakeClient struct {
	sent []string
}
func (f *fakeClient) SendAck(msg interface{}) error {
	f.sent = append(f.sent, "ack")
	return nil
}

func TestHandleWithPipelineCancelable_Cancel(t *testing.T) {
	c := &AgentClient{}
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
	if err == nil || ctx.Err() == nil {
		t.Errorf("expected cancellation error, got %v", err)
	}
}
