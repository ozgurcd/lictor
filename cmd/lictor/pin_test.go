package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func declarePin(t *testing.T, root, value string) {
	t.Helper()
	dir := filepath.Join(root, ".github", "workflows")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ci.yml"), []byte("env:\n  LICTOR_VERSION: "+value+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestPinFourCases(t *testing.T) {
	for _, tc := range []struct {
		name, declared string
		unpinned       bool
		code           int
	}{
		{"equal", version, false, 0}, {"different", "v99.0.0", false, 2},
		{"absent", "", false, 2}, {"absent-unpinned", "", true, 0},
		{"different-unpinned", "v99.0.0", true, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.declared != "" {
				declarePin(t, root, tc.declared)
			}
			scan := filepath.Join(root, "scan.json")
			if err := os.WriteFile(scan, []byte(`{"source":{"type":"image","target":{"userInput":"fixture:verify"}},"matches":[]}`), 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{"grype", "--repo", root, "-scan", scan, "--json"}
			if tc.unpinned {
				args = append(args, "--unpinned")
			}
			code, out, stderr := invoke(t, args...)
			if code != tc.code {
				t.Fatalf("exit=%d stdout=%s stderr=%s", code, out, stderr)
			}
			if code == 2 {
				want := "lictor " + version + ": refused — " + filepath.Base(root) + " declares no LICTOR_VERSION in .github/workflows/ci.yml env; pass --unpinned to run without a pin\n"
				if tc.declared != "" {
					want = "lictor " + version + ": refused — " + filepath.Base(root) + " declares LICTOR_VERSION " + tc.declared + "; install the declared version: brew install ozgurcd/tap/lictor\n"
				}
				if out != "" || stderr != want {
					t.Fatalf("refusal: stdout=%q stderr=%q want=%q", out, stderr, want)
				}
			} else {
				var doc map[string]any
				if err := json.Unmarshal([]byte(out), &doc); err != nil {
					t.Fatal(err)
				}
				if doc["pinned"] != (tc.declared != "") || (tc.declared == "" && doc["declared_version"] != nil) || (tc.declared != "" && doc["declared_version"] != tc.declared) {
					t.Fatalf("pin metadata: %s", out)
				}
				if tc.unpinned && (strings.Count(stderr, "\n") != 1 || !strings.Contains(stderr, "unpinned")) {
					t.Fatalf("unpinned diagnostic: %q", stderr)
				}
			}
			t.Logf("exit=%d stdout=%s stderr=%s", code, out, stderr)
		})
	}
}

func TestEveryRepositoryCommandChecksPinBeforeExecution(t *testing.T) {
	for _, command := range []string{"grype", "green", "clockfuse", "route"} {
		root := t.TempDir()
		code, out, stderr := invoke(t, command, "--repo", root, "--json")
		if code != 2 || out != "" || strings.Count(stderr, "\n") != 1 || !strings.Contains(stderr, "declares no LICTOR_VERSION") {
			t.Fatalf("%s: %d %q %q", command, code, out, stderr)
		}
		declarePin(t, root, "v99.0.0")
		code, out, stderr = invoke(t, command, "--repo", root, "--json", "--unpinned")
		if code != 2 || out != "" || !strings.Contains(stderr, "declares LICTOR_VERSION v99.0.0") {
			t.Fatalf("%s bypassed mismatch: %d %q %q", command, code, out, stderr)
		}
	}
}

func TestPinExactFirstMatchAndUnreadableInput(t *testing.T) {
	root := t.TempDir()
	declarePin(t, root, version)
	path := filepath.Join(root, ".github/workflows/ci.yml")
	body := "  # LICTOR_VERSION: v99.0.0\n    LICTOR_VERSION: v99.0.0\n  LICTOR_VERSION: " + version + " # first\n  LICTOR_VERSION: v99.0.0\n"
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := declaredVersion(root, "LICTOR_VERSION")
	if err != nil || v == nil || *v != version {
		t.Fatalf("first exact match: %v %v", v, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	code, out, stderr := invoke(t, "green", "--repo", root, "--json", "--unpinned")
	if code != 2 || out != "" || !strings.Contains(stderr, "cannot read") {
		t.Fatalf("unreadable declaration: %d %q %q", code, out, stderr)
	}
}
