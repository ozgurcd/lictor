package green

import (
	"context"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ozgurcd/lictor/internal/executor"
)

var sourceScript = flag.String("green-source-script", "", "optional read-only source script for port comparison; absent in normal offline tests")

type missingGo struct{ executor.GoTools }

func (missingGo) Version(context.Context, string) ([]byte, error) { return nil, exec.ErrNotFound }

func offline(t *testing.T) {
	t.Helper()
	for name, value := range map[string]string{"GOPROXY": "off", "GOSUMDB": "off", "GOTOOLCHAIN": "local", "GOWORK": "off", "GOENV": "off", "GOFLAGS": ""} {
		t.Setenv(name, value)
	}
}

// RULE: GREEN-FLOOR-1
func TestRule_GREEN_FLOOR_1(t *testing.T) {
	offline(t)
	for _, tc := range []struct {
		name, main, test, extra, outcome, cause string
		code                                    int
	}{
		{"green", "func main() {}", "", "", "GREEN", "builds, vets and tests clean", 0},
		{"no-go", "func main() {}", "", "", "CANNOT-EVALUATE", "no 'go' on PATH", 2},
		{"nocompile", "func main() { undefinedSymbol() }", "", "", "NOT-GREEN", "go build FAILED", 1},
		{"testnocompile", "func main() {}", "func TestX(t *testing.T) { missingHelper() }", "", "NOT-GREEN", "go vet FAILED", 1},
		{"testfail", "func main() {}", "func TestX(t *testing.T) { t.Fatal(\"red\") }", "", "NOT-GREEN", "go test FAILED", 1},
		{"vetfail", "import \"fmt\"\n\nfunc main() { fmt.Printf(\"%d\\n\", \"not-an-int\") }", "", "", "NOT-GREEN", "go vet FAILED", 1},
		{"gofmtfail", "func main() {}", "", "package main\n\nfunc  badlyFormatted( ) {\n\t\t_ = 1\n}\n", "NOT-GREEN", "gofmt FAILED — unformatted files", 1},
		{"nomod", "", "", "", "CANNOT-EVALUATE", "not a Go module", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			files := map[string]string{}
			if tc.main != "" {
				files["go.mod"] = "module example.com/" + tc.name + "\n\ngo 1.21\n"
				files["main.go"] = "package main\n\n" + tc.main + "\n"
			}
			if tc.test != "" {
				files["main_test.go"] = "package main\n\nimport \"testing\"\n\n" + tc.test + "\n"
			}
			if tc.extra != "" {
				files["ugly.go"] = tc.extra
			}
			for name, body := range files {
				if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			var tools Tools = executor.GoTools{}
			if tc.name == "no-go" {
				tools = missingGo{}
			}
			got := Run(t.Context(), root, tools)
			if got.ExitCode != tc.code {
				t.Fatalf("exit: want %d, got %+v", tc.code, got)
			}
			if got.Outcome != tc.outcome || got.Cause != tc.cause {
				t.Fatalf("verdict: want %s/%s, got %+v", tc.outcome, tc.cause, got)
			}
			if got.Subject != filepath.Base(root) || got.GoVersion == "" {
				t.Fatalf("identity: %+v", got)
			}
			if tc.code == 1 && got.Excerpt == "" {
				t.Fatalf("missing red evidence: %+v", got)
			}
			t.Logf("port exit=%d %s", got.ExitCode, got.Line())
			if *sourceScript != "" {
				cmd := exec.CommandContext(t.Context(), "/bin/bash", *sourceScript, "--check", root)
				cmd.Env = os.Environ()
				if tc.name == "no-go" {
					// Only the SOURCE uses its own historical PATH fixture. The port
					// above uses an executor refusal and never changes PATH.
					cmd.Env = append(cmd.Environ(), "PATH=/usr/bin:/bin")
				}
				out, err := cmd.CombinedOutput()
				code := 0
				if err != nil {
					if e, ok := err.(*exec.ExitError); ok {
						code = e.ExitCode()
					} else {
						t.Fatal(err)
					}
				}
				want := tc.code
				if want == 2 {
					want = 3
				}
				if code != want || !strings.HasPrefix(string(out), tc.outcome+":") || !strings.Contains(string(out), tc.cause) {
					t.Fatalf("source disagreement: exit=%d %s; port=%+v", code, out, got)
				}
				t.Logf("source exit=%d %s", code, strings.SplitN(string(out), "\n", 2)[0])
			}
		})
	}
}

type orderedTools struct {
	calls  []string
	fail   string
	err    error
	output []byte
}

func (o *orderedTools) Version(context.Context, string) ([]byte, error) {
	o.calls = append(o.calls, "version")
	return []byte("go version go1.27.1 fixture/fixture\n"), nil
}
func (o *orderedTools) step(name string) ([]byte, error) {
	o.calls = append(o.calls, name)
	if name == o.fail {
		return o.output, o.err
	}
	return nil, nil
}
func (o *orderedTools) Format(context.Context, string) ([]byte, error) { return o.step("gofmt") }
func (o *orderedTools) Build(context.Context, string) ([]byte, error)  { return o.step("build") }
func (o *orderedTools) Vet(context.Context, string) ([]byte, error)    { return o.step("vet") }
func (o *orderedTools) Test(context.Context, string) ([]byte, error)   { return o.step("test") }

func TestOrderStopsAtFirstRedAndCannotExecuteRefuses(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	order := []string{"version", "gofmt", "build", "vet", "test"}
	for i, name := range order[1:] {
		o := &orderedTools{fail: name, err: &exec.ExitError{}, output: []byte("FAIL fixture\n")}
		got := Run(t.Context(), root, o)
		if got.ExitCode != 1 || !reflect.DeepEqual(o.calls, order[:i+2]) {
			t.Fatalf("%s: %+v calls=%v", name, got, o.calls)
		}
	}
	o := &orderedTools{fail: "build", err: context.DeadlineExceeded}
	if got := Run(t.Context(), root, o); got.ExitCode != 2 || !strings.Contains(got.Excerpt, "deadline exceeded") {
		t.Fatalf("timeout: %+v", got)
	}
	o = &orderedTools{fail: "gofmt", output: []byte("unformatted.go\n")}
	if got := Run(t.Context(), root, o); got.ExitCode != 1 || len(o.calls) != 2 {
		t.Fatalf("gofmt output: %+v calls=%v", got, o.calls)
	}
}

func TestExcerptBoundsAndTestFilter(t *testing.T) {
	if got := excerpt([]byte(strings.Repeat("line\n", 20)), false); got != strings.TrimSuffix(strings.Repeat("line\n", 12), "\n") {
		t.Fatalf("line bound: %q", got)
	}
	if got := excerpt([]byte(strings.Repeat("x", 20000)), false); len(got) != 16*1024 {
		t.Fatalf("byte bound: %d", len(got))
	}
	if got := excerpt([]byte("noise\n--- FAIL: TestX\n    main_test.go:3: red\nFAIL\n"), true); got != "--- FAIL: TestX\n    main_test.go:3: red\nFAIL" {
		t.Fatalf("test filter: %q", got)
	}
}
