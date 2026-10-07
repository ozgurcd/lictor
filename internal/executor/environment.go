package executor

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

var recordEnvironmentNames = []string{"GRYPE_DB_CACHE_DIR", "GRYPE_DB_AUTO_UPDATE", "GRYPE_CHECK_FOR_APP_UPDATE"}
var environmentName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// GateEnvironment captures only explicitly declared names, once per run.
// Its values are private and never serialized or included in an error.
type GateEnvironment struct {
	names, entries, values []string
	present                map[string]bool
	short                  map[string]bool
}

// DeclareEnvironment refuses collisions before reading any declared value.
func DeclareEnvironment(names []string) (*GateEnvironment, error) {
	reserved := map[string]bool{"GIT_OPTIONAL_LOCKS": true, "LC_ALL": true}
	for _, name := range append(append([]string{}, goEnvironmentNames...), recordEnvironmentNames...) {
		reserved[name] = true
	}
	seen := make(map[string]bool)
	for _, name := range names {
		if !environmentName.MatchString(name) {
			// An accidental NAME=value must never echo the supplied value.
			label, _, _ := strings.Cut(name, "=")
			return nil, fmt.Errorf("environment name %q refused: expected a name only ([A-Za-z_][A-Za-z0-9_]*)", label)
		}
		if reserved[name] || strings.HasPrefix(name, "GO") {
			return nil, fmt.Errorf("environment name %q refused: controlled by lictor", name)
		}
		if seen[name] {
			return nil, fmt.Errorf("environment name %q refused: duplicate", name)
		}
		seen[name] = true
	}
	g := &GateEnvironment{names: append([]string(nil), names...), present: make(map[string]bool), short: make(map[string]bool)}
	for _, name := range names {
		value, present := os.LookupEnv(name)
		g.present[name] = present
		if present {
			g.entries = append(g.entries, name+"="+value)
			g.short[name] = len(value) < 8
			if !g.short[name] {
				g.values = append(g.values, value)
			}
		}
	}
	// Longest first keeps overlapping values from exposing a longer suffix.
	sort.Slice(g.values, func(i, j int) bool { return len(g.values[i]) > len(g.values[j]) })
	return g, nil
}

func (g *GateEnvironment) WritePresence(w io.Writer) {
	for _, name := range g.names {
		state := "absent"
		if g.present[name] {
			state = "present"
			if g.short[name] {
				state += " (short value, not redacted)"
			}
		}
		fmt.Fprintf(w, "environment: %s %s\n", name, state)
	}
}

func (g *GateEnvironment) writer(dst io.Writer) *environmentWriter {
	return &environmentWriter{dst: dst, values: g.values}
}

// environmentWriter withholds any suffix that could start a declared value.
// Thus neither process write boundaries nor the output ceiling expose it.
type environmentWriter struct {
	dst     io.Writer
	values  []string
	pending string
}

func (w *environmentWriter) Write(p []byte) (int, error) {
	w.pending += string(p)
	var b strings.Builder
	i := 0
	for i < len(w.pending) {
		rest := w.pending[i:]
		matched := false
		for _, value := range w.values {
			if len(rest) < len(value) && strings.HasPrefix(value, rest) {
				w.pending = rest
				_, err := io.WriteString(w.dst, b.String())
				return len(p), err
			}
			if strings.HasPrefix(rest, value) {
				b.WriteString("[redacted]")
				i += len(value)
				matched = true
				break
			}
		}
		if !matched {
			b.WriteByte(w.pending[i])
			i++
		}
	}
	w.pending = ""
	_, err := io.WriteString(w.dst, b.String())
	return len(p), err
}

func (w *environmentWriter) flush() error {
	if w.pending == "" {
		return nil
	}
	// A final prefix can be a value cut by the recorder ceiling. Hide it too.
	w.pending = ""
	_, err := io.WriteString(w.dst, "[redacted]")
	return err
}
