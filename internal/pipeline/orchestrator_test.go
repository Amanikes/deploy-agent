package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExecuteWithContext_Deploy_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	progressCalled := false
	progress := func(update ProgressUpdate) {
		progressCalled = true
	}

	payload := map[string]string{
		"repo_dir": "",
		"repo_url": "https://github.com/example/repo.git",
		"deploy_cmd": "sleep 2", // Simulate long-running
	}

	done := make(chan struct{})
	var err error
	go func() {
		_, err = ExecuteWithContext(ctx, "deploy", payload, progress)
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if !progressCalled {
		t.Errorf("progress callback was not called")
	}
}

func TestExecuteWithContext_UnsupportedAction(t *testing.T) {
	ctx := context.Background()
	_, err := ExecuteWithContext(ctx, "unknown", map[string]string{}, nil)
	if err == nil || err.Error() != "unsupported action: unknown" {
		t.Errorf("expected unsupported action error, got %v", err)
	}
}

func TestExecuteWithContext_StatusCmd(t *testing.T) {
	ctx := context.Background()
	payload := map[string]string{
		"status_cmd": "echo status-ok",
	}
	var gotOutput string
	progress := func(update ProgressUpdate) {}
	result, err := ExecuteWithContext(ctx, "status", payload, progress)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	gotOutput = result["output"]
	if gotOutput != "status-ok" {
		t.Errorf("expected output 'status-ok', got '%s'", gotOutput)
	}
}
