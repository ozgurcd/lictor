package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExecutorHelper(t *testing.T) {
	mode := os.Args[len(os.Args)-1]
	switch mode {
	case "lictor-output":
		_, _ = os.Stdout.Write(bytes.Repeat([]byte("x"), outputLimit+1))
		os.Exit(0)
	case "lictor-environment":
		cwd, _ := os.Getwd()
		fmt.Printf("cwd=%s unapproved=%s", cwd, os.Getenv("LICTOR_UNAPPROVED"))
		os.Exit(0)
	}
}

func TestEnvironmentAndSelectedDirectory(t *testing.T) {
	t.Setenv("LICTOR_UNAPPROVED", "fixture")
	root := t.TempDir()
	out, err := run(t.Context(), root, os.Args[0], 10*time.Second, nil, "-test.run=TestExecutorHelper", "--", "lictor-environment")
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "cwd="+canonical+" unapproved=" {
		t.Fatalf("unexpected child context: %s", out)
	}
}

func TestOutputLimitAndCancellationRefuse(t *testing.T) {
	_, err := run(t.Context(), t.TempDir(), os.Args[0], 10*time.Second, nil, "-test.run=TestExecutorHelper", "--", "lictor-output")
	if err == nil || !strings.Contains(err.Error(), "output exceeds") {
		t.Fatalf("output limit: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = run(ctx, t.TempDir(), os.Args[0], time.Second, nil, "-test.run=TestExecutorHelper")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}

func TestReportLimitRefusesBeforeReading(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "report")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(ReportLimit + 1); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadReport(f.Name()); err == nil {
		t.Fatal("oversized report accepted")
	}
}
