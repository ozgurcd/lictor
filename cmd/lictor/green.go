package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ozgurcd/lictor/internal/executor"
	"github.com/ozgurcd/lictor/internal/green"
)

func runGreen(ctx context.Context, args []string, wantJSON bool, out, errOut io.Writer) int {
	root, _ := os.Getwd()
	fs := flag.NewFlagSet("lictor green", flag.ContinueOnError)
	var diagnostic bytes.Buffer
	fs.SetOutput(&diagnostic)
	fs.StringVar(&root, "repo", root, "absolute repository path; default current working directory")
	fs.BoolVar(&wantJSON, "json", wantJSON, "emit lictor.green.v1 instead of the human evidence line")
	fs.Usage = func() {
		fmt.Fprintln(&diagnostic, "lictor green: gofmt, build, vet, test; stop at first red. Exits: 0 GREEN, 1 NOT-GREEN, 2 CANNOT-EVALUATE.")
		fs.PrintDefaults()
	}
	err := fs.Parse(args)
	if err == flag.ErrHelp {
		fmt.Fprint(out, diagnostic.String())
		return 0
	}
	r := green.Refusal(root, "")
	switch {
	case err != nil:
		r.Cause = err.Error()
	case fs.NArg() != 0:
		r.Cause = "unexpected positional argument"
	case !filepath.IsAbs(root):
		r.Cause = "--repo must be an absolute path"
	default:
		st, statErr := os.Stat(root)
		if statErr != nil || !st.IsDir() {
			r.Cause = "--repo must name a readable directory"
		} else {
			r = green.Run(ctx, filepath.Clean(root), executor.GoTools{})
		}
	}
	if r.Excerpt != "" {
		if _, err := fmt.Fprintln(errOut, r.Excerpt); err != nil {
			return 2
		}
	}
	if wantJSON {
		if emit(out, r) != 0 {
			return 2
		}
	} else if _, err := fmt.Fprintln(out, r.Line()); err != nil {
		return 2
	}
	return r.ExitCode
}
