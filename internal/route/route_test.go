package route

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ozgurcd/lictor/internal/executor"
)

type fixtureTools struct {
	version           string
	versionErr, error error
	output            string
	policies          []executor.RoutePolicy
}

func (f *fixtureTools) Version(context.Context, string, io.Writer) ([]byte, error) {
	return []byte(f.version), f.versionErr
}
func (f *fixtureTools) Check(_ context.Context, _ string, p executor.RoutePolicy, _ bool, _ io.Writer) ([]byte, error) {
	f.policies = append(f.policies, p)
	return []byte(f.output), f.error
}
func tools() *fixtureTools {
	return &fixtureTools{version: `{"schema_version":"achta.version.v1","version":"v0.5.10","version_agreement":"pass"}`, output: `{"schema_version":"achta.declared-route-check.v1","status":"pass"}`}
}
func workflow(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".github/workflows")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ci.yml"), []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return root
}

// RULE: INSTALL-ROUTE-1
func TestRule_INSTALL_ROUTE_1(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		keys       []string
	}{
		{"declarations-only", "env:\n  LICTOR_VERSION: v0.2.0\n# brew install rulefloor lictor\n", nil},
		{"rulefloor", "run: curl rulefloor/archive/refs/tags/${RULEFLOOR_VERSION}\n", []string{"RULEFLOOR_VERSION"}},
		{"lictor", "run: curl lictor/releases/download/${LICTOR_VERSION}/lictor_${LICTOR_VERSION#v}_\n", []string{"LICTOR_VERSION", "LICTOR_SHA256"}},
		{"both", "run: brew install rulefloor lictor\n", []string{"RULEFLOOR_VERSION", "LICTOR_VERSION", "LICTOR_SHA256"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := tools()
			r := Run(t.Context(), workflow(t, tc.body), nil, true, f, io.Discard)
			var keys []string
			for _, p := range f.policies {
				keys = append(keys, p.Key)
			}
			if r.ExitCode != 0 || !reflect.DeepEqual(keys, tc.keys) {
				t.Fatalf("selected keys=%v want=%v result=%+v", keys, tc.keys, r)
			}
			for _, set := range r.Sets {
				if set.Skipped && (!strings.Contains(r.Human, set.Tool+" set skipped: no install line") || len(set.Checks) != 0) {
					t.Fatalf("silent skip: %+v", r)
				}
			}
		})
	}
	f := tools()
	f.output = "native Achta evidence\n"
	root := workflow(t, "run: go install rulefloor\n")
	r := Run(t.Context(), root, nil, false, f, io.Discard)
	if r.Human != "native Achta evidence\nlictor set skipped: no install line\n" {
		t.Fatalf("rewritten native evidence: %q", r.Human)
	}
	f.error = exitError(t, "1")
	f.output = `{"schema_version":"achta.declared-route-check.v1","status":"fail"}`
	r = Run(t.Context(), root, nil, true, f, io.Discard)
	if r.ExitCode != 1 || r.Outcome != "fail" || string(r.Sets[0].Checks[0].Document) != f.output {
		t.Fatalf("lost native red: %+v", r)
	}
}

func TestRouteRefusals(t *testing.T) {
	root := workflow(t, "run: go install lictor\n")
	for _, name := range []string{"absent", "bad-version", "mismatch", "unstartable", "bad-json", "exit-two", "bad-exit"} {
		t.Run(name, func(t *testing.T) {
			f := tools()
			var declared *string
			switch name {
			case "absent":
				f.versionErr = exec.ErrNotFound
			case "bad-version":
				f.version = "{}"
			case "mismatch":
				v := "v1.0.0"
				declared = &v
			case "unstartable":
				f.error = errors.New("cannot execute")
			case "bad-json":
				f.output = "not JSON"
			case "exit-two":
				f.error = exitError(t, "2")
				f.output = `{"schema_version":"achta.error.v1"}`
			case "bad-exit":
				f.error = exitError(t, "3")
			}
			r := Run(t.Context(), root, declared, true, f, io.Discard)
			if r.ExitCode != 2 || r.Outcome != "cannot_evaluate" {
				t.Fatalf("refusal: %+v", r)
			}
		})
	}
	if r := Run(t.Context(), t.TempDir(), nil, true, tools(), io.Discard); r.ExitCode != 2 || !strings.Contains(r.Reason, "workflow") {
		t.Fatalf("missing workflow: %+v", r)
	}
	f := tools()
	v := "v0.5.10"
	var stderr bytes.Buffer
	if r := Run(t.Context(), root, &v, true, f, &stderr); r.ExitCode != 0 || stderr.String() != "achta: installed v0.5.10; declared v0.5.10\n" {
		t.Fatalf("matching version: %+v %s", r, stderr.String())
	}
}

func TestExactHousePatterns(t *testing.T) {
	r := Policies("rulefloor")
	want := []string{`brew install[^#]*rulefloor`, `go install[^#]*rulefloor`, `rulefloor/archive/refs/tags/v[0-9]`, `-X main[.]version=[$][{]RULEFLOOR_VERSION[^}]`, `-X main[.]version=[$][{][^R]`, `-X main[.]version=[$][^{]`, `-X main[.]version=[^$]`}
	if len(r) != 1 || r[0].Pattern != `rulefloor/archive/refs/tags/[$][{]RULEFLOOR_VERSION[}]` || !reflect.DeepEqual(r[0].Bans, want) || r[0].Key != "RULEFLOOR_VERSION" {
		t.Fatalf("rulefloor policy: %+v", r)
	}
	l := Policies("lictor")
	if len(l) != 2 || l[0].Key != "LICTOR_VERSION" || l[1].Key != "LICTOR_SHA256" || l[0].Pattern != `lictor/releases/download/[$][{]LICTOR_VERSION[}]/lictor_[$][{]LICTOR_VERSION#v[}]_` || !reflect.DeepEqual(l[0].Bans, []string{`brew install[^#]*lictor`, `go install[^#]*lictor`, `lictor/releases/download/v[0-9]`}) {
		t.Fatalf("lictor policy: %+v", l)
	}
}

func exitError(t *testing.T, code string) error {
	t.Helper()
	err := exec.CommandContext(t.Context(), os.Args[0], "-test.run=TestRouteExitHelper", "--", code).Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("exit helper: %v", err)
	}
	return err
}
func TestRouteExitHelper(t *testing.T) {
	if len(os.Args) > 2 && os.Args[len(os.Args)-2] == "--" {
		switch os.Args[len(os.Args)-1] {
		case "1":
			os.Exit(1)
		case "2":
			os.Exit(2)
		case "3":
			os.Exit(3)
		}
	}
}
