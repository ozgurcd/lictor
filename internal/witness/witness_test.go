package witness

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

var fixed = time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)

func TestMain(m *testing.M) {
	if filepath.Base(os.Args[0]) == "date" {
		if os.Args[1] == "+%s" {
			fmt.Println(fixed.Unix())
		} else {
			fmt.Println(fixed.Format("2006-01-02T15:04:05Z"))
		}
		os.Exit(0)
	}
	if len(os.Args) > 1 && os.Args[1] == "--fixture" {
		switch os.Args[2] {
		case "summary":
			fmt.Print("ordinary noise\ncheck OK: one\ncheck FAILED: two\n Tests  3 passed\nTest Files  1 passed\nwiki freshness: fresh\nsync violations: 0\nSELFTEST OK\ngate-witness OK: fixture\nok  example/pkg 0.01s\nok example/cached (cached)\nok unrelated summary\n")
		case "bytes":
			n, _ := strconv.Atoi(os.Args[3])
			fmt.Print("check OK: ")
			chunk := bytes.Repeat([]byte("x"), 65536)
			for n > 0 {
				k := min(n, len(chunk))
				_, _ = os.Stdout.Write(chunk[:k])
				n -= k
			}
			fmt.Println()
			if len(os.Args) > 4 {
				ec, _ := strconv.Atoi(os.Args[4])
				os.Exit(ec)
			}
		case "hold":
			_ = os.WriteFile(os.Args[3], []byte("ready"), 0600)
			if len(os.Args) > 4 {
				for {
					if _, err := os.Stat(os.Args[4]); err == nil {
						break
					}
					time.Sleep(time.Millisecond)
				}
			} else {
				time.Sleep(2 * time.Second)
			}
		case "exit":
			n, _ := strconv.Atoi(os.Args[3])
			os.Exit(n)
		case "touch":
			_ = os.WriteFile(os.Args[3], []byte("ran"), 0600)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=Fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=Fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
	b, e := c.CombinedOutput()
	if e != nil {
		t.Fatalf("git %v: %s: %v", args, b, e)
	}
	return string(b)
}

func fixture(t *testing.T) Options {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	if err := os.WriteFile(filepath.Join(dir, "tracked"), []byte("original\n"), 0600); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "--", "tracked")
	git(t, dir, "commit", "-qm", "fixture")
	return Options{Repo: dir, Record: "GATE-RUN.txt", Label: "fixture", Mode: "run", Timeout: time.Minute, Now: func() time.Time { return fixed }}
}

func command(t *testing.T, name, kind string, args ...string) string {
	t.Helper()
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	words := append([]string{exe, "--fixture", kind}, args...)
	for i, w := range words {
		words[i] = "'" + strings.ReplaceAll(w, "'", "'\\''") + "'"
	}
	return name + "=" + strings.Join(words, " ")
}

func invoke(o Options) (Result, string, string) {
	var out, diag bytes.Buffer
	r := Run(context.Background(), o, &out, &diag)
	return r, out.String(), diag.String()
}
func record(t *testing.T, o Options) string {
	t.Helper()
	p := o.Record
	if !filepath.IsAbs(p) {
		p = filepath.Join(o.Repo, p)
	}
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}

func TestPlanRefusesBeforeExecution(t *testing.T) {
	for _, bad := range []string{"x=echo a | cat", "x=echo a &", "x=echo a; true", "x=echo > a", "x=echo < a", "x=echo $HOME", "x=echo `id`", "x=echo (a)", "x=echo a\nb"} {
		o := fixture(t)
		marker := filepath.Join(o.Repo, "marker")
		o.Entries = []string{command(t, "first", "touch", marker), bad}
		r, _, _ := invoke(o)
		if r.ExitCode != 2 || !strings.Contains(r.Reason, "x") {
			t.Fatalf("missing named refusal: %+v", r)
		}
		if _, e := os.Stat(marker); !os.IsNotExist(e) {
			t.Fatal("target ran before plan was validated")
		}
		if _, e := os.Stat(filepath.Join(o.Repo, o.Record)); !os.IsNotExist(e) {
			t.Fatal("invalid plan wrote record")
		}
	}
	w, e := Words(`gograph capabilities --intention "repo-local make verify"`)
	if e != nil || len(w) != 4 || w[3] != "repo-local make verify" {
		t.Fatalf("POSIX words: %q %v", w, e)
	}
	for _, req := range []string{"b:missing", "a:b"} {
		_, _, e := Parse([]string{"a=true", "b=true"}, []string{req}, false)
		if e == nil || !strings.Contains(e.Error(), strings.Split(req, ":")[1]) {
			t.Fatalf("dependency accepted: %s %v", req, e)
		}
	}
}

// RULE: WITNESS-ONE-WRITER-1
func TestRule_WITNESS_ONE_WRITER_1(t *testing.T) {
	o := fixture(t)
	o.Entries = []string{command(t, "a", "summary")}
	if e := os.WriteFile(filepath.Join(o.Repo, o.Record), []byte("gate: old\ngate: older\n"), 0600); e != nil {
		t.Fatal(e)
	}
	r, _, _ := invoke(o)
	if r.ExitCode != 0 {
		t.Fatalf("run: %+v", r)
	}
	if got := record(t, o); strings.Count(got, "gate: ") != 1 || strings.Contains(got, "older") {
		t.Fatalf("open did not truncate: %s", got)
	}
	lock, session, e := paths(filepath.Join(o.Repo, o.Record))
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(lock, 0700); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _ = os.RemoveAll(lock); _ = os.Remove(session) })
	if e = os.WriteFile(filepath.Join(lock, "owner"), []byte(fmt.Sprintf("pid=%d label=fixture started=fixed", os.Getpid())), 0600); e != nil {
		t.Fatal(e)
	}
	before := record(t, o)
	r, _, diag := invoke(o)
	if r.ExitCode != 3 || !strings.Contains(diag, "waited 0s and refusing") || record(t, o) != before {
		t.Fatalf("live lock refusal: %+v %s", r, diag)
	}
	_ = os.RemoveAll(lock)
	o.Mode = "init"
	o.Entries = []string{"a"}
	r, _, _ = invoke(o)
	if r.ExitCode != 0 {
		t.Fatalf("init: %+v", r)
	}
	before = record(t, o)
	r, _, diag = invoke(o)
	if r.ExitCode != 4 || !strings.Contains(diag, "SESSION OPEN") || record(t, o) != before {
		t.Fatalf("session refusal: %+v %s", r, diag)
	}
}

func TestDirtyAndStaleSession(t *testing.T) {
	o := fixture(t)
	o.Entries = []string{command(t, "a", "summary")}
	p := filepath.Join(o.Repo, o.Record)
	_ = os.WriteFile(p, []byte("prior"), 0600)
	_ = os.WriteFile(filepath.Join(o.Repo, "tracked"), []byte("dirty"), 0600)
	r, out, diag := invoke(o)
	if r.ExitCode != 0 || r.Minted || record(t, o) != "prior" || !strings.Contains(diag, "NOT MINTING") || !strings.Contains(out, "green") {
		t.Fatalf("dirty run: %+v %s %s", r, out, diag)
	}
	_ = os.WriteFile(filepath.Join(o.Repo, "tracked"), []byte("original\n"), 0600)
	_, sf, e := paths(p)
	if e != nil {
		t.Fatal(e)
	}
	_ = os.WriteFile(sf, []byte("owner=unknown"), 0600)
	t.Cleanup(func() { _ = os.Remove(sf) })
	r, _, diag = invoke(o)
	if r.ExitCode != 0 || !strings.Contains(diag, "breaking a STALE session") {
		t.Fatalf("stale session: %+v %s", r, diag)
	}
}

func TestRecorderCeiling(t *testing.T) {
	for _, n := range []int{9 << 20, 65 << 20} {
		o := fixture(t)
		exit := "0"
		want := 0
		if n > int(Ceiling) {
			exit = "7"
			want = 1
		}
		o.Entries = []string{command(t, "a", "bytes", strconv.Itoa(n), exit)}
		r := Run(context.Background(), o, io.Discard, io.Discard)
		s := record(t, o)
		if r.ExitCode != want || !strings.Contains(s, "target: a exit="+exit) {
			t.Fatalf("real exit lost: %+v", r)
		}
		if n < int(Ceiling) {
			if !strings.Contains(s, "evidence: [a] check OK: "+strings.Repeat("x", n)+"\n") || strings.Contains(s, "truncated:") {
				t.Fatalf("9 MiB incomplete")
			}
		} else if !strings.Contains(s, "truncated: a") {
			t.Fatal("missing truncation evidence")
		}
		t.Logf("output=%d bytes plus summary prefix; target exit=%s preserved; truncated=%t", n, exit, strings.Contains(s, "truncated:"))
	}
}

func TestConcurrentRunBoundedRefusal(t *testing.T) {
	o := fixture(t)
	ready := filepath.Join(t.TempDir(), "ready")
	release := filepath.Join(filepath.Dir(ready), "release")
	o.Entries = []string{command(t, "holder", "hold", ready, release)}
	done := make(chan Result, 1)
	go func() { done <- Run(context.Background(), o, io.Discard, io.Discard) }()
	defer func() {
		_ = os.WriteFile(release, []byte("release"), 0600)
		if r := <-done; r.ExitCode != 0 {
			t.Errorf("holder failed: %+v", r)
		}
	}()
	deadline := time.After(10 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("holder did not start")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	before := record(t, o)
	o.LockWait = time.Second
	r, _, diag := invoke(o)
	if r.ExitCode != 3 || !strings.Contains(diag, "waited 1s and refusing") || record(t, o) != before {
		t.Fatalf("concurrent writer not refused: %+v %s", r, diag)
	}
	t.Log("live concurrent writer: bound 1s, exit 3, record untouched; default CLI bound remains 120s")
}
