package logger

import (
	"bytes"
	"io"
	"log"
	"strings"
	"testing"
)

func TestStream(t *testing.T) {
	input := "line1\nline2\n"
	r := io.NopCloser(strings.NewReader(input))
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	Stream(r, "TEST")

	out := buf.String()
	if !strings.Contains(out, "line1") || !strings.Contains(out, "line2") {
		t.Errorf("expected lines in output, got %s", out)
	}
}
