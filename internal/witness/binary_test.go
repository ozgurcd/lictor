package witness

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestBinaryOutputPreservesExitAndTextRecord(t *testing.T) {
	for _, tc := range []struct{ name, output, kind string }{
		{"nul-summary", "check OK: \x00\nok pkg 0.1s\n", "NUL"},
		{"nul", "check OK: before\nok pkg 0.1s\nignored\x00bytes\ncheck OK: after\n", "NUL"},
		{"nul-first", "\x00\ncheck OK: after\n", "NUL"},
		{"invalid-utf8", "check OK: \xff\nok pkg 0.1s\n", "invalid UTF-8"},
		{"invalid-unmatched-line", "check OK: before\nignored \xff\n", "invalid UTF-8"},
		{"nul-and-invalid", "\xff\ncheck OK: \x00\n", "NUL"},
	} {
		for _, name := range []string{"probe", "tool-versions"} {
			for _, exit := range []int{0, 7} {
				t.Run(fmt.Sprintf("%s/%s/exit-%d", tc.name, name, exit), func(t *testing.T) {
					o := fixture(t)
					file := filepath.Join(t.TempDir(), "output")
					if err := os.WriteFile(file, []byte(tc.output), 0600); err != nil {
						t.Fatal(err)
					}
					e := Entry{Name: name, Argv: []string{os.Args[0], "-test.run=TestBinaryOutputHelper", "--", "--binary-output", file, strconv.Itoa(exit)}}
					var previous string
					for repeat := 0; repeat < 2; repeat++ {
						var record, diagnostic bytes.Buffer
						code, err := one(t.Context(), o, e, &record, &diagnostic)
						want := fmt.Sprintf("binary-output: %s contains %s; evidence omitted\nelapsed: %s 0s\ntarget: %s exit=%d\n", name, tc.kind, name, name, exit)
						if err != nil || code != exit || record.String() != want || !utf8.Valid(record.Bytes()) || bytes.ContainsRune(record.Bytes(), 0) {
							t.Fatalf("exit=%d err=%v record=%q; want exit=%d record=%q", code, err, record.String(), exit, want)
						}
						if diagnostic.String() != "==> gate-witness: "+name+"\n"+tc.output {
							t.Fatalf("diagnostic stream changed: %q", diagnostic.String())
						}
						if repeat > 0 && previous != record.String() {
							t.Fatal("binary record depends on temporary path or run")
						}
						previous = record.String()
					}
					t.Logf("twice byte-identical: %q", previous)
				})
			}
		}
	}
}

func TestBinaryOutputHelper(t *testing.T) {
	if len(os.Args) < 4 || os.Args[len(os.Args)-3] != "--binary-output" {
		return
	}
	b, err := os.ReadFile(os.Args[len(os.Args)-2])
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stdout.Write(b); err != nil {
		t.Fatal(err)
	}
	code, err := strconv.Atoi(os.Args[len(os.Args)-1])
	if err != nil {
		t.Fatal(err)
	}
	os.Exit(code)
}

func TestBinaryTargetFinalizesWithRealVerdict(t *testing.T) {
	for _, exit := range []int{0, 7} {
		t.Run(strconv.Itoa(exit), func(t *testing.T) {
			o := fixture(t)
			file := filepath.Join(t.TempDir(), "output")
			if err := os.WriteFile(file, []byte("check OK: before\n\x00\n"), 0600); err != nil {
				t.Fatal(err)
			}
			o.Entries = []string{fmt.Sprintf("probe=%q -test.run=TestBinaryOutputHelper -- --binary-output %q %d", os.Args[0], file, exit)}
			r := Run(t.Context(), o, io.Discard, io.Discard)
			want, verdict := 0, "green"
			if exit != 0 {
				want, verdict = 1, "red"
			}
			body := record(t, o)
			if r.ExitCode != want || !r.Minted || !strings.Contains(body, "result: "+verdict+"\n") || !strings.Contains(body, fmt.Sprintf("target: probe exit=%d\n", exit)) || !strings.Contains(body, "binary-output: probe contains NUL; evidence omitted\n") || strings.Contains(body, "evidence:") || !utf8.ValidString(body) || strings.ContainsRune(body, 0) {
				t.Fatalf("result=%+v record=%q", r, body)
			}
		})
	}
}

func TestRecordMetadataMustBeTextBeforeOpening(t *testing.T) {
	for _, field := range []string{"label", "cites"} {
		for _, value := range []string{"bad\x00text", "bad\xfftext"} {
			t.Run(fmt.Sprintf("%s/%q", field, value), func(t *testing.T) {
				o := fixture(t)
				if field == "label" {
					o.Label = value
				} else {
					o.Cites = value
				}
				o.Entries = []string{command(t, "probe", "summary")}
				r := Run(t.Context(), o, io.Discard, io.Discard)
				if r.ExitCode != 2 || !strings.Contains(r.Reason, "UTF-8 text without NUL") {
					t.Fatalf("invalid metadata accepted: %+v", r)
				}
				if _, err := os.Stat(filepath.Join(o.Repo, o.Record)); !os.IsNotExist(err) {
					t.Fatalf("record opened: %v", err)
				}
			})
		}
	}
}
