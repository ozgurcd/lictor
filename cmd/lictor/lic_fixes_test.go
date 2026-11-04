package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ozgurcd/lictor/internal/executor"
)

// LIC-01: a declared pin beats --unpinned; the refusal must say so and how to proceed.
func TestUnpinnedDoesNotOverrideDeclaredPinSaysSo(t *testing.T) {
	for _, command := range []string{"grype", "green", "clockfuse", "route"} {
		root := t.TempDir()
		declarePin(t, root, "v99.0.0")
		code, out, stderr := invoke(t, command, "--repo", root, "--json", "--unpinned")
		if code != 2 || out != "" || strings.Count(stderr, "\n") != 1 {
			t.Fatalf("%s: %d %q %q", command, code, out, stderr)
		}
		for _, want := range []string{
			"declares LICTOR_VERSION v99.0.0",
			"--unpinned does not override a declared pin",
			"brew install ozgurcd/tap/lictor",
			"remove LICTOR_VERSION from .github/workflows/ci.yml",
		} {
			if !strings.Contains(stderr, want) {
				t.Fatalf("%s: refusal lacks %q: %q", command, want, stderr)
			}
		}
	}
}

// Without --unpinned the mismatch line must not suggest a flag that cannot help.
func TestMismatchWithoutUnpinnedDoesNotMentionTheFlag(t *testing.T) {
	root := t.TempDir()
	declarePin(t, root, "v99.0.0")
	_, _, stderr := invoke(t, "green", "--repo", root, "--json")
	if strings.Contains(stderr, "--unpinned") {
		t.Fatalf("misleading hint: %q", stderr)
	}
}

// LIC-02: witness run names the inherited variables it drops, never their values.
func TestWitnessRunReportsDroppedEnvironment(t *testing.T) {
	t.Setenv("GOFLAGS", "-mod=mod")
	t.Setenv("GOMAXPROCS", "2")
	t.Setenv("LICTOR_TEST_SECRETISH", "do-not-print-this-value")
	t.Setenv("LICTOR_TEST_DECLARED", "declared-value-should-pass")
	for _, mode := range []string{"human", "json"} {
		t.Run(mode, func(t *testing.T) {
			dir := environmentFixture(t)
			args := []string{"witness", "run", "--repo", dir, "--unpinned", "--record", "GATE-RUN.txt", "--label", "dropped", "--as-of", "2026-10-05T00:00:00Z", "--env", "LICTOR_TEST_DECLARED"}
			if mode == "json" {
				args = append(args, "--json")
			}
			code, out, diag := environmentInvoke(t, append(args, "--", "probe=/usr/bin/true"))
			if code != 0 {
				t.Fatalf("exit=%d diag=%s", code, diag)
			}
			var notice string
			for line := range strings.SplitSeq(string(diag), "\n") {
				if strings.HasPrefix(line, "gate-witness: inherited variables not passed to targets") {
					notice = line
				}
			}
			if notice == "" {
				t.Fatalf("no dropped-variable notice in diagnostics: %q", diag)
			}
			_, list, _ := strings.Cut(notice, "): ")
			listed := map[string]bool{}
			for name := range strings.SplitSeq(list, ", ") {
				listed[name] = true
			}
			for _, name := range []string{"GOFLAGS", "GOMAXPROCS", "LICTOR_TEST_SECRETISH"} {
				if !listed[name] {
					t.Fatalf("notice lacks %s: %q", name, notice)
				}
			}
			for _, name := range []string{"LICTOR_TEST_DECLARED", "PATH", "HOME"} {
				if listed[name] {
					t.Fatalf("notice lists %s, which the target receives: %q", name, notice)
				}
			}
			for _, b := range [][]byte{out, diag} {
				if bytes.Contains(b, []byte("do-not-print-this-value")) || bytes.Contains(b, []byte("-mod=mod")) {
					t.Fatal("a dropped value leaked")
				}
			}
			record, err := os.ReadFile(filepath.Join(dir, "GATE-RUN.txt"))
			if err != nil || bytes.Contains(record, []byte("inherited variables")) {
				t.Fatalf("notice must stay out of the record: %v", err)
			}
		})
	}
}

// Human --all keeps the source driver's exact streams, so no notice there.
func TestWitnessRunAllHumanStaysSilent(t *testing.T) {
	t.Setenv("GOFLAGS", "-mod=mod")
	dir := environmentFixture(t)
	args := []string{"witness", "run", "--repo", dir, "--unpinned", "--record", "GATE-RUN.txt", "--label", "all", "--as-of", "2026-10-05T00:00:00Z", "--all", "--", "probe=/usr/bin/true"}
	code, out, diag := environmentInvoke(t, args)
	if code != 0 || bytes.Contains(append(out, diag...), []byte("inherited variables")) {
		t.Fatalf("exit=%d out=%q diag=%q", code, out, diag)
	}
}

// The README must list every allowlisted name, so the docs cannot drift.
func TestReadmeDocumentsEveryAllowedEnvironmentName(t *testing.T) {
	b, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range executor.AllowedEnvironment() {
		if !strings.Contains(string(b), "`"+name+"`") {
			t.Errorf("README does not document environment name %s", name)
		}
	}
}
