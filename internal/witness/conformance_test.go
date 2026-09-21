package witness

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var sourceScript = flag.String("witness-source-script", "", "explicit read-only source for conformance")
var allScript = flag.String("witness-all-script", "", "explicit read-only verify-all source for conformance")
var consumer = flag.String("witness-consumer", "", "explicit read-only OSS repository for conformance")

func sourceRun(t *testing.T, o Options, args ...string) {
	t.Helper()
	cmd := exec.Command("bash", append([]string{*sourceScript}, args...)...)
	cmd.Dir = o.Repo
	cmd.Env = append(os.Environ(), "GATE_WITNESS_CITES="+o.Cites, "GATE_WITNESS_TIE="+o.Tie, "GATE_WITNESS_XREPO="+strings.Join(o.Siblings, " "), "GATE_WITNESS_LOCK_WAIT=0", "GATE_WITNESS_ABORT_AFTER=")
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("source %v: %v\n%s", args, err, b)
	}
}

func TestSourceConformance(t *testing.T) {
	if *sourceScript == "" {
		t.Skip("explicit source conformance target only")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	if err = os.Symlink(exe, filepath.Join(bin, "date")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, tie := range []string{"digest", "commit"} {
		t.Run(tie, func(t *testing.T) {
			o := fixture(t)
			o.Tie = tie
			o.Cites = "one declared plan"
			o.Entries = []string{command(t, "tool-versions", "summary"), command(t, "test", "summary")}
			r := Run(context.Background(), o, io.Discard, io.Discard)
			if r.ExitCode != 0 {
				t.Fatalf("port: %+v", r)
			}
			port := record(t, o)
			sourceRun(t, o, append([]string{"run", o.Record, o.Label}, o.Entries...)...)
			if got := record(t, o); got != port {
				t.Fatalf("NONZERO RECORD DIFF\nSOURCE:\n%s\nPORT:\n%s", got, port)
			}
			t.Log("source/port byte diff empty: header, evidence, go packages, elapsed, exits, finalization")
		})
	}
	for _, all := range []bool{false, true} {
		t.Run(fmt.Sprintf("red-all-%t", all), func(t *testing.T) {
			if all && *allScript == "" {
				t.Skip("explicit verify-all source only")
			}
			o := fixture(t)
			o.All = all
			o.Entries = []string{command(t, "bad", "exit", "7"), command(t, "blocked", "summary"), command(t, "independent", "summary")}
			if all {
				o.Requires = []string{"blocked:bad"}
			}
			if r, _, _ := invoke(o); r.ExitCode != 1 {
				t.Fatal(r)
			}
			port := record(t, o)
			args := []string{*sourceScript, "run", o.Record, o.Label}
			if all {
				args = []string{*allScript, o.Record, o.Label, "--requires", "blocked:bad", "--"}
			}
			c := exec.Command("bash", append(args, o.Entries...)...)
			c.Dir = o.Repo
			c.Env = append(os.Environ(), "GATE_WITNESS_LOCK_WAIT=0", "GATE_WITNESS_TIE=digest", "GATE_WITNESS_CITES=", "GATE_WITNESS_XREPO=", "GATE_WITNESS_ABORT_AFTER=")
			b, err := c.CombinedOutput()
			var ec *exec.ExitError
			if !errors.As(err, &ec) || ec.ExitCode() != 1 {
				t.Fatalf("source red exit: %v %s", err, b)
			}
			if got := record(t, o); got != port {
				t.Fatalf("NONZERO RED DIFF\nSOURCE:\n%s\nPORT:\n%s", got, port)
			}
			t.Logf("red record source/port byte diff empty; all=%t", all)
		})
	}
	t.Run("stepwise", func(t *testing.T) {
		o := fixture(t)
		steps := []string{command(t, "tool-versions", "summary"), command(t, "test", "summary")}
		o.Mode = "init"
		o.Entries = []string{"tool-versions", "test"}
		if r, _, _ := invoke(o); r.ExitCode != 0 {
			t.Fatal(r)
		}
		for _, entry := range steps {
			o.Mode = "step"
			o.Entries = []string{entry}
			if r, _, _ := invoke(o); r.ExitCode != 0 {
				t.Fatal(r)
			}
		}
		o.Mode = "finalize"
		o.Entries = nil
		if r, _, _ := invoke(o); r.ExitCode != 0 {
			t.Fatal(r)
		}
		port := record(t, o)
		sourceRun(t, o, "init", o.Record, o.Label, "tool-versions", "test")
		for _, entry := range steps {
			sourceRun(t, o, "step", o.Record, entry)
		}
		sourceRun(t, o, "finalize", o.Record)
		if got := record(t, o); got != port {
			t.Fatalf("NONZERO STEPWISE DIFF\nSOURCE:\n%s\nPORT:\n%s", got, port)
		}
		t.Log("init/step/finalize source/port byte diff empty")
	})
	if *consumer != "" {
		t.Run("OSS-read-only", func(t *testing.T) {
			before := git(t, *consumer, "status", "--porcelain")
			head := git(t, *consumer, "rev-parse", "HEAD")
			o := fixture(t)
			o.Repo = *consumer
			o.Record = filepath.Join(t.TempDir(), "external-record.txt")
			o.Label = "OSS read-only recorder proof"
			o.Entries = []string{command(t, "first", "summary"), command(t, "second", "summary")}
			r := Run(context.Background(), o, io.Discard, io.Discard)
			if r.ExitCode != 0 {
				t.Fatalf("port: %+v", r)
			}
			port := record(t, o)
			sourceRun(t, o, append([]string{"run", o.Record, o.Label}, o.Entries...)...)
			if got := record(t, o); got != port {
				t.Fatalf("NONZERO OSS RECORD DIFF\nSOURCE:\n%s\nPORT:\n%s", got, port)
			}
			if after := git(t, *consumer, "status", "--porcelain"); before != after || head != git(t, *consumer, "rev-parse", "HEAD") {
				t.Fatalf("OSS changed: before %q after %q", before, after)
			}
			t.Logf("OSS head=%s porcelain before=after=%q; byte diff empty; external digest EMPTY-TREE is diagnostic only", strings.TrimSpace(head), before)
		})
	}
}

func TestOSSPlan(t *testing.T) {
	if *consumer == "" {
		t.Skip("explicit OSS plan measurement only")
	}
	b, err := os.ReadFile(filepath.Join(*consumer, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	_, body, ok := strings.Cut(string(b), "define VERIFY_PLAN\n")
	if !ok {
		t.Fatal("VERIFY_PLAN absent")
	}
	body, _, ok = strings.Cut(body, "\nendef")
	if !ok {
		t.Fatal("plan end absent")
	}
	var entries []string
	for _, line := range strings.Split(body, "\n") {
		s := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), "\\"))
		if s == "" {
			continue
		}
		words, e := Words(s)
		if e != nil || len(words) != 1 {
			t.Fatalf("outer plan quoting: %q %v", s, e)
		}
		entries = append(entries, strings.ReplaceAll(words[0], "$(MAKE)", "make"))
	}
	plan, _, err := Parse(entries, []string{"gograph-boundaries:gograph-build"}, false)
	if err != nil || len(plan) != 34 {
		t.Fatalf("OSS 34 entries: %d %v", len(plan), err)
	}
	makes := 0
	for _, e := range plan {
		if e.Argv[0] == "make" {
			makes++
		}
	}
	if makes != 28 {
		t.Fatalf("make entries=%d", makes)
	}
	entries[0] += " | cat"
	_, _, err = Parse(entries, nil, false)
	if err == nil || !strings.Contains(err.Error(), "tool-versions") || !strings.Contains(err.Error(), "'|'") {
		t.Fatalf("doctored pipe: %v", err)
	}
	t.Logf("34 accepted (28 make, six direct); doctored plan refused before execution: %v", err)
}

func TestModeAndFailureSemantics(t *testing.T) {
	o := fixture(t)
	o.Entries = []string{command(t, "a", "exit", "9"), command(t, "b", "summary"), command(t, "c", "summary")}
	o.Requires = []string{"b:a"}
	r, _, _ := invoke(o)
	s := record(t, o)
	if r.ExitCode != 1 || strings.Contains(s, "target: b") || !strings.Contains(s, "result: red") {
		t.Fatalf("fail-fast: %+v %s", r, s)
	}
	o.All = true
	r, _, _ = invoke(o)
	s = record(t, o)
	if r.ExitCode != 1 || !strings.Contains(s, "NOT-RUN b: dependency a recorded exit=9") || !strings.Contains(s, "target: b exit=125") || !strings.Contains(s, "target: c exit=0") {
		t.Fatalf("all-targets: %+v %s", r, s)
	}
	o.Mode = "init"
	o.Entries = []string{"a", "b"}
	o.Requires = nil
	r, _, _ = invoke(o)
	if r.ExitCode != 0 {
		t.Fatal(r)
	}
	o.Mode = "step"
	o.Entries = []string{command(t, "a", "exit", "7")}
	r, _, _ = invoke(o)
	if r.ExitCode != 7 {
		t.Fatalf("step exit: %+v", r)
	}
	o.Mode = "finalize"
	o.Entries = nil
	r, _, _ = invoke(o)
	if r.ExitCode != 1 || !strings.HasSuffix(record(t, o), "result: red\n") {
		t.Fatalf("incomplete finalized green: %+v", r)
	}
	_, sf, _ := paths(filepath.Join(o.Repo, o.Record))
	if _, err := os.Stat(sf); !os.IsNotExist(err) {
		t.Fatal("finalize retained session")
	}
}

func TestWordsAndInvalidPlans(t *testing.T) {
	w, err := Words(`a '' "b c" d\ e '|' "$HOME"`)
	if err != nil || strings.Join(w, "/") != "a//b c/d e/|/$HOME" {
		t.Fatalf("words: %q %v", w, err)
	}
	for _, s := range []string{`a 'broken`, `a "broken`, `a \`} {
		if _, err = Words(s); err == nil {
			t.Fatalf("invalid words accepted: %s", s)
		}
	}
	for _, p := range [][]string{nil, {"a"}, {"a="}, {"a=true", "a=false"}, {"bad name=true"}} {
		if _, _, err = Parse(p, nil, false); err == nil {
			t.Fatalf("invalid plan accepted: %v", p)
		}
	}
	if _, _, err = Parse([]string{"a=true"}, []string{"missing:a"}, false); err == nil || !strings.Contains(err.Error(), "unknown dependent missing") {
		t.Fatal(err)
	}
}

func TestMissingAndUnsafeRecord(t *testing.T) {
	o := fixture(t)
	o.Mode = "finalize"
	r, _, _ := invoke(o)
	if r.ExitCode != 2 {
		t.Fatal(r)
	}
	target := filepath.Join(o.Repo, "tracked")
	if err := os.Symlink(target, filepath.Join(o.Repo, o.Record)); err != nil {
		t.Fatal(err)
	}
	o.Mode = "run"
	o.Entries = []string{command(t, "a", "summary")}
	r, _, _ = invoke(o)
	if r.ExitCode != 2 || !strings.Contains(r.Reason, "regular") {
		t.Fatal(r)
	}
	b, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(b, []byte("original\n")) {
		t.Fatal("unsafe record wrote tracked data")
	}
}
