package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/ozgurcd/lictor/internal/grype"
)

const version = "v0.4.0"

type evaluation struct {
	Outcome string `json:"outcome"`
	Reason  string `json:"reason"`
}

type grypeResult struct {
	pinMetadata
	SchemaVersion string     `json:"schema_version"`
	Repository    string     `json:"repository"`
	AsOf          *string    `json:"as_of"`
	Evaluation    evaluation `json:"evaluation"`
	ExitCode      int        `json:"exit_code"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr, time.Now))
}

func emit(w io.Writer, value any) int {
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return 2
	}
	return 0
}

func commandError(wantJSON bool, out, errOut io.Writer, reason string) int {
	if wantJSON {
		_ = emit(out, map[string]any{"schema_version": "lictor.error.v1", "evaluation": evaluation{"cannot_evaluate", reason}, "exit_code": 2})
	} else {
		fmt.Fprintln(errOut, "CANNOT-EVALUATE: "+reason)
	}
	return 2
}

func run(ctx context.Context, args []string, out, errOut io.Writer, now func() time.Time) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		fmt.Fprintln(out, "lictor: Identuum gate execution\nCommands: version, capabilities, grype, green, clockfuse, route, witness\nUse lictor COMMAND --help for command options.")
		return 0
	}
	wantJSON := false
	for _, arg := range args[1:] {
		if arg == "--json" || arg == "-json" || arg == "--json=true" {
			wantJSON = true
		}
	}
	if args[0] == "version" || args[0] == "capabilities" {
		fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
		fs.SetOutput(errOut)
		fs.BoolVar(&wantJSON, "json", wantJSON, "emit one JSON document")
		if err := fs.Parse(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			return commandError(wantJSON, out, errOut, err.Error())
		}
		if fs.NArg() != 0 {
			return commandError(wantJSON, out, errOut, "unexpected positional argument")
		}
		if args[0] == "version" {
			if wantJSON {
				return emit(out, map[string]string{"schema_version": "lictor.version.v1", "version": version})
			}
			fmt.Fprintln(out, "lictor "+version)
			return 0
		}
		value := map[string]any{"schema_version": "lictor.capabilities.v1", "version": version, "commands": []string{"version", "capabilities", "grype", "green", "clockfuse", "route", "witness"}, "machine_interfaces": []string{"lictor.version.v1", "lictor.capabilities.v1", "lictor.grype.v1", "lictor.green.v1", "lictor.clockfuse.v1", "lictor.route.v1", "lictor.witness.v1"}, "exit_codes": map[string]int{"pass": 0, "fail": 1, "cannot_evaluate": 2}, "repository_selection": "--repo absolute path; default current working directory", "replay": "grype -scan JSON -inventory JSON --as-of RFC3339; unchanged repository inputs required", "witness_exits": "run/finalize 0 green, 1 red, 2 cannot-evaluate; lock 3; session 4; step preserves target exit"}
		if wantJSON {
			return emit(out, value)
		}
		fmt.Fprintln(out, "lictor "+version+": version, capabilities, grype, green, clockfuse, route, witness; exits 0 pass, 1 fail, 2 cannot_evaluate; witness lock 3, session 4, step target exit; --repo absolute path (default cwd); grype -scan with --as-of replays Lictor's dated judgement")
		return 0
	}
	if args[0] == "witness" {
		return runWitness(ctx, args[1:], wantJSON, out, errOut)
	}
	if args[0] == "green" {
		return runGreen(ctx, args[1:], wantJSON, out, errOut)
	}
	if args[0] == "clockfuse" || args[0] == "route" {
		return runHouse(ctx, args[0], args[1:], wantJSON, out, errOut)
	}
	if args[0] != "grype" {
		if wantJSON {
			returnCode := emit(out, map[string]any{"schema_version": "lictor.error.v1", "evaluation": evaluation{"cannot_evaluate", "unknown command: " + args[0]}, "exit_code": 2})
			if returnCode != 0 {
				return returnCode
			}
			return 2
		}
		fmt.Fprintln(errOut, "CANNOT-EVALUATE: unknown command: "+args[0])
		return 2
	}
	return runGrype(ctx, args[1:], wantJSON, out, errOut, now)
}

func runGrype(ctx context.Context, args []string, wantJSON bool, out, errOut io.Writer, now func() time.Time) int {
	cwd, cwdErr := os.Getwd()
	opts := grype.Options{Repository: cwd, RepositoryExplicit: true}
	fs := flag.NewFlagSet("lictor grype", flag.ContinueOnError)
	var diagnostic bytes.Buffer
	fs.SetOutput(&diagnostic)
	fs.StringVar(&opts.Repository, "repo", cwd, "absolute repository path; compatibility default is current working directory")
	unpinned := fs.Bool("unpinned", false, "permit an absent version declaration; never bypass a mismatch")
	fs.StringVar(&opts.Scan, "scan", "", "saved Grype JSON report; otherwise run grype dir:. in the repository")
	fs.StringVar(&opts.Inventory, "inventory", "", "matching CycloneDX JSON inventory; required for saved directory reports")
	fs.StringVar(&opts.Allowlist, "allowlist", "grype-allowlist.json", "allowlist path; relative paths resolve under --repo")
	fs.BoolVar(&opts.CoverageOnly, "coverage-only", false, "judge inventory coverage only; no scanner")
	asOf := fs.String("as-of", "", "RFC3339 evaluation time for Lictor suppression expiry only; Grype's clock is unchanged")
	fs.BoolVar(&wantJSON, "json", wantJSON, "emit lictor.grype.v1; human evidence text is evaluation.reason")
	r := grypeResult{SchemaVersion: "lictor.grype.v1", Repository: cwd}
	finish := func(code int, reason string) int {
		r.ExitCode, r.Evaluation = code, evaluation{[]string{"pass", "fail", "cannot_evaluate"}[code], reason}
		if wantJSON {
			if emit(out, r) != 0 {
				return 2
			}
		} else {
			fmt.Fprintf(errOut, "repository: %s\n", r.Repository)
			if r.AsOf != nil {
				fmt.Fprintf(errOut, "as_of: %s (Lictor judgements only)\n", *r.AsOf)
			}
			if _, err := fmt.Fprintln(out, reason); err != nil {
				return 2
			}
		}
		return code
	}
	err := fs.Parse(args)
	r.Repository = opts.Repository
	if err == flag.ErrHelp {
		fmt.Fprint(out, diagnostic.String())
		return 0
	}
	if err != nil {
		return finish(2, "CANNOT-EVALUATE: grype-gate: "+err.Error())
	}
	if cwdErr != nil && opts.Repository == "" {
		return finish(2, "CANNOT-EVALUATE: repository: "+cwdErr.Error())
	}
	if fs.NArg() != 0 {
		return finish(2, "CANNOT-EVALUATE: grype-gate: unexpected positional argument")
	}
	if !filepath.IsAbs(opts.Repository) {
		return finish(2, "CANNOT-EVALUATE: --repo must be an absolute path")
	}
	opts.Repository = filepath.Clean(opts.Repository)
	r.Repository = opts.Repository
	st, err := os.Stat(opts.Repository)
	if err != nil {
		return finish(2, "CANNOT-EVALUATE: repository: "+err.Error())
	}
	if !st.IsDir() {
		return finish(2, "CANNOT-EVALUATE: --repo must name a directory")
	}
	var allowed bool
	r.pinMetadata, allowed = checkPin(opts.Repository, *unpinned, errOut)
	if !allowed {
		return 2
	}
	if *asOf == "" {
		opts.AsOf = now().UTC()
	} else {
		opts.AsOf, err = time.Parse(time.RFC3339, *asOf)
		if err != nil {
			return finish(2, "CANNOT-EVALUATE: --as-of must be RFC3339: "+err.Error())
		}
		opts.AsOf = opts.AsOf.UTC()
	}
	stamp := opts.AsOf.Format(time.RFC3339Nano)
	r.AsOf = &stamp
	if !filepath.IsAbs(opts.Allowlist) {
		opts.Allowlist = filepath.Join(opts.Repository, opts.Allowlist)
	}
	var evidence bytes.Buffer
	code := grype.Run(ctx, opts, &evidence, errOut)
	return finish(code, strings.TrimSuffix(evidence.String(), "\n"))
}
