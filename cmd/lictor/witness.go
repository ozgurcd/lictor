package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ozgurcd/lictor/internal/witness"
)

type many []string

func (m *many) String() string     { return fmt.Sprint([]string(*m)) }
func (m *many) Set(s string) error { *m = append(*m, s); return nil }

func runWitness(ctx context.Context, args []string, wantJSON bool, out, diag io.Writer) int {
	o := witness.Options{Mode: "run"}
	if len(args) > 0 && len(args[0]) > 0 && args[0][0] != '-' {
		o.Mode = args[0]
		args = args[1:]
	}
	o.Repo, _ = os.Getwd()
	fs := flag.NewFlagSet("lictor witness", flag.ContinueOnError)
	var help bytes.Buffer
	fs.SetOutput(&help)
	fs.StringVar(&o.Repo, "repo", o.Repo, "absolute repository; default current working directory")
	fs.StringVar(&o.Record, "record", "", "record path; relative to repository")
	fs.StringVar(&o.Label, "label", "", "gate label")
	fs.StringVar(&o.Cites, "cites", "", "explicit declaration citation")
	fs.StringVar(&o.Tie, "tie", "digest", "digest or commit, matching source CI tie")
	var requires, siblings many
	fs.Var(&requires, "requires", "target:earlier-prerequisite; repeatable")
	fs.Var(&siblings, "sibling", "NAME=absolute-path to include at finalization; repeatable")
	fs.BoolVar(&o.All, "all", false, "attempt all independent targets, as verify-all; default run stops on first failure")
	fs.DurationVar(&o.LockWait, "lock-wait", 120*time.Second, "bounded per-record wait; source default 120s; refusal exit 3")
	fs.DurationVar(&o.Timeout, "timeout", 30*time.Minute, "per-target execution bound")
	var asOf string
	fs.StringVar(&asOf, "as-of", "", "explicit RFC3339 clock, freezes record timestamps and elapsed seconds for replay")
	fs.BoolVar(&wantJSON, "json", wantJSON, "one lictor.witness.v1 result; target output goes to stderr")
	unpinned := fs.Bool("unpinned", false, "permit missing pin only; never a mismatch")
	fs.Usage = func() {
		fmt.Fprintln(&help, "lictor witness [run|init|step|finalize] --repo ABS --record PATH --label TEXT -- name=argv...\ninit also accepts plan names; step takes one entry; finalize takes none.\nNo shell expansion. All entries and dependencies are validated before execution.\nExits: run/finalize 0 green, 1 red; 2 cannot-evaluate; lock 3; live session 4; step preserves target exit.\nAn external legacy digest record is diagnostic only (EMPTY-TREE). check is not implemented.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Fprint(out, help.String())
			return 0
		}
		return commandError(wantJSON, out, diag, err.Error())
	}
	if !filepath.IsAbs(o.Repo) {
		return commandError(wantJSON, out, diag, "--repo must be an absolute path")
	}
	o.Repo = filepath.Clean(o.Repo)
	if asOf != "" {
		t, err := time.Parse(time.RFC3339, asOf)
		if err != nil {
			return commandError(wantJSON, out, diag, "invalid --as-of: "+err.Error())
		}
		o.Now = func() time.Time { return t }
	}
	pin, allowed := checkPin(o.Repo, *unpinned, diag)
	if !allowed {
		return 2
	}
	o.Entries = fs.Args()
	o.Requires = requires
	o.Siblings = siblings
	fmt.Fprintf(diag, "repository: %s\n", o.Repo)
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()
	dst := out
	if wantJSON {
		dst = io.Discard
	}
	r := witness.Run(ctx, o, dst, diag)
	if wantJSON {
		if emit(out, struct {
			witness.Result
			pinMetadata
		}{r, pin}) != 0 {
			return 2
		}
	} else if r.ExitCode >= 2 {
		fmt.Fprintln(diag, r.Reason)
	} else if o.Mode == "init" || o.Mode == "step" {
		fmt.Fprintln(out, r.Reason)
	}
	return r.ExitCode
}
