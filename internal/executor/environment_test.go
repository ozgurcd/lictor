package executor

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestEnvironmentControlledNames(t *testing.T) {
	names := append(append([]string{}, goEnvironmentNames...), recordEnvironmentNames...)
	names = append(names, "GOFLAGS", "GOWORK", "GOENV", "GOFUTURE", "GIT_OPTIONAL_LOCKS", "LC_ALL")
	for _, name := range names {
		if _, err := DeclareEnvironment([]string{name}); err == nil || !strings.Contains(err.Error(), name) {
			t.Errorf("controlled name %s was not refused by name", name)
		}
	}
}

func TestEnvironmentStreamSecrecy(t *testing.T) {
	value := "postgres://fixture:" + strings.Repeat("synthetic", 5) + "@localhost/fixture"
	input := "check OK: " + value + "\ncheck FAILED: " + value + "\n"
	for split := 0; split <= len(input); split++ {
		var out bytes.Buffer
		w := (&GateEnvironment{values: []string{value}}).writer(&out)
		if _, err := w.Write([]byte(input[:split])); err != nil {
			t.Fatal("first write failed")
		}
		if _, err := w.Write([]byte(input[split:])); err != nil || w.flush() != nil {
			t.Fatal("second write or flush failed")
		}
		if out.String() != "check OK: [redacted]\ncheck FAILED: [redacted]\n" {
			t.Fatalf("unsafe or incorrect stream at split %d", split)
		}
	}
	var out bytes.Buffer
	w := (&GateEnvironment{values: []string{value}}).writer(&out)
	for i := range input {
		if _, err := w.Write([]byte{input[i]}); err != nil {
			t.Fatal("single-byte write failed")
		}
		if bytes.Contains(out.Bytes(), []byte(value)) {
			t.Fatal("value leaked before flush")
		}
	}
	if w.flush() != nil || strings.Count(out.String(), "[redacted]") != 2 {
		t.Fatal("single-byte redaction failed")
	}
	// A recorder ceiling may end inside a value. No retained prefix escapes.
	out.Reset()
	w = (&GateEnvironment{values: []string{value}}).writer(&out)
	_, err := w.Write([]byte(value[:len(value)-1]))
	if err != nil || w.flush() != nil || out.String() != "[redacted]" {
		t.Fatal("truncated prefix escaped")
	}
	// Sorted overlapping values must not expose the longer value's suffix.
	out.Reset()
	t.Setenv("LICTOR_TEST_LONG", "abcde")
	t.Setenv("LICTOR_TEST_SHORT", "abc")
	g, err := DeclareEnvironment([]string{"LICTOR_TEST_SHORT", "LICTOR_TEST_LONG"})
	if err != nil {
		t.Fatal("declaration failed")
	}
	w = g.writer(&out)
	_, _ = w.Write([]byte("abc"))
	_, _ = w.Write([]byte("de!"))
	if w.flush() != nil || out.String() != "[redacted]!" {
		t.Fatal("overlapping value escaped")
	}
	t.Log("every two-chunk split, single-byte writes, truncated prefixes and overlapping values remain secret")
}

type failingEnvironmentWriter struct{}

func (failingEnvironmentWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestEnvironmentOutputFailure(t *testing.T) {
	w := (&GateEnvironment{}).writer(failingEnvironmentWriter{})
	if _, err := w.Write([]byte("ordinary")); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("output failure lost")
	}
}
