package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func fixedClock() time.Time { return time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC) }

func invoke(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, diagnostic bytes.Buffer
	code := run(context.Background(), args, &out, &diagnostic, fixedClock)
	return code, strings.TrimSpace(out.String()), diagnostic.String()
}

func validateDocument(t *testing.T, name, document string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("../../schemas", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec, value any
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(document), &value); err != nil {
		t.Fatal(err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", spec); err != nil {
		t.Fatal(err)
	}
	schema, err := c.Compile("schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(value); err != nil {
		t.Fatal(err)
	}
}

func TestVersionAndCapabilitiesSchemas(t *testing.T) {
	for _, command := range []string{"version", "capabilities"} {
		code, document, _ := invoke(t, command, "--json")
		if code != 0 {
			t.Fatalf("%s: exit %d: %s", command, code, document)
		}
		validateDocument(t, "lictor."+command+".v1", document)
	}
}

func TestHumanAndJSONAgreeForImageAndRefusals(t *testing.T) {
	root := t.TempDir()
	declarePin(t, root, version)
	scan := filepath.Join(root, "scan.json")
	if err := os.WriteFile(scan, []byte(`{"source":{"type":"image","target":{"userInput":"fixture:verify"}},"matches":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		args []string
		code int
	}{
		{[]string{"grype", "--repo", root, "-scan", scan, "--as-of", "2026-09-16T14:00:00+02:00"}, 0},
		{[]string{"grype", "--repo", root, "-scan", filepath.Join(root, "absent.json"), "--as-of", "2026-09-16T00:00:00Z"}, 2},
		{[]string{"grype", "--repo", "relative"}, 2},
		{[]string{"grype", "--repo", root, "--as-of", "yesterday"}, 2},
		{[]string{"grype", "--repo", root, "--unknown"}, 2},
	} {
		code, human, diagnostic := invoke(t, tc.args...)
		jcode, document, _ := invoke(t, append(tc.args, "--json")...)
		var result grypeResult
		if err := json.Unmarshal([]byte(document), &result); err != nil {
			t.Fatal(err)
		}
		if code != tc.code || jcode != code || result.ExitCode != code || result.Evaluation.Reason != human {
			t.Fatalf("disagreement: human=(%d,%q), JSON=(%d,%s)", code, human, jcode, document)
		}
		if !strings.Contains(diagnostic, "repository: "+result.Repository) {
			t.Fatalf("repository missing from human context: %s", diagnostic)
		}
		validateDocument(t, "lictor.grype.v1", document)
		if code == 0 && (result.AsOf == nil || *result.AsOf != "2026-09-16T12:00:00Z" || !strings.Contains(human, "config: not applicable") || !strings.Contains(human, "coverage: not applicable")) {
			t.Fatalf("image context: %s", document)
		}
	}
}

func TestDefaultClockReadOnceAndOverrideNeverReadsClock(t *testing.T) {
	root := t.TempDir()
	declarePin(t, root, version)
	calls := 0
	clock := func() time.Time { calls++; return fixedClock() }
	var out bytes.Buffer
	run(context.Background(), []string{"grype", "--repo", root, "-coverage-only", "--json"}, &out, &bytes.Buffer{}, clock)
	if calls != 1 {
		t.Fatalf("default clock reads=%d", calls)
	}
	out.Reset()
	calls = 0
	run(context.Background(), []string{"grype", "--repo", root, "-coverage-only", "--as-of", "2026-09-16T00:00:00Z", "--json"}, &out, &bytes.Buffer{}, clock)
	if calls != 0 {
		t.Fatalf("override read the clock %d times", calls)
	}
}

func TestDefaultRepositoryIsCWD(t *testing.T) {
	root := t.TempDir()
	declarePin(t, root, version)
	t.Chdir(root)
	_, document, _ := invoke(t, "grype", "-coverage-only", "--json", "--as-of", "2026-09-16T00:00:00Z")
	var result grypeResult
	if err := json.Unmarshal([]byte(document), &result); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if result.Repository != cwd {
		t.Fatalf("selected %q, want cwd %q", result.Repository, cwd)
	}
}

func TestReplayIsByteIdenticalAndMissingScannerRefuses(t *testing.T) {
	root := t.TempDir()
	declarePin(t, root, version)
	scan := filepath.Join(root, "scan.json")
	input := []byte(`{"source":{"type":"image","target":{"userInput":"fixture:verify"}},"matches":[]}`)
	if err := os.WriteFile(scan, input, 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{"grype", "--repo", root, "-scan", scan, "--as-of", "2026-09-16T00:00:00Z", "--json"}
	one, a, _ := invoke(t, args...)
	two, b, _ := invoke(t, args...)
	after, err := os.ReadFile(scan)
	if err != nil {
		t.Fatal(err)
	}
	if one != 0 || two != 0 || a != b || !bytes.Equal(input, after) {
		t.Fatalf("replay changed: %d %d %q %q", one, two, a, b)
	}
	if err := os.WriteFile(filepath.Join(root, ".grype.yaml"), []byte("exclude:\n  - ./bin/**\ndb:\n  validate-age: true\n  max-allowed-built-age: 120h\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", root)
	code, document, _ := invoke(t, "grype", "--repo", root, "--as-of", "2026-09-16T00:00:00Z", "--json")
	if code != 2 || !strings.Contains(document, "grype could not run") {
		t.Fatalf("scanner absence: %d %s", code, document)
	}
	validateDocument(t, "lictor.grype.v1", document)
}

func TestCommandErrorsHaveJSONDocuments(t *testing.T) {
	for _, args := range [][]string{{"unknown", "--json"}, {"version", "--bad", "--json"}, {"capabilities", "extra", "--json"}} {
		code, document, _ := invoke(t, args...)
		if code != 2 {
			t.Fatalf("exit=%d: %s", code, document)
		}
		validateDocument(t, "lictor.error.v1", document)
	}
}
