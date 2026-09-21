package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGreenCommandContract(t *testing.T) {
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOSUMDB", "off")
	t.Setenv("GOTOOLCHAIN", "local")
	t.Setenv("GOWORK", "off")
	t.Setenv("GOENV", "off")
	t.Setenv("GOFLAGS", "")
	root := t.TempDir()
	for _, tc := range []struct {
		name, source, outcome, cause string
		code                         int
	}{
		{"nomod", "", "CANNOT-EVALUATE", "not a Go module", 2},
		{"green", "package fixture\n\nfunc Value() int { return 1 }\n", "GREEN", "builds, vets and tests clean", 0},
		{"red", "package fixture\n\nfunc Value() int { return missing }\n", "NOT-GREEN", "go build FAILED", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.source != "" {
				for name, body := range map[string]string{"go.mod": "module example.com/fixture\n\ngo 1.21\n", "fixture.go": tc.source} {
					if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			code, human, diagnostic := invoke(t, "green", "--repo", root)
			if code != tc.code || !strings.HasPrefix(human, tc.outcome+":") || !strings.Contains(human, tc.cause) {
				t.Fatalf("exit=%d evidence=%q stderr=%q", code, human, diagnostic)
			}
			jcode, document, _ := invoke(t, "green", "--repo", root, "--json")
			if jcode != tc.code {
				t.Fatalf("JSON exit=%d: %s", jcode, document)
			}
			validateDocument(t, "lictor.green.v1", document)
			var got struct {
				Outcome, Cause, Subject string
				GoVersion               string `json:"go_version"`
				Excerpt                 string
			}
			if err := json.Unmarshal([]byte(document), &got); err != nil {
				t.Fatal(err)
			}
			if got.Outcome != tc.outcome || got.Cause != tc.cause || got.Subject != filepath.Base(root) || got.GoVersion == "" {
				t.Fatalf("contract: %s", document)
			}
			if strings.Contains(human, root) || !strings.Contains(human, got.GoVersion) {
				t.Fatalf("identity: %s", human)
			}
			if strings.TrimSpace(diagnostic) != strings.TrimSpace(got.Excerpt) {
				t.Fatalf("excerpt mismatch: %q / %q", diagnostic, got.Excerpt)
			}
		})
	}
}
