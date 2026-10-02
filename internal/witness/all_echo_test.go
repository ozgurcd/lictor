package witness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// RULE: WITNESS-ALL-NOT-RUN-ECHO-1
func TestRule_WITNESS_ALL_NOT_RUN_ECHO_1(t *testing.T) {
	o := fixture(t)
	o.All = true
	marker := filepath.Join(t.TempDir(), "blocked-ran")
	o.Entries = []string{command(t, "green", "exit", "0"), command(t, "red", "exit", "7"), command(t, "blocked", "touch", marker), command(t, "last", "exit", "0")}
	o.Requires = []string{"blocked:red"}
	r, out, diag := invoke(o)
	want := "==> gate-witness: green\n==> gate-witness: red\n==> gate-witness: blocked\ncheck FAILED: NOT-RUN blocked: dependency red recorded exit=7\n==> gate-witness: last\n"
	if diag != want {
		t.Errorf("plan-order console mismatch\nwant:\n%sgot:\n%s", want, diag)
	}
	if out != "gate-witness: GATE-RUN.txt result: red\n" || r.ExitCode != 1 || !r.Minted {
		t.Errorf("verdict/summary changed: %+v stdout=%q", r, out)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Errorf("blocked target executed: %v", err)
	}
	wantRecord := "evidence: [blocked] check FAILED: NOT-RUN blocked: dependency red recorded exit=7\nelapsed: blocked 0s\ntarget: blocked exit=125\n"
	if !strings.Contains(record(t, o), wantRecord) || !strings.Contains(record(t, o), "target: last exit=0\n") {
		t.Errorf("blocked record or independent target missing: %s", record(t, o))
	}
}

// The dirty cases remain covered by TestDirtyEchoSourceStreams, unchanged.
func TestCleanAllSourceStreams(t *testing.T) {
	if *allScript == "" || *witnessCLI == "" {
		t.Skip("explicit source and built CLI required")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	if err := os.Symlink(exe, filepath.Join(bin, "date")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, kind := range []string{"green", "red", "not-run"} {
		t.Run(kind, func(t *testing.T) {
			o := fixture(t)
			o.All = true
			echoCLIPin(t, o)
			want := echoPlan(t, &o, kind)
			args := []string{*allScript, o.Record, o.Label}
			for _, dependency := range o.Requires {
				args = append(args, "--requires", dependency)
			}
			args = append(append(args, "--"), o.Entries...)
			sout, serr, scode := echoProcess(t, o.Repo, "bash", args...)
			srecord := record(t, o)
			pout, perr, pcode := echoProcess(t, o.Repo, *witnessCLI, echoCLIArgs(o)...)
			// Assert the complete declared clean console shape, not a filter
			// that could discard an unexpected line or missing target.
			verdict := "green"
			if want != 0 {
				verdict = "red"
			}
			if serr != "" || perr != sout+"repository: "+o.Repo+"\n" || pout != "gate-witness: GATE-RUN.txt result: "+verdict+"\n" {
				t.Errorf("clean stream mismatch\nsource stdout=%q stderr=%q\nport stdout=%q stderr=%q", sout, serr, pout, perr)
			}
			if scode != want || pcode != scode || record(t, o) != srecord {
				t.Errorf("record/exit mismatch: source=%d port=%d want=%d records equal=%t", scode, pcode, want, record(t, o) == srecord)
			}
			if !t.Failed() {
				t.Logf("clean %s: records byte-equal; exit source=port=%d; complete streams equal except declared routing, repository context and summary", kind, pcode)
			}
		})
	}
}

func TestStepBytesUnchanged(t *testing.T) {
	if *witnessCLI == "" || *witnessBeforeCLI == "" {
		t.Skip("explicit before/after CLIs of the same version required")
	}
	for _, dirty := range []bool{false, true} {
		t.Run(fmt.Sprintf("dirty-%t", dirty), func(t *testing.T) {
			o := fixture(t)
			echoCLIPin(t, o)
			if dirty {
				if err := os.WriteFile(filepath.Join(o.Repo, "tracked"), []byte("dirty\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			steps := []struct {
				mode    string
				entries []string
				code    int
			}{
				{"init", []string{"green", "red", "last"}, 0},
				{"step", []string{command(t, "green", "exit", "0")}, 0},
				{"step", []string{command(t, "red", "exit", "7")}, 7},
				{"step", []string{command(t, "last", "exit", "0")}, 0},
				{"finalize", nil, 1},
			}
			type observation struct {
				out, diag, record string
				code              int
			}
			var before []observation
			for pass, cli := range []string{*witnessBeforeCLI, *witnessCLI} {
				for i, step := range steps {
					o.Entries = step.entries
					args := echoCLIArgs(o)
					args = append([]string{"witness", step.mode}, args[1:]...)
					out, diag, code := echoProcess(t, o.Repo, cli, args...)
					got := observation{out, diag, record(t, o), code}
					if code != step.code {
						t.Fatalf("%s exit=%d want=%d: %s", step.mode, code, step.code, diag)
					}
					if pass == 0 {
						before = append(before, got)
					} else if got != before[i] {
						t.Errorf("%s output/record changed: before=%+v after=%+v", step.mode, before[i], got)
					}
				}
			}
			if !t.Failed() {
				t.Logf("stepwise dirty=%t: init, green/red/last steps and finalize stdout, stderr, records and exits byte-identical", dirty)
			}
		})
	}
}
