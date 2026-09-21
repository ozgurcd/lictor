package executor

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"
)

// GoTools exposes only the fixed invocations used by the green floor.
// Its zero value uses the installed toolchain; it accepts no command strings.
type GoTools struct{}

func (GoTools) Version(ctx context.Context, root string) ([]byte, error) {
	return runGo(ctx, root, "go", 30*time.Second, "version")
}

func (GoTools) Format(ctx context.Context, root string) ([]byte, error) {
	return runGo(ctx, root, "gofmt", 5*time.Minute, "-l", ".")
}

func (GoTools) Build(ctx context.Context, root string) ([]byte, error) {
	return runGo(ctx, root, "go", 5*time.Minute, "build", "./...")
}

func (GoTools) Vet(ctx context.Context, root string) ([]byte, error) {
	return runGo(ctx, root, "go", 5*time.Minute, "vet", "./...")
}

func (GoTools) Test(ctx context.Context, root string) ([]byte, error) {
	return runGo(ctx, root, "go", 5*time.Minute, "test", "./...", "-count=1", "-timeout=120s")
}

func goEnvironment() []string {
	var env []string
	for _, name := range []string{
		"PATH", "HOME", "TMPDIR", "XDG_CACHE_HOME", "GOCACHE", "GOMODCACHE",
		"GOPATH", "GOROOT", "GOENV", "GOTOOLCHAIN", "GOWORK", "GOFLAGS",
		"GOPROXY", "GOSUMDB", "GOPRIVATE", "GONOPROXY", "GONOSUMDB",
		"GOOS", "GOARCH", "CGO_ENABLED", "CC", "CXX", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET",
	} {
		if value, ok := os.LookupEnv(name); ok {
			env = append(env, name+"="+value)
		}
	}
	return append(env, "GIT_OPTIONAL_LOCKS=0", "LC_ALL=C")
}

func runGo(ctx context.Context, root, name string, timeout time.Duration, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir, cmd.Env, cmd.WaitDelay = root, goEnvironment(), time.Second
	var out boundedBuffer
	// One writer preserves combined stdout/stderr order, as the source's 2>&1.
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	if out.exceeded {
		return out.Bytes(), errors.New("tool output exceeds 8 MiB")
	}
	if ctx.Err() != nil {
		return out.Bytes(), ctx.Err()
	}
	return out.Bytes(), err
}
