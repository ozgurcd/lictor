package executor

import (
	"context"
	"io"
	"path/filepath"
	"time"
)

// RoutePolicy is trusted house policy, never a user-supplied command string.
type RoutePolicy struct {
	Pattern string
	Bans    []string
	Key     string
}
type Achta struct{ Workspace string }

func (Achta) Version(ctx context.Context, root string, stderr io.Writer) ([]byte, error) {
	return run(ctx, root, "achta", 30*time.Second, stderr, "version", "--json")
}

func (a Achta) Check(ctx context.Context, root string, policy RoutePolicy, wantJSON bool, stderr io.Writer) ([]byte, error) {
	args := []string{"declared-route", "check", "--dir", filepath.Join(root, ".github", "workflows"), "--route-cardinality", "per-file-any", "--route-pattern", policy.Pattern, "--required-route-pattern", policy.Pattern}
	for _, ban := range policy.Bans {
		args = append(args, "--ban-pattern", ban)
	}
	args = append(args, "--required-key", policy.Key, "--required-scope", "/env")
	if wantJSON {
		args = append(args, "--json")
	}
	if a.Workspace != "" {
		args = append([]string{"--workspace", a.Workspace}, args...)
	}
	return run(ctx, root, "achta", time.Minute, stderr, args...)
}
