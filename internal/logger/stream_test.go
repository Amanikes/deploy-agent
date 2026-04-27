package logger

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestStream(t *testing.T) {
	input := "line1\nline2\n"
	r := io.NopCloser(strings.NewReader(input))
	var buf bytes.Buffer
	logFunc := func(format string, v ...interface{}) {
		buf.WriteString(v[1].(string))
		buf.WriteString("\n")
	}
	// Patch log.Printf for test
	orig := logPrintf
	logPrintf = logFunc
	defer func() { logPrintf = orig }()
	Stream(r, "TEST")
	if !strings.Contains(buf.String(), "line1") || !strings.Contains(buf.String(), "line2") {
		t.Errorf("expected lines in output, got %s", buf.String())
	}
}

// Patchable log.Printf for test
var logPrintf = func(format string, v ...interface{}) {
	// default to real log.Printf
}
