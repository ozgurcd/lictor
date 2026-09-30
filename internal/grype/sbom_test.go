package grype

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ozgurcd/lictor/internal/executor"
)

// The test executable acts as an offline scanner, checking the actual argv and
// snapshot bytes. Real Syft/Grype provenance is recorded in docs/sbom.md.
func TestMain(m *testing.M) {
	if filepath.Base(os.Args[0]) == "grype" {
		if len(os.Args) != 4 || !strings.HasPrefix(os.Args[1], "sbom:") || os.Args[2] != "--output" || !strings.HasPrefix(os.Args[3], "json=") {
			os.Exit(10)
		}
		b, err := os.ReadFile(strings.TrimPrefix(os.Args[1], "sbom:"))
		if err != nil {
			os.Exit(11)
		}
		for _, name := range []string{"high.spdx.json", "clean.cdx.json"} {
			fixture, err := os.ReadFile(filepath.Join("testdata", name))
			if err != nil {
				os.Exit(12)
			}
			if bytes.Equal(b, fixture) {
				report, err := os.ReadFile(filepath.Join("testdata", strings.Split(name, ".")[0]+".report.json"))
				if err != nil {
					os.Exit(13)
				}
				if os.WriteFile(strings.TrimPrefix(os.Args[3], "json="), report, 0600) != nil {
					os.Exit(14)
				}
				os.Exit(0)
			}
		}
		os.Exit(15) // An unparseable input never produces a clean report.
	}
	os.Exit(m.Run())
}

func sbomOptions(t *testing.T, name string) Options {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	o := Options{Repository: root, Allowlist: filepath.Join(t.TempDir(), "absent.json"), AsOf: time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)}
	field := reflect.ValueOf(&o).Elem().FieldByName("SBOM")
	if !field.IsValid() {
		t.Fatal("explicit SBOM request is unavailable")
	}
	field.SetString(filepath.Join(root, "testdata", name))
	return o
}

func sbomScanner(t *testing.T) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Symlink(exe, filepath.Join(dir, "grype")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
}

// RULE: GRYPE-SBOM-1
func TestRule_GRYPE_SBOM_1(t *testing.T) {
	sbomScanner(t)
	for _, tc := range []struct {
		name, allow string
		code        int
	}{
		{"high.spdx.json", "", 1}, {"clean.cdx.json", "", 0},
		{"high.spdx.json", "high.allowlist.json", 1}, {"not-sbom.json", "", 2},
	} {
		t.Run(tc.name+tc.allow, func(t *testing.T) {
			o := sbomOptions(t, tc.name)
			if tc.allow != "" {
				o.Allowlist = filepath.Join(o.Repository, "testdata", tc.allow)
			}
			var out, diagnostic bytes.Buffer
			code := Run(context.Background(), o, &out, &diagnostic)
			if code != tc.code {
				t.Fatalf("exit=%d want=%d: %s %s", code, tc.code, out.String(), diagnostic.String())
			}
			b, err := os.ReadFile(filepath.Join(o.Repository, "testdata", tc.name))
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"subject sbom:" + tc.name, fmt.Sprintf("sha256:%x", sha256.Sum256(b)), "config: not applicable", "coverage: not applicable"} {
				if !strings.Contains(out.String(), want) {
					t.Errorf("missing %q: %s", want, out.String())
				}
			}
			if strings.Contains(out.String(), o.Repository) {
				t.Errorf("absolute subject path leaked: %s", out.String())
			}
			if tc.code == 1 && !strings.Contains(out.String(), "GHSA-35jh-r3h4-6jhm") {
				t.Errorf("High finding lost: %s", out.String())
			}
			t.Log(strings.TrimSpace(out.String()))
		})
	}
	t.Run("replay refuses unbound bytes", func(t *testing.T) {
		o := sbomOptions(t, "clean.cdx.json")
		o.Scan = "testdata/clean.report.json"
		var out bytes.Buffer
		if code := Run(context.Background(), o, &out, &out); code != 2 || !strings.Contains(out.String(), "binding") {
			t.Fatalf("unbound replay: %d %s", code, out.String())
		}
	})
	t.Run("missing scanner never passes", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		o := sbomOptions(t, "clean.cdx.json")
		var out bytes.Buffer
		if code := Run(context.Background(), o, &out, &out); code != 2 {
			t.Fatalf("missing scanner: %d %s", code, out.String())
		}
	})
	for _, name := range []string{"missing", "directory", "oversize", "bad-spdx", "bad-cdx", "bad-xml", "coverage", "inventory", "no-time", "cancelled"} {
		t.Run(name, func(t *testing.T) {
			o := sbomOptions(t, "clean.cdx.json")
			ctx := context.Background()
			switch name {
			case "missing":
				o.SBOM = filepath.Join(t.TempDir(), "missing")
			case "directory":
				o.SBOM = t.TempDir()
			case "oversize":
				o.SBOM = filepath.Join(t.TempDir(), "large.json")
				f, err := os.Create(o.SBOM)
				if err != nil {
					t.Fatal(err)
				}
				if err := f.Truncate(executor.ReportLimit + 1); err != nil {
					t.Fatal(err)
				}
				if err := f.Close(); err != nil {
					t.Fatal(err)
				}
			case "bad-spdx", "bad-cdx", "bad-xml":
				o.SBOM = filepath.Join(t.TempDir(), "bad.json")
				body := map[string]string{"bad-spdx": `{"spdxVersion":"SPDX-2.3"}`, "bad-cdx": `{"bomFormat":"CycloneDX"}`, "bad-xml": `<bom xmlns="http://cyclonedx.org/schema/bom/1.6"><broken>`}[name]
				if err := os.WriteFile(o.SBOM, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			case "coverage":
				o.CoverageOnly = true
			case "inventory":
				o.Inventory = "irrelevant.json"
			case "no-time":
				o.AsOf = time.Time{}
			case "cancelled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			var out bytes.Buffer
			if code := Run(ctx, o, &out, &out); code != 2 {
				t.Fatalf("refusal=%d: %s", code, out.String())
			}
			if !strings.Contains(out.String(), "config: not applicable") || !strings.Contains(out.String(), "coverage: not applicable") {
				t.Errorf("missing predicate evidence: %s", out.String())
			}
			if strings.Contains(out.String(), filepath.Dir(o.SBOM)) {
				t.Errorf("input path leaked: %s", out.String())
			}
		})
	}
}

func TestSBOMFormatFamilies(t *testing.T) {
	for _, raw := range []string{`{"spdxVersion":"SPDX-2.3"}`, `{"bomFormat":"CycloneDX"}`, `<?xml version="1.0"?><bom xmlns="http://cyclonedx.org/schema/bom/1.6"/>`, "SPDXVersion: SPDX-2.3\n"} {
		if !sbomFormat([]byte(raw)) {
			t.Errorf("format not recognized: %q", raw)
		}
	}
	for _, raw := range []string{`{}`, `{"artifacts":[]}`, `<bom/>`, `random text`} {
		if sbomFormat([]byte(raw)) {
			t.Errorf("non-SPDX/CycloneDX recognized: %q", raw)
		}
	}
}

func TestSBOMVerdictParity(t *testing.T) {
	raw, err := os.ReadFile("testdata/high.report.json")
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Matches []json.RawMessage `json:"matches"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatal(err)
	}
	var matches []string
	for _, m := range d.Matches {
		matches = append(matches, string(m))
	}
	allow, err := filepath.Abs("testdata/high.allowlist.json")
	if err != nil {
		t.Fatal(err)
	}
	root := subjectRepo(t, false)
	for _, source := range []string{dirSource(root), imageSource} {
		scan := writeScan(t, report(source, appliedConfig(root), matches...))
		code, line := judge(t, "-scan", scan, "-inventory", writeInventory(t, "/go.mod"), "-allowlist", allow, "-repo", root)
		if code != 1 || !strings.Contains(line, "matches=5 fixable=0 severe=2") || !strings.Contains(line, "GHSA-35jh-r3h4-6jhm") {
			t.Fatalf("parity: %d %s", code, line)
		}
		t.Log(line)
	}
}
