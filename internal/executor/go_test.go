package executor

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGoEnvironmentIsBounded(t *testing.T) {
	t.Setenv("LICTOR_UNAPPROVED", "fixture")
	t.Setenv("GOCACHE", t.TempDir())
	t.Setenv("GOPROXY", "off")
	env := strings.Join(goEnvironment(), "\n")
	if strings.Contains(env, "LICTOR_UNAPPROVED") || !strings.Contains(env, "GOCACHE="+os.Getenv("GOCACHE")) || !strings.Contains(env, "GOPROXY=off") {
		t.Fatalf("Go environment allowlist did not preserve only allowed settings")
	}
}

func TestGoOutputLimitAndCancellation(t *testing.T) {
	_, err := runGo(t.Context(), t.TempDir(), os.Args[0], 10*time.Second, "-test.run=TestExecutorHelper", "--", "lictor-output")
	if err == nil || !strings.Contains(err.Error(), "output exceeds") {
		t.Fatalf("output bound: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = runGo(ctx, t.TempDir(), os.Args[0], time.Second, "-test.run=TestExecutorHelper")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}
