// Package green ports the Identuum repository green floor, not its hook mode.
package green

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Tools permits deterministic executor refusal tests without changing PATH.
type Tools interface {
	Version(context.Context, string) ([]byte, error)
	Format(context.Context, string) ([]byte, error)
	Build(context.Context, string) ([]byte, error)
	Vet(context.Context, string) ([]byte, error)
	Test(context.Context, string) ([]byte, error)
}

type Result struct {
	SchemaVersion string `json:"schema_version"`
	Outcome       string `json:"outcome"`
	Cause         string `json:"cause"`
	Subject       string `json:"subject"`
	GoVersion     string `json:"go_version"`
	Platform      string `json:"platform"`
	Excerpt       string `json:"excerpt"`
	ExitCode      int    `json:"exit_code"`
}

func (r Result) Line() string {
	token := "go version unavailable"
	if fields := strings.Fields(r.GoVersion); len(fields) >= 4 {
		token = fields[2]
	}
	return fmt.Sprintf("%s: %s — subject %s; %s", r.Outcome, r.Cause, r.Subject, token)
}

func Refusal(root, cause string) Result {
	return Result{SchemaVersion: "lictor.green.v1", Outcome: "CANNOT-EVALUATE", Cause: cause, Subject: filepath.Base(filepath.Clean(root)), GoVersion: "go version unavailable", ExitCode: 2}
}

var testFailure = regexp.MustCompile(`^(--- FAIL|FAIL|\s+\S+_test\.go:)`)

// Excerpts keep the source's first twelve lines (filtered for go test), with
// an additional 16 KiB byte ceiling. The executor separately bounds capture.
func excerpt(out []byte, test bool) string {
	var lines []string
	for _, line := range strings.Split(strings.TrimRight(string(out), "\r\n"), "\n") {
		if !test || testFailure.MatchString(line) {
			lines = append(lines, line)
		}
		if len(lines) == 12 {
			break
		}
	}
	s := strings.Join(lines, "\n")
	if len(s) > 16*1024 {
		s = s[:16*1024]
	}
	return strings.ToValidUTF8(s, "?")
}

func Run(ctx context.Context, root string, tools Tools) Result {
	r := Refusal(root, "")
	version, err := tools.Version(ctx, root)
	if err != nil {
		r.Cause = "go version unavailable"
		if errors.Is(err, exec.ErrNotFound) {
			r.Cause = "no 'go' on PATH"
		}
		r.Excerpt = excerpt(append(version, []byte(err.Error())...), false)
		return r
	}
	r.GoVersion = strings.TrimSpace(string(version))
	if !strings.HasPrefix(r.GoVersion, "go version ") || strings.ContainsAny(r.GoVersion, "\r\n") {
		r.Cause, r.GoVersion = "unreadable go version", "go version unavailable"
		return r
	}
	fields := strings.Fields(r.GoVersion)
	if len(fields) >= 4 {
		r.Platform = fields[len(fields)-1]
	}
	st, err := os.Stat(filepath.Join(root, "go.mod"))
	if err != nil || !st.Mode().IsRegular() {
		r.Cause = "not a Go module"
		return r
	}
	for i, step := range []struct {
		cause string
		run   func(context.Context, string) ([]byte, error)
	}{
		{"gofmt FAILED — unformatted files", tools.Format},
		{"go build FAILED", tools.Build},
		{"go vet FAILED", tools.Vet},
		{"go test FAILED", tools.Test},
	} {
		out, err := step.run(ctx, root)
		if err == nil && (i != 0 || len(strings.TrimRight(string(out), "\n")) == 0) {
			continue
		}
		r.Cause, r.Outcome, r.ExitCode = step.cause, "NOT-GREEN", 1
		var exit *exec.ExitError
		if err != nil && !errors.As(err, &exit) {
			r.Cause, r.Outcome, r.ExitCode = step.cause+" — could not execute", "CANNOT-EVALUATE", 2
			out = append(out, []byte("\n"+err.Error())...)
		}
		r.Excerpt = excerpt(out, i == 3 && r.ExitCode == 1)
		return r
	}
	r.Outcome, r.Cause, r.ExitCode = "GREEN", "builds, vets and tests clean", 0
	return r
}
