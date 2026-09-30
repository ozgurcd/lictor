package grype

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/ozgurcd/lictor/internal/executor"
)

// runSBOM never infers its subject from Grype's source: that field describes
// the SBOM's original source, not the SBOM file. Grype 0.119.0 provides no input
// SBOM content digest, so a saved report cannot establish the requested binding.
func runSBOM(ctx context.Context, opts Options, out, errOut io.Writer) int {
	subject := Subject{Kind: "sbom", Target: filepath.Base(opts.SBOM), SHA256: "unavailable"}
	// Keep even refusal evidence to one line without disclosing a home path.
	subject.Target = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return '_'
		}
		return r
	}, subject.Target)
	finish := func(line string, code int) int {
		fmt.Fprintln(out, line+" — subject "+subject.Label()+"; config: not applicable (sbom subject: no directory declaration to compare); coverage: not applicable (sbom subject: no gitignored tree to enumerate)")
		return code
	}
	refuse := func(reason string) int { return finish("CANNOT-EVALUATE: grype-gate: "+reason, 2) }
	// Stat first so a named pipe cannot block the bounded regular-file reader.
	st, err := os.Stat(opts.SBOM)
	if err != nil || !st.Mode().IsRegular() || st.Size() > executor.ReportLimit {
		return refuse("SBOM must be a readable regular file no larger than 64 MiB")
	}
	raw, err := executor.ReadReport(opts.SBOM)
	if err != nil {
		return refuse("cannot read SBOM")
	}
	subject.SHA256 = fmt.Sprintf("%x", sha256.Sum256(raw))
	if opts.AsOf.IsZero() {
		return refuse("evaluation time is required")
	}
	if opts.Scan != "" {
		return refuse("SBOM replay has no verifiable report-to-SBOM content binding; run without -scan")
	}
	if opts.CoverageOnly || opts.Inventory != "" {
		return refuse("SBOM subjects do not admit -coverage-only or -inventory")
	}
	if !sbomFormat(raw) {
		return refuse("input is not an SPDX or CycloneDX SBOM")
	}

	tmp, err := os.MkdirTemp("", "lictor-sbom-")
	if err != nil {
		return refuse("cannot create SBOM scan scratch directory")
	}
	defer os.RemoveAll(tmp)
	input, report := filepath.Join(tmp, "input.sbom"), filepath.Join(tmp, "report.json")
	if err := os.WriteFile(input, raw, 0600); err != nil {
		return refuse("cannot snapshot SBOM")
	}
	if err := executor.ScanSBOM(ctx, opts.Repository, input, report, errOut); err != nil {
		return refuse("grype could not scan the SBOM; a scanner that cannot run is never a pass")
	}
	raw, err = executor.ReadReport(report)
	if err != nil {
		return refuse("grype wrote no readable bounded JSON report")
	}
	// A valid empty matches array is evidence; a missing/null report is not.
	var envelope struct {
		Matches json.RawMessage `json:"matches"`
	}
	if json.Unmarshal(raw, &envelope) != nil || len(envelope.Matches) == 0 || envelope.Matches[0] != '[' {
		return refuse("grype report has no matches array")
	}
	doc, err := ParseDoc(raw)
	if err != nil {
		return refuse("grype report is malformed")
	}
	allow, err := readAllowlist(opts.Allowlist)
	if err != nil {
		// Detailed input diagnostics belong on stderr, never in gate evidence.
		fmt.Fprintln(errOut, err)
		return refuse("allowlist is unreadable or malformed")
	}
	_, summary, ok := Decide(doc, allow)
	if !ok {
		return finish(summary, 1)
	}
	return finish(summary, 0)
}

// sbomFormat restricts the admitted format families, not their semantics.
// Grype's parser remains authoritative; this recognition alone never passes.
func sbomFormat(raw []byte) bool {
	var header struct {
		SPDXVersion string `json:"spdxVersion"`
		BOMFormat   string `json:"bomFormat"`
	}
	if json.Unmarshal(raw, &header) == nil {
		return strings.HasPrefix(header.SPDXVersion, "SPDX-") || header.BOMFormat == "CycloneDX"
	}
	decoder := xml.NewDecoder(bytes.NewReader(raw))
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		if start, ok := token.(xml.StartElement); ok {
			return start.Name.Local == "bom" && strings.HasPrefix(start.Name.Space, "http://cyclonedx.org/schema/bom/")
		}
	}
	for _, line := range bytes.Split(raw, []byte("\n")) {
		if strings.HasPrefix(strings.TrimSpace(string(line)), "SPDXVersion: SPDX-") {
			return true
		}
	}
	return false
}
