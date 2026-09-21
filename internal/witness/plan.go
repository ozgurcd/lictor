package witness

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type Entry struct {
	Name string
	Argv []string
}

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Parse validates the entire plan before any record or process is opened.
func Parse(raw, requires []string, namesOnly bool) ([]Entry, map[string][]string, error) {
	if len(raw) == 0 {
		return nil, nil, fmt.Errorf("verify-all: empty plan")
	}
	var entries []Entry
	positions := map[string]int{}
	for i, s := range raw {
		n, cmd, found := strings.Cut(s, "=")
		_, duplicate := positions[n]
		if !namePattern.MatchString(n) || duplicate || (!found && !namesOnly) {
			return nil, nil, fmt.Errorf("verify-all: invalid or duplicate plan entry: %s", n)
		}
		var argv []string
		if found {
			var err error
			argv, err = Words(cmd)
			if err != nil {
				return nil, nil, fmt.Errorf("target %s: %w", n, err)
			}
			if len(argv) == 0 || argv[0] == "" {
				return nil, nil, fmt.Errorf("target %s: empty argv", n)
			}
			switch filepath.Base(argv[0]) {
			case "sh", "bash", "dash", "zsh", "ksh", "fish":
				return nil, nil, fmt.Errorf("target %s: shell executable %q refused; put the step in a make target", n, argv[0])
			}
		}
		entries = append(entries, Entry{n, argv})
		positions[n] = i
	}
	deps := map[string][]string{}
	for _, s := range requires {
		n, p, ok := strings.Cut(s, ":")
		i, exists := positions[n]
		if !exists {
			return nil, nil, fmt.Errorf("verify-all: unknown dependent %s", n)
		}
		j, exists := positions[p]
		if !ok || !exists || j >= i {
			return nil, nil, fmt.Errorf("verify-all: %s requires an earlier planned target: %s", n, p)
		}
		deps[n] = append(deps[n], p)
	}
	return entries, deps, nil
}

// Words is POSIX quote removal, without expansions or shell evaluation.
func Words(s string) ([]string, error) {
	var out []string
	var word strings.Builder
	var quote byte
	started := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if quote == '\'' {
			if c == '\'' {
				quote = 0
			} else {
				word.WriteByte(c)
			}
			continue
		}
		if c == '\\' {
			if i+1 == len(s) {
				return nil, fmt.Errorf("unfinished escape")
			}
			next := s[i+1]
			if quote == '"' && !strings.ContainsRune("$`\"\\\n", rune(next)) {
				word.WriteByte(c)
				continue
			}
			i++
			if next != '\n' {
				word.WriteByte(next)
				started = true
			}
			continue
		}
		if quote == '"' {
			if c == '"' {
				quote = 0
			} else {
				word.WriteByte(c)
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			started = true
			continue
		}
		if strings.ContainsRune("|&;<>$`()\n", rune(c)) {
			return nil, fmt.Errorf("unquoted shell metacharacter %q refused", c)
		}
		if c == ' ' || c == '\t' {
			if started {
				out = append(out, word.String())
				word.Reset()
				started = false
			}
			continue
		}
		word.WriteByte(c)
		started = true
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated %c quote", quote)
	}
	if started {
		out = append(out, word.String())
	}
	return out, nil
}
