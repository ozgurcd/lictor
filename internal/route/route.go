// Package route selects Identuum installation policy. Achta alone judges YAML.
package route

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ozgurcd/lictor/internal/executor"
)

type Tools interface {
	Version(context.Context, string, io.Writer) ([]byte, error)
	Check(context.Context, string, executor.RoutePolicy, bool, io.Writer) ([]byte, error)
}
type Check struct {
	Key      string          `json:"key"`
	ExitCode int             `json:"exit_code"`
	Document json.RawMessage `json:"document"`
}
type Set struct {
	Tool    string  `json:"tool"`
	Skipped bool    `json:"skipped"`
	Reason  string  `json:"reason"`
	Checks  []Check `json:"checks"`
}
type Result struct {
	SchemaVersion        string  `json:"schema_version"`
	Subject              string  `json:"subject"`
	Outcome              string  `json:"outcome"`
	Reason               string  `json:"reason"`
	AchtaVersion         string  `json:"achta_version"`
	DeclaredAchtaVersion *string `json:"declared_achta_version"`
	Sets                 []Set   `json:"sets"`
	ExitCode             int     `json:"exit_code"`
	Human                string  `json:"-"`
}

func Policies(tool string) []executor.RoutePolicy {
	if tool == "rulefloor" {
		return []executor.RoutePolicy{{Pattern: `rulefloor/archive/refs/tags/[$][{]RULEFLOOR_VERSION[}]`, Key: "RULEFLOOR_VERSION", Bans: []string{
			`brew install[^#]*rulefloor`, `go install[^#]*rulefloor`, `rulefloor/archive/refs/tags/v[0-9]`,
			`-X main[.]version=[$][{]RULEFLOOR_VERSION[^}]`, `-X main[.]version=[$][{][^R]`, `-X main[.]version=[$][^{]`, `-X main[.]version=[^$]`,
		}}}
	}
	if tool == "lictor" {
		var policies []executor.RoutePolicy
		for _, key := range []string{"LICTOR_VERSION", "LICTOR_SHA256"} {
			policies = append(policies, executor.RoutePolicy{Pattern: `lictor/releases/download/[$][{]LICTOR_VERSION[}]/lictor_[$][{]LICTOR_VERSION#v[}]_`, Bans: []string{`brew install[^#]*lictor`, `go install[^#]*lictor`, `lictor/releases/download/v[0-9]`}, Key: key})
		}
		return policies
	}
	return nil
}

func Selected(root string) (map[string]bool, error) {
	repo, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("workflow repository: %w", err)
	}
	defer repo.Close()
	entries, err := fs.ReadDir(repo.FS(), ".github/workflows")
	if err != nil {
		return nil, fmt.Errorf("workflow directory: %w", err)
	}
	selected := map[string]bool{}
	for _, entry := range entries {
		ext := filepath.Ext(entry.Name())
		if entry.IsDir() || (ext != ".yml" && ext != ".yaml") {
			continue
		}
		f, err := repo.Open(filepath.Join(".github/workflows", entry.Name()))
		if err != nil {
			return nil, err
		}
		st, err := f.Stat()
		if err != nil || !st.Mode().IsRegular() {
			f.Close()
			return nil, fmt.Errorf("workflow %s is not a regular readable file", entry.Name())
		}
		b, err := io.ReadAll(io.LimitReader(f, 1<<20+1))
		f.Close()
		if err != nil {
			return nil, err
		}
		if len(b) > 1<<20 {
			return nil, fmt.Errorf("workflow %s exceeds 1 MiB", entry.Name())
		}
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.ToLower(strings.TrimSpace(line))
			if strings.HasPrefix(line, "#") {
				continue
			}
			for _, tool := range []string{"rulefloor", "lictor"} {
				if !strings.Contains(line, tool) {
					continue
				}
				for _, marker := range []string{"install", "releases/download", "archive/refs/tags", "brew", "go install"} {
					if strings.Contains(line, marker) {
						selected[tool] = true
					}
				}
			}
		}
	}
	return selected, nil
}

func Run(ctx context.Context, root string, declared *string, wantJSON bool, tools Tools, stderr io.Writer) Result {
	r := Result{SchemaVersion: "lictor.route.v1", Subject: filepath.Base(root), Outcome: "cannot_evaluate", ExitCode: 2, DeclaredAchtaVersion: declared, Sets: []Set{}}
	refuse := func(reason string) Result {
		r.ExitCode = 2
		r.Outcome = "cannot_evaluate"
		r.Reason = "CANNOT-EVALUATE: " + reason
		r.Human += r.Reason + "\n"
		return r
	}
	selected, err := Selected(root)
	if err != nil {
		return refuse(err.Error())
	}
	b, err := tools.Version(ctx, root, stderr)
	if err != nil {
		return refuse("achta version could not run: " + err.Error())
	}
	var v struct {
		Schema    string `json:"schema_version"`
		Version   string `json:"version"`
		Agreement string `json:"version_agreement"`
	}
	if err := json.Unmarshal(b, &v); err != nil || v.Schema != "achta.version.v1" || v.Version == "" || v.Agreement != "pass" {
		return refuse("achta version response is not an agreed achta.version.v1")
	}
	r.AchtaVersion = v.Version
	if declared != nil {
		fmt.Fprintf(stderr, "achta: installed %s; declared %s\n", v.Version, *declared)
		if *declared != v.Version {
			return refuse("achta installed version differs from declared ACHTA_VERSION")
		}
	} else {
		fmt.Fprintf(stderr, "achta: installed %s; ACHTA_VERSION not declared\n", v.Version)
	}
	r.ExitCode = 0
	for _, tool := range []string{"rulefloor", "lictor"} {
		set := Set{Tool: tool, Checks: []Check{}}
		if !selected[tool] {
			set.Skipped = true
			set.Reason = "no install line"
			r.Human += tool + " set skipped: no install line\n"
			r.Sets = append(r.Sets, set)
			continue
		}
		set.Reason = "install line selects Achta judgement"
		for _, policy := range Policies(tool) {
			b, err := tools.Check(ctx, root, policy, wantJSON, stderr)
			code := 0
			if err != nil {
				var exit *exec.ExitError
				if !errors.As(err, &exit) {
					return refuse("achta declared-route could not run: " + err.Error())
				}
				code = exit.ExitCode()
				if code < 1 || code > 2 {
					return refuse(fmt.Sprintf("achta returned unsupported exit %d", code))
				}
			}
			check := Check{Key: policy.Key, ExitCode: code}
			if wantJSON {
				if !json.Valid(b) {
					return refuse("achta returned invalid JSON")
				}
				if code != 2 {
					var native struct {
						Schema string `json:"schema_version"`
						Status string `json:"status"`
					}
					if err := json.Unmarshal(b, &native); err != nil || native.Schema != "achta.declared-route-check.v1" || native.Status != []string{"pass", "fail"}[code] {
						return refuse("achta returned an incompatible or exit-inconsistent declared-route document")
					}
				}
				check.Document = append(json.RawMessage(nil), b...)
			} else {
				r.Human += string(b)
			}
			set.Checks = append(set.Checks, check)
			if code > r.ExitCode {
				r.ExitCode = code
			}
		}
		r.Sets = append(r.Sets, set)
	}
	r.Outcome = []string{"pass", "fail", "cannot_evaluate"}[r.ExitCode]
	r.Reason = "Achta installation-route judgements; tools without install lines are explicitly skipped"
	return r
}
