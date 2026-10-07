package executor

import (
	"bytes"
	"testing"
)

func TestEnvironmentRedactionLength(t *testing.T) {
	for _, tc := range []struct{ name, value, input, want, presence string }{
		{"one", "1", "1.25s\nok 1\n", "1.25s\nok 1\n", "present (short value, not redacted)"},
		{"seven", "abcdefg", "abcdefg!", "abcdefg!", "present (short value, not redacted)"},
		{"eight", "abcdefgh", "abcdefgh!", "[redacted]!", "present"},
		{"eight-ceiling", "abcdefgh", "abcdefg", "[redacted]", "present"},
		{"empty", "", "ok 1\n", "ok 1\n", "present (short value, not redacted)"},
		{"eight-bytes", "éééé", "éééé!", "[redacted]!", "present"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("LICTOR_LENGTH_FIXTURE", tc.value)
			g, err := DeclareEnvironment([]string{"LICTOR_LENGTH_FIXTURE"})
			if err != nil {
				t.Fatal(err)
			}
			var record bytes.Buffer
			g.WritePresence(&record)
			if record.String() != "environment: LICTOR_LENGTH_FIXTURE "+tc.presence+"\n" {
				t.Errorf("wrong presence annotation: %q", record.String())
			}
			for split := 0; split <= len(tc.input); split++ {
				var out bytes.Buffer
				w := g.writer(&out)
				if _, err := w.Write([]byte(tc.input[:split])); err != nil {
					t.Fatal(err)
				}
				if _, err := w.Write([]byte(tc.input[split:])); err != nil {
					t.Fatal(err)
				}
				if err := w.flush(); err != nil {
					t.Fatal(err)
				}
				if out.String() != tc.want {
					t.Errorf("incorrect output at split %d", split)
				}
			}
		})
	}
}
