package witness

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStaleLockAndSiblingPins(t *testing.T) {
	o := fixture(t)
	sibling := fixture(t)
	o.Siblings = []string{"control=" + sibling.Repo}
	o.Entries = []string{command(t, "a", "summary")}
	lock, _, err := paths(filepath.Join(o.Repo, o.Record))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(lock, 0700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(lock) })
	if err = os.WriteFile(filepath.Join(lock, "owner"), []byte("pid=2147483647 label=dead started=fixed"), 0600); err != nil {
		t.Fatal(err)
	}
	r, _, diag := invoke(o)
	if r.ExitCode != 0 || !strings.Contains(diag, "breaking a STALE lock") {
		t.Fatalf("stale lock: %+v %s", r, diag)
	}
	head := strings.TrimSpace(git(t, sibling.Repo, "rev-parse", "--short", "HEAD"))
	if !strings.Contains(record(t, o), "xrepo: control head="+head+" tree=sha256:") {
		t.Fatal("missing explicit sibling pin")
	}
}

func TestTargetUnavailableOrTimedOutNeverFinalizesGreen(t *testing.T) {
	for _, timeout := range []bool{false, true} {
		t.Run(fmt.Sprint(timeout), func(t *testing.T) {
			o := fixture(t)
			o.Entries = []string{"missing=/nonexistent/lictor-fixture-command"}
			if timeout {
				o.Entries = []string{command(t, "slow", "hold", filepath.Join(t.TempDir(), "ready"))}
				o.Timeout = 10 * time.Millisecond
			}
			r := Run(context.Background(), o, io.Discard, io.Discard)
			if r.ExitCode != 2 || strings.Contains(record(t, o), "result: green") {
				t.Fatalf("unavailable target passed: %+v", r)
			}
		})
	}
}

func TestDirtyRedLeavesRecordAndRunsTargets(t *testing.T) {
	o := fixture(t)
	o.All = true
	marker := filepath.Join(t.TempDir(), "ran")
	o.Entries = []string{command(t, "bad", "exit", "8"), command(t, "independent", "touch", marker)}
	_ = os.WriteFile(filepath.Join(o.Repo, o.Record), []byte("prior"), 0600)
	_ = os.WriteFile(filepath.Join(o.Repo, "tracked"), []byte("dirty"), 0600)
	r, out, diag := invoke(o)
	if r.ExitCode != 1 || record(t, o) != "prior" || !strings.Contains(out, "GATE-WITNESS NOT MINTED: dirty work; GATE-RUN.txt is untouched") || !strings.Contains(diag, "NOT MINTING") {
		t.Fatalf("dirty red: %+v %s %s", r, out, diag)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("independent target did not run")
	}
}
