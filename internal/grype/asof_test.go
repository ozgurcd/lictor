package grype

import (
	"strings"
	"testing"
)

func TestAsOfOverridesSuppressionDate(t *testing.T) {
	root := subjectRepo(t, false)
	scan := writeScan(t, report(dirSource(root), appliedConfig(root)))
	inventory := writeInventory(t, "/go.mod")
	for _, tc := range []struct {
		asOf string
		code int
		text string
	}{
		{"2026-09-16T00:00:00Z", 0, "none lapsed on 2026-09-16"},
		{"2099-01-02T00:00:00Z", 1, "has lapsed, today is 2099-01-02"},
	} {
		code, line := judge(t, "-scan", scan, "-inventory", inventory, "--as-of", tc.asOf)
		if code != tc.code || !strings.Contains(line, tc.text) {
			t.Errorf("as-of %s: got exit %d, %q; want exit %d containing %q", tc.asOf, code, line, tc.code, tc.text)
		}
	}
}
