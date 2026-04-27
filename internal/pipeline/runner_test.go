package pipeline

import (
	
	"testing"
)

func TestNewPipeline(t *testing.T) {
	dir := t.TempDir()
	p := NewPipeline(dir)
	if p.WorkDir != dir {
		t.Errorf("expected WorkDir %s, got %s", dir, p.WorkDir)
	}
}

func TestRunStep_Err(t *testing.T) {
	dir := t.TempDir()
	p := NewPipeline(dir)
	err := p.RunStep("false")
	if err == nil {
		t.Errorf("expected error for 'false' command")
	}
}
