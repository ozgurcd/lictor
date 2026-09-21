package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ozgurcd/lictor/internal/clockfuse"
	"github.com/ozgurcd/lictor/internal/executor"
	"github.com/ozgurcd/lictor/internal/route"
)

func runHouse(ctx context.Context, command string, args []string, wantJSON bool, out, errOut io.Writer) int {
	root, _ := os.Getwd()
	fs := flag.NewFlagSet("lictor "+command, flag.ContinueOnError)
	var diagnostic bytes.Buffer
	fs.SetOutput(&diagnostic)
	fs.StringVar(&root, "repo", root, "absolute repository path; default current working directory")
	unpinned := fs.Bool("unpinned", false, "permit an absent version declaration; never bypass a mismatch")
	fs.BoolVar(&wantJSON, "json", wantJSON, "emit one lictor."+command+".v1 document")
	var snapshot bool
	var achtaWorkspace string
	if command == "clockfuse" {
		fs.BoolVar(&snapshot, "snapshot", false, "explicitly rewrite .clockfuse-snapshot; default check writes nothing")
	}
	if command == "route" {
		fs.StringVar(&achtaWorkspace, "achta-workspace", "", "absolute workspace passed to Achta; resolves its ambiguous discovery without adding Lictor discovery")
	}
	fs.Usage = func() {
		fmt.Fprintf(&diagnostic, "lictor %s: exits 0 pass, 1 fail, 2 cannot-evaluate.\n", command)
		fs.PrintDefaults()
	}
	err := fs.Parse(args)
	if err == flag.ErrHelp {
		fmt.Fprint(out, diagnostic.String())
		return 0
	}
	if err != nil {
		return commandError(wantJSON, out, errOut, err.Error())
	}
	if fs.NArg() != 0 {
		return commandError(wantJSON, out, errOut, "unexpected positional argument")
	}
	if !filepath.IsAbs(root) {
		return commandError(wantJSON, out, errOut, "--repo must be an absolute path")
	}
	root = filepath.Clean(root)
	if achtaWorkspace != "" && !filepath.IsAbs(achtaWorkspace) {
		return commandError(wantJSON, out, errOut, "--achta-workspace must be an absolute path")
	}
	st, err := os.Stat(root)
	if err != nil || !st.IsDir() {
		return commandError(wantJSON, out, errOut, "--repo must name a readable directory")
	}
	pin, allowed := checkPin(root, *unpinned, errOut)
	if !allowed {
		return 2
	}
	if command == "clockfuse" {
		r := clockfuse.Run(ctx, root, snapshot, executor.GoTools{})
		if wantJSON {
			if emit(out, struct {
				clockfuse.Result
				pinMetadata
			}{r, pin}) != 0 {
				return 2
			}
		} else {
			dst := out
			if r.ExitCode == 2 {
				dst = errOut
			}
			if _, err := fmt.Fprintln(dst, r.Reason); err != nil {
				return 2
			}
		}
		return r.ExitCode
	}
	declared, err := declaredVersion(root, "ACHTA_VERSION")
	if err != nil {
		return commandError(wantJSON, out, errOut, err.Error())
	}
	r := route.Run(ctx, root, declared, wantJSON, executor.Achta{Workspace: achtaWorkspace}, errOut)
	if wantJSON {
		if emit(out, struct {
			route.Result
			pinMetadata
		}{r, pin}) != 0 {
			return 2
		}
	} else if _, err := fmt.Fprint(out, r.Human); err != nil {
		return 2
	}
	return r.ExitCode
}
