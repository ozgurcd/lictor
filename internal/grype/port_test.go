package grype

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"
)

// The ported fixtures keep their argument-shaped helper. Production flag
// parsing lives in cmd/lictor. Every fixture supplies a fixed evaluation date.
func run(args []string, out, errOut io.Writer) int {
	var opts Options
	fs := flag.NewFlagSet("fixture", flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.StringVar(&opts.Repository, "repo", ".", "repository")
	fs.StringVar(&opts.Scan, "scan", "", "scan")
	fs.StringVar(&opts.Inventory, "inventory", "", "inventory")
	fs.StringVar(&opts.Allowlist, "allowlist", defaultAllowlist, "allowlist")
	fs.BoolVar(&opts.CoverageOnly, "coverage-only", false, "coverage")
	asOf := fs.String("as-of", "2026-09-16T00:00:00Z", "evaluation time")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(out, err)
		return 2
	}
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "repo" {
			opts.RepositoryExplicit = true
		}
	})
	var err error
	opts.AsOf, err = time.Parse(time.RFC3339, *asOf)
	if err != nil {
		fmt.Fprintln(out, err)
		return 2
	}
	return Run(context.Background(), opts, out, errOut)
}
