package executor

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// RecordOutput streams and drains a command without treating truncation as a
// target failure. Judge adapters continue to use their original 8 MiB limits.
func RecordOutput(ctx context.Context, repo string, argv []string, timeout time.Duration, ceiling int64, spool, diagnostic io.Writer) (int, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	c := exec.CommandContext(ctx, argv[0], argv[1:]...)
	c.Dir = repo
	c.Env = goEnvironment()
	for _, k := range []string{"GRYPE_DB_CACHE_DIR", "GRYPE_DB_AUTO_UPDATE", "GRYPE_CHECK_FOR_APP_UPDATE"} {
		if v, ok := os.LookupEnv(k); ok {
			c.Env = append(c.Env, k+"="+v)
		}
	}
	w := &recordWriter{limit: ceiling, spool: spool, diagnostic: diagnostic}
	c.Stdout = w
	c.Stderr = w
	c.WaitDelay = time.Second
	err := c.Run()
	if w.err != nil {
		return 2, w.total, w.err
	}
	if ctx.Err() != nil {
		return 2, w.total, ctx.Err()
	}
	if err == nil {
		return 0, w.total, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		code := ee.ExitCode()
		if status, ok := ee.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			code = 128 + int(status.Signal())
		}
		return code, w.total, nil
	}
	return 2, w.total, err
}

type recordWriter struct {
	mu                sync.Mutex
	limit, total      int64
	spool, diagnostic io.Writer
	err               error
}

func (w *recordWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := len(p)
	keep := min(int64(n), max(int64(0), w.limit-w.total))
	w.total += int64(n)
	if keep > 0 && w.err == nil {
		_, w.err = io.MultiWriter(w.spool, w.diagnostic).Write(p[:keep])
	}
	// Keep draining even after an output failure so a chatty child cannot hang.
	return n, nil
}
