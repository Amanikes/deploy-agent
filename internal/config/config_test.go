package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("FOO", "bar")
	if getEnv("FOO", "baz") != "bar" {
		t.Error("expected bar")
	}
	os.Unsetenv("FOO")
	if getEnv("FOO", "baz") != "baz" {
		t.Error("expected baz")
	}
}
