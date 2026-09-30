package main

import (
	"strings"
	"testing"
)

func TestSBOMCLI(t *testing.T) {
	for _, args := range [][]string{{"capabilities"}, {"capabilities", "--json"}, {"grype", "--help"}} {
		code, out, _ := invoke(t, args...)
		if code != 0 || !strings.Contains(out, "-sbom") {
			t.Fatalf("%v: %d %s", args, code, out)
		}
	}
	code, out, _ := invoke(t, "grype", "--repo", t.TempDir(), "--unpinned", "--sbom", "../../internal/grype/testdata/clean.cdx.json", "--scan", "unbound.json", "--as-of", "2026-09-30T00:00:00Z", "--json")
	if code != 2 || !strings.Contains(out, "content binding") || !strings.Contains(out, "sbom:clean.cdx.json") {
		t.Fatalf("replay: %d %s", code, out)
	}
	validateDocument(t, "lictor.grype.v1", out)
}
