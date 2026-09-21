package witness

import (
	"strings"
	"testing"
)

func TestAdversarialPlan(t *testing.T) {
	for _, tc := range []struct {
		name      string
		raw       []string
		wantError string
	}{
		{"newline-name", []string{"bad\nname=echo ok"}, "invalid or duplicate"},
		{"nul-name", []string{"bad\x00name=echo ok"}, "invalid or duplicate"},
		{"no-equals", []string{"echo"}, "invalid or duplicate"},
		{"empty-argv", []string{"test="}, "empty argv"},
		{"unbalanced", []string{"test=echo 'x"}, "unterminated"},
		{"one-MiB", []string{"test=echo " + strings.Repeat("x", (1<<20)-10)}, ""},
		{"case-distinct", []string{"Test=echo one", "test=echo two"}, ""},
		{"exact-duplicate", []string{"test=echo one", "test=echo two"}, "invalid or duplicate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entries, _, err := Parse(tc.raw, nil, false)
			t.Logf("entries=%d input-bytes=%d error=%v", len(entries), len(tc.raw[0]), err)
			if (tc.wantError == "" && err != nil) || (tc.wantError != "" && (err == nil || !strings.Contains(err.Error(), tc.wantError))) {
				t.Fatalf("unexpected parse: %v", err)
			}
			if err == nil && len(entries) != len(tc.raw) {
				t.Fatal("entry lost")
			}
		})
	}
}

func TestAdversarialEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, raw, want string
		packages        int
	}{
		{"prose", "ok this is prose\n", "", 0},
		{"prose-shaped-like-package", "ok everything 1s later was wrong\n", "", 1},
		{"ansi", "\x1b[32mcheck OK: colored\x1b[0m\n\x1b[32mok pkg 0.1s\x1b[0m\n", "", 0},
		{"crlf", "check OK: yes\r\nok pkg 0.1s\r\n", "check OK: yes\r\n", 1},
		{"invalid-utf8", "check OK: \xff\nok pkg\xff 0.1s\n", "check OK: \xff\n", 1},
		{"nul-matcher-only-recorder-filters-first", "check OK: \x00\nok pkg 0.1s\n", "check OK: \x00\n", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got strings.Builder
			count := 0
			err := lines(strings.NewReader(tc.raw), func(s string) error {
				if evidence.MatchString(s) {
					got.WriteString(s + "\n")
				}
				if packages.MatchString(s) {
					count++
				}
				return nil
			})
			t.Logf("evidence=%q packages=%d error=%v", got.String(), count, err)
			if err != nil || got.String() != tc.want || count != tc.packages {
				t.Fatalf("evidence=%q packages=%d err=%v", got.String(), count, err)
			}
		})
	}
}
