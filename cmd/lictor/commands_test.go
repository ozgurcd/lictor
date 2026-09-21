package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCommandsRefuseMissingInputs(t *testing.T) {
	for _, tc := range []struct{ command, reason string }{
		{"clockfuse", "no tools/clockfuse"}, {"route", "achta version could not run"},
	} {
		t.Run(tc.command, func(t *testing.T) {
			root := t.TempDir()
			declarePin(t, root, version)
			if tc.command == "route" {
				t.Setenv("PATH", root)
			}
			code, out, stderr := invoke(t, tc.command, "--repo", root, "--json")
			if code != 2 || !strings.Contains(out, tc.reason) {
				t.Fatalf("exit=%d stdout=%s stderr=%s", code, out, stderr)
			}
			validateDocument(t, "lictor."+tc.command+".v1", out)
		})
	}
}

func TestClockfuseCLIExplicitSnapshotThenReadOnlyCheck(t *testing.T) {
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOSUMDB", "off")
	t.Setenv("GOTOOLCHAIN", "local")
	root := t.TempDir()
	declarePin(t, root, version)
	if err := os.MkdirAll(filepath.Join(root, "tools/clockfuse"), 0700); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"go.mod":                  "module fixture\n\ngo 1.21\n",
		"tools/clockfuse/main.go": "package main\nimport (\"fmt\"; \"os\")\nfunc main() {fmt.Println(\"  a_test.go:10: Config{...} omits `Now`\");os.Exit(1)}\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	code, document, stderr := invoke(t, "clockfuse", "--repo", root, "--snapshot", "--json")
	if code != 0 {
		t.Fatalf("snapshot: %d %s %s", code, document, stderr)
	}
	validateDocument(t, "lictor.clockfuse.v1", document)
	before, err := os.ReadFile(filepath.Join(root, ".clockfuse-snapshot"))
	if err != nil {
		t.Fatal(err)
	}
	code, human, stderr := invoke(t, "clockfuse", "--repo", root)
	after, err := os.ReadFile(filepath.Join(root, ".clockfuse-snapshot"))
	if code != 0 || err != nil || human != "clockfuse-gate: no findings above snapshot." || string(before) != string(after) || !strings.Contains(stderr, "repository: "+root) {
		t.Fatalf("check: %d %q %q %v", code, human, stderr, err)
	}
	_, doc, _ := invoke(t, "clockfuse", "--repo", root, "--json")
	var parsed map[string]any
	if err := json.Unmarshal([]byte(doc), &parsed); err != nil || parsed["reason"] != human || parsed["snapshot_written"] != false || parsed["pinned"] != true {
		t.Fatalf("JSON agreement: %s %v", doc, err)
	}
}

func TestNewCommandHelpAndArgumentRefusals(t *testing.T) {
	for _, command := range []string{"clockfuse", "route"} {
		code, out, _ := invoke(t, command, "--help")
		if code != 0 || !strings.Contains(out, "unpinned") || !strings.Contains(out, "repo") {
			t.Fatalf("help %s: %d %s", command, code, out)
		}
		for _, args := range [][]string{{"--repo", "relative"}, {"--bad"}, {"extra"}} {
			args = append([]string{command, "--json"}, args...)
			code, out, _ := invoke(t, args...)
			if code != 2 {
				t.Fatalf("arguments: %v: %d %s", args, code, out)
			}
			validateDocument(t, "lictor.error.v1", out)
		}
	}
}
