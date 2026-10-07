package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var environmentCLI = flag.String("environment-cli", "", "explicit released CLI for environment proofs")

func environmentValue() string {
	return "postgres://fixture:" + strings.Repeat("synthetic", 5) + "@localhost/fixture"
}

func TestDeclaredEnvironmentChild(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "--lictor-env-child" {
		return
	}
	v, ok := os.LookupEnv("LICTOR_TEST_DATABASE_URL")
	_, absent := os.LookupEnv("LICTOR_TEST_ABSENT")
	_, hidden := os.LookupEnv("LICTOR_TEST_UNDECLARED")
	empty, emptyPresent := os.LookupEnv("LICTOR_TEST_EMPTY")
	if !ok || v != environmentValue() || absent || hidden || !emptyPresent || empty != "" {
		os.Exit(19)
	}
	fmt.Fprintln(os.Stdout, "check OK: declared environment reached gate")
	// Exercise separate OS writes and both output streams without logging values
	// in test failures. Lictor must scrub before mirroring or spooling any byte.
	fmt.Fprint(os.Stdout, "check OK: ")
	for i := range v {
		_, _ = os.Stdout.Write([]byte{v[i]})
	}
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stderr, "check OK: "+v)
	os.Exit(0)
}

func environmentFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "--allow-empty", "-qm", "fixture"}} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if err := c.Run(); err != nil {
			t.Fatal("fixture git setup failed")
		}
	}
	return dir
}

func environmentInvoke(t *testing.T, args []string) (int, []byte, []byte) {
	t.Helper()
	var out, diag bytes.Buffer
	if *environmentCLI == "" {
		code := run(context.Background(), args, &out, &diag, func() time.Time { return time.Unix(0, 0) })
		return code, out.Bytes(), diag.Bytes()
	}
	c := exec.Command(*environmentCLI, args...)
	c.Stdout, c.Stderr = &out, &diag
	err := c.Run()
	var ec *exec.ExitError
	if errors.As(err, &ec) {
		return ec.ExitCode(), out.Bytes(), diag.Bytes()
	}
	if err != nil {
		t.Fatal("fixture CLI could not start")
	}
	return 0, out.Bytes(), diag.Bytes()
}

func TestDeclaredEnvironment(t *testing.T) {
	t.Setenv("LICTOR_TEST_DATABASE_URL", environmentValue())
	t.Setenv("LICTOR_TEST_EMPTY", "")
	t.Setenv("LICTOR_TEST_UNDECLARED", "not inherited")
	t.Setenv("LICTOR_TEST_ABSENT", "")
	if err := os.Unsetenv("LICTOR_TEST_ABSENT"); err != nil {
		t.Fatal("cannot prepare absent variable")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal("test executable unavailable")
	}
	for _, mode := range []string{"human", "json", "dirty-all"} {
		t.Run(mode, func(t *testing.T) {
			dir := environmentFixture(t)
			args := []string{"witness", "run", "--repo", dir, "--unpinned", "--record", "GATE-RUN.txt", "--label", "environment", "--as-of", "2026-10-05T00:00:00Z", "--env", "LICTOR_TEST_DATABASE_URL", "--env", "LICTOR_TEST_ABSENT", "--env", "LICTOR_TEST_EMPTY"}
			if mode == "json" {
				args = append(args, "--json")
			}
			if mode == "dirty-all" {
				if os.WriteFile(filepath.Join(dir, "dirty"), []byte("dirty"), 0600) != nil {
					t.Fatal("cannot dirty fixture")
				}
				args = append(args, "--all")
			}
			args = append(args, "--", "probe='"+strings.ReplaceAll(exe, "'", "'\\''")+"' '-test.run=^TestDeclaredEnvironmentChild$' -- --lictor-env-child")
			code, out, diag := environmentInvoke(t, args)
			record, readErr := os.ReadFile(filepath.Join(dir, "GATE-RUN.txt"))
			if mode == "dirty-all" {
				if !os.IsNotExist(readErr) {
					t.Fatal("dirty run wrote requested record")
				}
				record = out
			} else if readErr != nil && code == 0 {
				t.Fatal("missing record")
			}
			for _, b := range [][]byte{out, diag, record} {
				if bytes.Contains(b, []byte(environmentValue())) {
					t.Fatal("declared value leaked in output bytes")
				}
			}
			if code != 0 || !bytes.Contains(record, []byte("check OK: declared environment reached gate")) {
				t.Fatalf("declared environment did not reach gate: exit=%d", code)
			}
			for _, line := range []string{"environment: LICTOR_TEST_DATABASE_URL present\n", "environment: LICTOR_TEST_ABSENT absent\n", "environment: LICTOR_TEST_EMPTY present (short value, not redacted)\n"} {
				if bytes.Count(record, []byte(line)) != 1 {
					t.Fatal("name/presence record missing or duplicated")
				}
			}
			t.Log("declared set reached gate; undeclared and absent names absent; all output bytes exclude credential-shaped value")
		})
	}
}

func TestDeclaredEnvironmentRefusals(t *testing.T) {
	for _, name := range []string{"PATH", "HOME", "GOFLAGS", "GOEXPERIMENT", "CGO_ENABLED", "LC_ALL", "CC", "GRYPE_DB_CACHE_DIR", "bad-name", "1NAME", "", "DUPLICATE"} {
		t.Run(name, func(t *testing.T) {
			dir := environmentFixture(t)
			args := []string{"witness", "run", "--repo", dir, "--unpinned", "--record", "GATE-RUN.txt", "--env", name}
			if name == "DUPLICATE" {
				args = append(args, "--env", name)
			}
			code, out, diag := environmentInvoke(t, append(args, "--", "probe=not-executed"))
			if code != 2 || !bytes.Contains(append(out, diag...), []byte(fmt.Sprintf("environment name %q", name))) {
				t.Fatal("missing named environment refusal")
			}
			if _, err := os.Stat(filepath.Join(dir, "GATE-RUN.txt")); !os.IsNotExist(err) {
				t.Fatal("refusal touched record")
			}
		})
	}
}

func TestDeclaredEnvironmentAssignmentNeverEchoesValue(t *testing.T) {
	dir := environmentFixture(t)
	code, out, diag := environmentInvoke(t, []string{"witness", "run", "--repo", dir, "--unpinned", "--record", "GATE-RUN.txt", "--env", "LICTOR_TEST_DATABASE_URL=" + environmentValue(), "--", "probe=not-executed"})
	b := append(out, diag...)
	if code != 2 || bytes.Contains(b, []byte(environmentValue())) || !bytes.Contains(b, []byte(`environment name "LICTOR_TEST_DATABASE_URL"`)) {
		t.Fatal("assignment refusal leaked a value or lost the name")
	}
}
