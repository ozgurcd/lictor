package clockfuse

import (
	"strings"
	"testing"
)

func TestAdversarialClockfuse(t *testing.T) {
	for _, tc := range []struct{ name, raw, want string }{
		{"colon-path", "a:b_test.go:12: Foo{...} omits `Now` -> falls back\n", "1|a:b_test.go:12:"},
		{"dotted-type", "a_test.go:12: pkg.Foo{...} omits `Now` -> falls back\n", "1|a_test.go|pkg.Foo|Now"},
		{"no-fallback", "a_test.go:12: Foo{...} omits `Now`\n", "1|a_test.go|Foo|Now"},
		{"non-ascii-space-path", "a\u00a0b_test.go:12: Foo{...} omits `Now`\n", "1|a\u00a0b_test.go|Foo|Now"},
		{"carriage-return-path", "a\rb_test.go:12: Foo{...} omits `Now`\n", "1|a\rb_test.go|Foo|Now"},
		{"vertical-tab-path", "a\vb_test.go:12: Foo{...} omits `Now`\n", "1|a\vb_test.go|Foo|Now"},
		{"form-feed-path", "a\fb_test.go:12: Foo{...} omits `Now`\n", "1|a\fb_test.go|Foo|Now"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := strings.Join(Normalize(tc.raw), "\n")
			t.Logf("normalized=%q", got)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
			if reason, code := Compare(Normalize(tc.raw), tc.want+"\n", "<repo>"); code != 0 {
				t.Fatalf("unchanged source snapshot failed: %d %s", code, reason)
			}
		})
	}
	t.Run("overflow-snapshot", func(t *testing.T) {
		reason, code := Compare([]string{"1|a_test.go|Foo|Now"}, "999999999999999999999999999999|a_test.go|Foo|Now\n", "<repo>")
		t.Logf("exit=%d reason=%q", code, reason)
		if code != 2 || reason != "CANNOT-EVALUATE: malformed snapshot count for a_test.go|Foo|Now" {
			t.Fatalf("overflow passed: %d %s", code, reason)
		}
	})
}
