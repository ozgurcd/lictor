// Package executor contains the fixed command lines Lictor is allowed to run.
package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

const outputLimit = 8 << 20
const ReportLimit = 64 << 20

type boundedBuffer struct {
	buffer   bytes.Buffer
	exceeded bool
}

func (b *boundedBuffer) Bytes() []byte { return b.buffer.Bytes() }

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(p) > outputLimit-b.buffer.Len() {
		b.exceeded = true
		return 0, errors.New("tool output exceeds 8 MiB")
	}
	return b.buffer.Write(p)
}

// Environment carries paths and scanner cache controls, never credentials or
// configuration overrides for vulnerability, exclusion, or database-age policy.
func Environment() []string {
	var env []string
	for _, name := range []string{"PATH", "HOME", "TMPDIR", "XDG_CACHE_HOME", "GRYPE_DB_CACHE_DIR", "GRYPE_DB_AUTO_UPDATE", "GRYPE_CHECK_FOR_APP_UPDATE"} {
		if value, ok := os.LookupEnv(name); ok {
			env = append(env, name+"="+value)
		}
	}
	return append(env, "GIT_OPTIONAL_LOCKS=0", "LC_ALL=C")
}

func run(ctx context.Context, dir, name string, timeout time.Duration, stderr io.Writer, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir, cmd.Env = dir, Environment()
	cmd.WaitDelay = time.Second
	var out, diagnostic boundedBuffer
	cmd.Stdout, cmd.Stderr = &out, &diagnostic
	err := cmd.Run()
	if stderr != nil {
		if _, writeErr := stderr.Write(diagnostic.Bytes()); writeErr != nil {
			return nil, fmt.Errorf("write tool diagnostic: %w", writeErr)
		}
	}
	if out.exceeded || diagnostic.exceeded {
		return nil, errors.New("tool output exceeds 8 MiB")
	}
	if ctx.Err() != nil {
		return nil, fmt.Errorf("%s: %w", name, ctx.Err())
	}
	return out.Bytes(), err
}

// Scan asks Grype for the same pair of reports as the source judge.
func Scan(ctx context.Context, root, report, inventory string, stderr io.Writer) error {
	_, err := run(ctx, root, "grype", 5*time.Minute, stderr,
		"dir:.", "--output", "json="+report, "--output", "cyclonedx-json="+inventory)
	return err
}

// Ignored asks Git for ignored paths without index refresh writes or fsmonitor.
func Ignored(ctx context.Context, root string) ([]byte, error) {
	return run(ctx, root, "git", 30*time.Second, nil,
		"-c", "core.fsmonitor=false", "-C", root, "status", "--ignored", "--porcelain")
}

// ReadReport bounds allocations for supplied reports and scanner output files.
func ReadReport(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() || st.Size() > ReportLimit {
		return nil, fmt.Errorf("%s must be a regular file no larger than 64 MiB", path)
	}
	b, err := io.ReadAll(io.LimitReader(f, ReportLimit+1))
	if len(b) > ReportLimit {
		return nil, fmt.Errorf("%s exceeds 64 MiB", path)
	}
	return b, err
}
