package executor

import (
	"context"
	"io"
	"time"
)

func (GoTools) Clockfuse(ctx context.Context, root string) ([]byte, error) {
	return runGo(ctx, root, "go", 5*time.Minute, "run", "./tools/clockfuse", ".")
}

func (GoTools) ShortHead(ctx context.Context, root string) ([]byte, error) {
	return run(ctx, root, "git", 30*time.Second, io.Discard, "-c", "core.fsmonitor=false", "rev-parse", "--short", "HEAD")
}
