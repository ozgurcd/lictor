package witness

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ozgurcd/lictor/internal/executor"
)

const Ceiling int64 = 64 << 20

type Options struct {
	Repo, Record, Label, Mode, Cites, Tie string
	Entries, Requires, Siblings           []string
	All                                   bool
	LockWait, Timeout                     time.Duration
	Now                                   func() time.Time
	// TargetOutput overrides the source-compatible dirty --all target stream
	// for JSON callers, whose stdout must remain one result document.
	TargetOutput io.Writer
}

type Result struct {
	Schema     string `json:"schema_version"`
	Repository string `json:"repository"`
	Record     string `json:"record"`
	Reason     string `json:"reason"`
	ExitCode   int    `json:"exit_code"`
	Minted     bool   `json:"minted"`
}

func Run(ctx context.Context, o Options, out, diagnostic io.Writer) Result {
	r := Result{Schema: "lictor.witness.v1", Repository: o.Repo, Record: o.Record}
	fail := func(code int, err error) Result { r.ExitCode = code; r.Reason = err.Error(); return r }
	if o.Mode == "" {
		o.Mode = "run"
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Timeout == 0 {
		o.Timeout = 30 * time.Minute
	}
	if o.Record == "" || !filepath.IsAbs(o.Repo) || o.LockWait < 0 || o.Timeout < 0 {
		return fail(2, fmt.Errorf("witness requires --repo ABS, --record and nonnegative limits"))
	}
	if strings.ContainsAny(o.Label+o.Cites, "\n\r") {
		return fail(2, fmt.Errorf("label and cites must be single lines"))
	}
	if !utf8.ValidString(o.Label) || !utf8.ValidString(o.Cites) || strings.ContainsRune(o.Label+o.Cites, 0) {
		return fail(2, fmt.Errorf("label and cites must be UTF-8 text without NUL"))
	}
	if o.Tie != "" && o.Tie != "digest" && o.Tie != "commit" {
		return fail(2, fmt.Errorf("unknown tie %q", o.Tie))
	}
	if o.Mode != "run" && o.Mode != "init" && o.Mode != "step" && o.Mode != "finalize" {
		return fail(2, fmt.Errorf("unsupported witness mode %s", o.Mode))
	}
	var entries []Entry
	var deps map[string][]string
	var err error
	if o.Mode == "finalize" {
		if len(o.Entries) != 0 || len(o.Requires) != 0 {
			return fail(2, fmt.Errorf("finalize takes no plan"))
		}
	} else {
		entries, deps, err = Parse(o.Entries, o.Requires, o.Mode == "init")
		if err != nil {
			return fail(2, err)
		}
		if o.Mode == "step" && (len(entries) != 1 || len(o.Requires) != 0) {
			return fail(2, fmt.Errorf("step takes exactly one target and no dependencies"))
		}
	}
	for _, s := range o.Siblings {
		n, p, ok := strings.Cut(s, "=")
		if !ok || !namePattern.MatchString(n) || !filepath.IsAbs(p) {
			return fail(2, fmt.Errorf("sibling must be NAME=ABS: %s", s))
		}
	}
	path := o.Record
	if !filepath.IsAbs(path) {
		path = filepath.Join(o.Repo, path)
	}
	if st, e := os.Lstat(path); e == nil && !st.Mode().IsRegular() {
		return fail(2, fmt.Errorf("record must be a regular file: %s", path))
	} else if e != nil && !os.IsNotExist(e) {
		return fail(2, e)
	}
	lock, sf, err := paths(path)
	if err != nil {
		return fail(2, err)
	}
	if o.Mode == "run" || o.Mode == "init" {
		what := "a one-shot run"
		if o.Mode == "init" {
			what = "a second stepwise session"
		}
		if code, e := sessionGuard(sf, o.Record, what, diagnostic); e != nil {
			return fail(code, e)
		}
	}
	unlock, code, err := acquire(ctx, lock, o.Record, o.Label, o.LockWait, o.Now, diagnostic)
	if err != nil {
		return fail(code, err)
	}
	defer unlock()
	// Recheck under the lock: another init may have opened while we waited.
	if o.Mode == "run" || o.Mode == "init" {
		what := "a one-shot run"
		if o.Mode == "init" {
			what = "a second stepwise session"
		}
		if code, e := sessionGuard(sf, o.Record, what, diagnostic); e != nil {
			return fail(code, e)
		}
	}
	dirty := false
	original := o.Record
	var workState string
	if o.Mode == "run" {
		workState, err = state(ctx, o.Repo, o.Record, true)
		if err != nil {
			return fail(2, err)
		}
		dirty = strings.HasSuffix(workState, " (dirty)")
	}
	if dirty {
		f, e := os.CreateTemp("", "gate-witness-unminted.*")
		if e != nil {
			return fail(2, e)
		}
		path = f.Name()
		_ = f.Close()
		defer os.Remove(path)
		o.Record = path
		if o.All {
			fmt.Fprintf(diagnostic, "GATE-WITNESS NOT MINTING: dirty work; %s remains untouched\n", original)
		} else {
			fmt.Fprintf(diagnostic, "gate-witness: NOT MINTING — the tree (%s) is dirty beyond the gate records; every target runs and the verdict is printed, but no record is written and %s stays as it is\n", workState, original)
		}
	}
	flags := os.O_WRONLY | os.O_APPEND
	if o.Mode == "run" || o.Mode == "init" {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	f, err := os.OpenFile(path, flags, 0600)
	if err != nil {
		return fail(2, fmt.Errorf("gate-witness: record %s: %w", original, err))
	}
	defer f.Close()
	if o.Mode == "run" || o.Mode == "init" {
		if err = header(ctx, o, entries, f); err != nil {
			return fail(2, err)
		}
	}
	if o.Mode == "init" {
		err = os.WriteFile(sf, []byte(fmt.Sprintf("owner=%d label=%s started=%s\n", os.Getppid(), o.Label, stamp(o.Now()))), 0600)
		if err != nil {
			return fail(2, err)
		}
		r.Minted = true
		r.Reason = "gate-witness: initialized " + original
		return r
	}
	overall := 0
	targetOutput := diagnostic
	if dirty && o.All {
		targetOutput = out
		if o.TargetOutput != nil {
			targetOutput = o.TargetOutput
		}
	}
	completed := map[string]int{}
	for _, e := range entries {
		if ctx.Err() != nil {
			return fail(2, ctx.Err())
		}
		blocked := ""
		for _, p := range deps[e.Name] {
			ec, ok := completed[p]
			if !ok || ec != 0 {
				status := "missing"
				if ok {
					status = fmt.Sprint(ec)
				}
				blocked = "dependency " + p + " recorded exit=" + status
				break
			}
		}
		var ec int
		if blocked != "" {
			ec = 125
			if dirty && o.All {
				fmt.Fprintf(targetOutput, "==> gate-witness: %s\ncheck FAILED: NOT-RUN %s: %s\n", e.Name, e.Name, blocked)
			}
			_, err = fmt.Fprintf(f, "evidence: [%s] check FAILED: NOT-RUN %s: %s\nelapsed: %s 0s\ntarget: %s exit=125\n", e.Name, e.Name, blocked, e.Name, e.Name)
		} else {
			ec, err = one(ctx, o, e, f, targetOutput)
		}
		if err != nil {
			return fail(2, err)
		}
		completed[e.Name] = ec
		if ec != 0 {
			overall = 1
		}
		if o.Mode == "step" {
			r.ExitCode = ec
			r.Minted = true
			r.Reason = fmt.Sprintf("target: %s exit=%d", e.Name, ec)
			return r
		}
		if overall != 0 && !o.All {
			break
		}
	}
	verdict, err := finalize(ctx, o, path, f)
	if err != nil {
		return fail(2, err)
	}
	if o.Mode == "finalize" {
		if err = os.Remove(sf); err != nil && !os.IsNotExist(err) {
			return fail(2, err)
		}
	}
	if verdict != "green" {
		overall = 1
	}
	r.ExitCode = overall
	r.Minted = !dirty
	r.Reason = "gate-witness: " + original + " result: " + verdict
	if dirty && o.All {
		// Reopen the completed scratch record for a streaming, exact-byte echo.
		// The requested record has never been opened for writing.
		echo, e := os.Open(path)
		if e != nil {
			return fail(2, e)
		}
		_, e = io.Copy(out, echo)
		closeErr := echo.Close()
		if e != nil {
			return fail(2, e)
		}
		if closeErr != nil {
			return fail(2, closeErr)
		}
		r.Reason = fmt.Sprintf("GATE-WITNESS NOT MINTED: dirty work; %s is untouched", original)
	} else if dirty {
		remedy := "Commit the work, then run this gate again at the clean HEAD to mint the record and witness it."
		if overall != 0 {
			remedy = "Fix the red target(s) above, commit, and run again at the clean HEAD."
		}
		r.Reason = fmt.Sprintf("GATE-WITNESS NOT MINTED: '%s' is %s on a DIRTY tree (%s) — no record was written and %s is untouched. %s", o.Label, verdict, workState, original, remedy)
	}
	if _, err = fmt.Fprintln(out, r.Reason); err != nil {
		return fail(2, err)
	}
	return r
}

func stamp(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05Z") }
func header(ctx context.Context, o Options, entries []Entry, w io.Writer) error {
	s, err := state(ctx, o.Repo, o.Record, true)
	if err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "schema: gate-run.v1\ngate: %s\nnote: evidence lines are the tools' own summary lines; a summary format this script does not match is recorded only as an exit code\n", o.Label)
	if o.Cites != "" {
		fmt.Fprintf(&b, "cites: %s\n", o.Cites)
	}
	fmt.Fprintf(&b, "repo-head: %s\nstarted: %s\nplan:", s, stamp(o.Now()))
	for _, e := range entries {
		fmt.Fprintf(&b, " %s", e.Name)
	}
	b.WriteByte('\n')
	_, err = io.WriteString(w, b.String())
	return err
}

var evidence = regexp.MustCompile(`^check OK:|^check FAILED:|Tests  [0-9]|Test Files |wiki freshness:|sync violations:|SELFTEST OK|gate-witness OK:`)
var packages = regexp.MustCompile(`^ok[\t\n\v\f\r ]+\S+[\t\n\v\f\r ]+([0-9.]+s|\(cached\))`)

func one(ctx context.Context, o Options, e Entry, w, diag io.Writer) (int, error) {
	f, err := os.CreateTemp("", "gate-witness-out.*")
	if err != nil {
		return 2, err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	fmt.Fprintf(diag, "==> gate-witness: %s\n", e.Name)
	start := o.Now().Unix()
	ec, total, err := executor.RecordOutput(ctx, o.Repo, e.Argv, o.Timeout, Ceiling, f, diag)
	elapsed := o.Now().Unix() - start
	if err != nil {
		return 2, fmt.Errorf("target %s: %w", e.Name, err)
	}
	if _, err = f.Seek(0, 0); err != nil {
		return 2, err
	}
	// Inspect all captured bytes before copying any line: a late NUL or
	// invalid UTF-8 byte makes the entire target's output non-text.
	kind, err := outputKind(f)
	if err != nil {
		return 2, err
	}
	if kind != "" {
		if _, err = fmt.Fprintf(w, "binary-output: %s contains %s; evidence omitted\n", e.Name, kind); err != nil {
			return 2, err
		}
	} else {
		if _, err = f.Seek(0, 0); err != nil {
			return 2, err
		}
		// The source writes all tool lines before all evidence lines.
		if e.Name == "tool-versions" {
			if err = lines(f, func(s string) error { _, e := fmt.Fprintf(w, "tool: %s\n", s); return e }); err != nil {
				return 2, err
			}
			if _, err = f.Seek(0, 0); err != nil {
				return 2, err
			}
		}
		count := 0
		err = lines(f, func(s string) error {
			if packages.MatchString(s) {
				count++
			}
			if evidence.MatchString(s) {
				_, err := fmt.Fprintf(w, "evidence: [%s] %s\n", e.Name, s)
				return err
			}
			return nil
		})
		if err != nil {
			return 2, err
		}
		if count > 0 {
			if _, err = fmt.Fprintf(w, "evidence: [%s] go packages ok: %d\n", e.Name, count); err != nil {
				return 2, err
			}
		}
	}
	if total > Ceiling {
		if _, err = fmt.Fprintf(w, "truncated: %s output exceeded %d bytes; retained=%d total=%d\n", e.Name, Ceiling, Ceiling, total); err != nil {
			return 2, err
		}
	}
	_, err = fmt.Fprintf(w, "elapsed: %s %ds\ntarget: %s exit=%d\n", e.Name, elapsed, e.Name, ec)
	return ec, err
}

func outputKind(r io.Reader) (string, error) {
	kind := ""
	err := lines(r, func(s string) error {
		if strings.ContainsRune(s, 0) {
			kind = "NUL"
		} else if kind == "" && !utf8.ValidString(s) {
			kind = "invalid UTF-8"
		}
		return nil
	})
	return kind, err
}

func lines(r io.Reader, fn func(string) error) error {
	b := bufio.NewScanner(r)
	b.Buffer(make([]byte, 64<<10), int(Ceiling)+(1<<20))
	b.Split(func(data []byte, eof bool) (int, []byte, error) {
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			return i + 1, data[:i], nil
		}
		if eof && len(data) > 0 {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	for b.Scan() {
		if err := fn(b.Text()); err != nil {
			return err
		}
	}
	return b.Err()
}

func finalize(ctx context.Context, o Options, path string, w io.Writer) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var plan []string
	seenPlan := false
	passed := map[string]bool{}
	err = lines(f, func(s string) error {
		if strings.HasPrefix(s, "plan: ") && !seenPlan {
			plan = strings.Fields(strings.TrimPrefix(s, "plan: "))
			seenPlan = true
		}
		if strings.HasPrefix(s, "target: ") && strings.HasSuffix(s, " exit=0") {
			passed[strings.TrimSuffix(strings.TrimPrefix(s, "target: "), " exit=0")] = true
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	verdict := "green"
	for _, n := range plan {
		if !passed[n] {
			verdict = "red"
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "finished: %s\n", stamp(o.Now()))
	if verdict == "green" {
		if o.Tie == "commit" {
			head, e := gitRead(ctx, o.Repo, "", "rev-parse", "HEAD")
			if e != nil {
				return "", e
			}
			dirty, e := gitRead(ctx, o.Repo, "", "status", "--porcelain")
			if e != nil {
				return "", e
			}
			if dirty != "" {
				head += " (dirty-at-finalize)"
			}
			fmt.Fprintln(&b, "tie-note: CI cannot commit this record, so the committed-alongside witness chain does not exist here; the checkout IS an unmodified commit, so the full commit SHA below pins the tree content exactly as the local digest does — valid only while the tree stays clean, which check enforces")
			fmt.Fprintf(&b, "tree: commit=%s\n", head)
		} else {
			d, e := digest(ctx, o.Repo, o.Record)
			if e != nil {
				return "", e
			}
			fmt.Fprintf(&b, "tree: sha256=%s\n", d)
		}
		for _, s := range o.Siblings {
			n, p, _ := strings.Cut(s, "=")
			head, e := state(ctx, p, "GATE-RUN.txt", false)
			if e != nil {
				return "", e
			}
			d, e := digest(ctx, p, ".gate-witness-nonexistent")
			if e != nil {
				return "", e
			}
			fmt.Fprintf(&b, "xrepo: %s head=%s tree=sha256:%s\n", n, head, d)
		}
	}
	fmt.Fprintf(&b, "result: %s\n", verdict)
	_, err = io.WriteString(w, b.String())
	return verdict, err
}
