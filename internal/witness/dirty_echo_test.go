package witness

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var witnessCLI = flag.String("witness-cli", "", "explicit built CLI for full-stream conformance")
var witnessBeforeCLI = flag.String("witness-before-cli", "", "same-version pre-change CLI for run compatibility")

func dirtyEchoFixture(t *testing.T) Options {
	t.Helper()
	o := fixture(t)
	if err := os.WriteFile(filepath.Join(o.Repo, o.Record), []byte("last committed record\n"), 0600); err != nil {
		t.Fatal(err)
	}
	git(t, o.Repo, "add", "--", o.Record)
	git(t, o.Repo, "commit", "-qm", "record")
	if err := os.WriteFile(filepath.Join(o.Repo, "tracked"), []byte("dirty\n"), 0600); err != nil {
		t.Fatal(err)
	}
	o.All = true
	return o
}

func echoPlan(t *testing.T, o *Options, kind string) int {
	t.Helper()
	o.Entries = []string{command(t, "first", "summary"), command(t, "middle", "summary"), command(t, "last", "summary")}
	if kind == "green" {
		return 0
	}
	o.Entries[0] = command(t, "first", "exit", "7")
	if kind == "not-run" {
		o.Requires = []string{"middle:first"}
	}
	return 1
}

// RULE: WITNESS-DIRTY-ECHO-1
func TestRule_WITNESS_DIRTY_ECHO_1(t *testing.T) {
	for _, kind := range []string{"green", "red", "not-run"} {
		t.Run(kind, func(t *testing.T) {
			o := dirtyEchoFixture(t)
			want := echoPlan(t, &o, kind)
			before := sha256.Sum256([]byte(record(t, o)))
			r, out, diag := invoke(o)
			if r.ExitCode != want || r.Minted {
				t.Fatalf("dirty verdict: %+v", r)
			}
			if !strings.Contains(out, "GATE-WITNESS NOT MINTED: dirty work; GATE-RUN.txt is untouched\n") {
				t.Errorf("NOT MINTED absent: %s", out)
			}
			if n := strings.Count(out, "\ntarget: "); n != len(o.Entries) {
				t.Errorf("target lines=%d want=%d", n, len(o.Entries))
			}
			if !strings.Contains(out, "target: last exit=0\n") {
				t.Errorf("last target absent: %s", out)
			}
			if after := sha256.Sum256([]byte(record(t, o))); before != after {
				t.Errorf("committed record changed: before=%x after=%x", before, after)
			}
			if diag != "GATE-WITNESS NOT MINTING: dirty work; GATE-RUN.txt remains untouched\n" {
				t.Errorf("stderr=%q", diag)
			}
			if kind == "not-run" && !strings.Contains(out, "target: middle exit=125\n") {
				t.Errorf("missing NOT-RUN 125: %s", out)
			}
			if !t.Failed() {
				t.Logf("NOT MINTED present; exactly 3 target lines; last target present; committed record unchanged sha256=%x", before)
			}
		})
	}
}

func echoProcess(t *testing.T, dir, executable string, args ...string) (string, string, int) {
	t.Helper()
	c := exec.Command(executable, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GATE_WITNESS_LOCK_WAIT=0", "GATE_WITNESS_TIE=digest", "GATE_WITNESS_CITES=", "GATE_WITNESS_XREPO=", "GATE_WITNESS_ABORT_AFTER=")
	var out, diag bytes.Buffer
	c.Stdout, c.Stderr = &out, &diag
	err := c.Run()
	var ec *exec.ExitError
	if errors.As(err, &ec) {
		return out.String(), diag.String(), ec.ExitCode()
	}
	if err != nil {
		t.Fatal(err)
	}
	return out.String(), diag.String(), 0
}

func echoCLIArgs(o Options) []string {
	args := []string{"witness", "--repo", o.Repo, "--record", o.Record, "--label", o.Label, "--as-of", fixed.Format("2006-01-02T15:04:05Z")}
	if o.All {
		args = append(args, "--all")
	}
	for _, dependency := range o.Requires {
		args = append(args, "--requires", dependency)
	}
	return append(append(args, "--"), o.Entries...)
}

func echoCLIPin(t *testing.T, o Options) {
	t.Helper()
	out, diag, code := echoProcess(t, o.Repo, *witnessCLI, "version", "--json")
	var v struct{ Version string }
	if code != 0 || json.Unmarshal([]byte(out), &v) != nil || v.Version == "" {
		t.Fatalf("CLI version: %d %s %s", code, out, diag)
	}
	dir := filepath.Join(o.Repo, ".github", "workflows")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ci.yml"), []byte("env:\n  LICTOR_VERSION: "+v.Version+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	git(t, o.Repo, "add", "--", ".github/workflows/ci.yml")
	git(t, o.Repo, "commit", "-qm", "pin")
}

func TestDirtyEchoSourceStreams(t *testing.T) {
	if *allScript == "" || *witnessCLI == "" {
		t.Skip("explicit source and built CLI required")
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
	for _, kind := range []string{"green", "red", "not-run"} {
		t.Run(kind, func(t *testing.T) {
			o := dirtyEchoFixture(t)
			echoCLIPin(t, o)
			want := echoPlan(t, &o, kind)
			before := sha256.Sum256([]byte(record(t, o)))
			args := []string{*allScript, o.Record, o.Label}
			for _, dependency := range o.Requires {
				args = append(args, "--requires", dependency)
			}
			args = append(append(args, "--"), o.Entries...)
			sout, serr, scode := echoProcess(t, o.Repo, "bash", args...)
			afterSource := sha256.Sum256([]byte(record(t, o)))
			pout, perr, pcode := echoProcess(t, o.Repo, *witnessCLI, echoCLIArgs(o)...)
			afterPort := sha256.Sum256([]byte(record(t, o)))
			if sout != pout {
				t.Errorf("stdout diff nonempty\nSOURCE:\n%s\nPORT:\n%s", sout, pout)
			}
			if serr != perr {
				t.Errorf("stderr diff nonempty\nSOURCE: %q\nPORT: %q", serr, perr)
			}
			if scode != want || pcode != scode {
				t.Errorf("exit source=%d port=%d want=%d", scode, pcode, want)
			}
			if before != afterSource || before != afterPort {
				t.Errorf("record hashes before=%x source=%x port=%x", before, afterSource, afterPort)
			}
			if !t.Failed() {
				t.Log("stdout diff: empty (unfiltered full streams)")
				t.Log("stderr diff: empty (unfiltered full streams)")
				t.Logf("exit: source=%d lictor=%d", scode, pcode)
				t.Logf("record sha256: before=%x after-source=%x after-lictor=%x", before, afterSource, afterPort)
			}
		})
	}
}

func TestRunBytesUnchanged(t *testing.T) {
	if *witnessCLI == "" || *witnessBeforeCLI == "" {
		t.Skip("explicit before/after CLIs of the same version required")
	}
	for _, dirty := range []bool{false, true} {
		t.Run(fmt.Sprintf("dirty-%t", dirty), func(t *testing.T) {
			o := fixture(t)
			echoCLIPin(t, o)
			o.Entries = []string{command(t, "first", "summary"), command(t, "middle", "exit", "7"), command(t, "last", "summary")}
			if dirty {
				if err := os.WriteFile(filepath.Join(o.Repo, "tracked"), []byte("dirty\n"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(o.Repo, o.Record), []byte("retained\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			beforeOut, beforeErr, beforeCode := echoProcess(t, o.Repo, *witnessBeforeCLI, echoCLIArgs(o)...)
			beforeRecord := record(t, o)
			afterOut, afterErr, afterCode := echoProcess(t, o.Repo, *witnessCLI, echoCLIArgs(o)...)
			if beforeOut != afterOut || beforeErr != afterErr || beforeCode != afterCode || beforeRecord != record(t, o) {
				t.Fatalf("run changed: exits=%d/%d stdout=%t stderr=%t record=%t", beforeCode, afterCode, beforeOut == afterOut, beforeErr == afterErr, beforeRecord == record(t, o))
			}
			t.Logf("run dirty=%t: stdout, stderr, record byte-identical; exit before=after=%d", dirty, beforeCode)
		})
	}
}

func TestDirtyEchoJSONTargetStream(t *testing.T) {
	o := dirtyEchoFixture(t)
	echoPlan(t, &o, "green")
	var diag bytes.Buffer
	o.TargetOutput = &diag
	r := Run(context.Background(), o, &bytes.Buffer{}, &diag)
	if r.ExitCode != 0 || !strings.Contains(diag.String(), "ordinary noise") {
		t.Fatalf("JSON target diagnostic lost: %+v %s", r, diag.String())
	}
}
