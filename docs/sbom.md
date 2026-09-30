# SBOM subject measurement — 2026-09-30

Task: agent-lictor-sbom. Baseline: 217daa7, v0.4.2. Installed tools:
Grype 0.119.0 (embedded Syft v1.52.0), Syft 1.52.0.
Evaluation time for Lictor proofs: 2026-09-30T00:00:00Z.

## Real scanner measurement before implementation

A fixture directory held package.json and a lockfileVersion 3 package-lock.json
with lodash 4.17.20 in packages["node_modules/lodash"]. Its .grype.yaml declared
`exclude: ["./bin/**"]`. No package download or install was needed.
Commands, relative to the ignored .cache/sbom-measure scratch directory:

```text
syft scan dir:fixture -o cyclonedx-json=fixture.cdx.json -o spdx-json=fixture.spdx.json
```

Then, from the fixture directory:

```text
grype sbom:../fixture.spdx.json -o json=../scan.spdx.json
```

The report's source and relevant descriptor fields, quoted exactly as JSON
values (this is an excerpt, not the entire descriptor):

```json
{
  "source": {"type": "directory", "target": "fixture"},
  "descriptor": {
    "name": "grype",
    "version": "0.119.0",
    "configuration": {"exclude": ["./bin/**"]}
  }
}
```

A CycloneDX scan with `grype sbom:../fixture.cdx.json -o json=../scan.cdx.json`
on the initial empty catalog instead wrote:

```json
{"source":{"type":"file","target":"fixture"},"descriptor":{"name":"grype","version":"0.119.0","configuration":{"exclude":["./bin/**"]}}}
```

Thus the brief's source-type description reproduces for SPDX but is not
universal across formats: this CycloneDX source is file, not directory.
The first fixture contained only node_modules/lodash/package.json, and Syft
catalogued no dependency packages. That fixture was corrected by adding the
manifest and lockfile before the vulnerability proofs. This was fixture setup,
not a clean vulnerability claim.

Baseline `go run ./cmd/lictor grype --repo <fixture-absolute-path> --unpinned
-scan .cache/sbom-measure/scan.spdx.json --as-of 2026-09-30T00:00:00Z` printed:

```text
check FAILED: grype-gate config: what grype applied is not the committed .grype.yaml — db.max-allowed-built-age declared  is not a duration; db.validate-age effective=true declared=false; declared exclude ./bin/** is not in effect; effective exclude ./bin/** is not declared
```

The minimal declaration deliberately supplied only the exclude, so database
setting mismatches also appear. The quoted exclude mismatch is the defect under
investigation. This run stopped at configuration, before inventory judgement.

## Exact-byte replay binding

The entire JSON reports have root keys descriptor, distro, matches and source.
Descriptor has configuration, db, name, timestamp and version. Adding one newline
to the SPDX SBOM changed its digest:

```text
fixture.spdx.json      6d2fddfc02a5ac02c34764438733f843787403e2c6e9f5881836381cd127b8b9
same-content.spdx.json 06356514002fd4c58fff2d8d382106606d382388a12a6c436b39492cdd06030d
```

Scanning each with real Grype produced equal complete JSON documents after
removing only descriptor.timestamp and descriptor.configuration.output. No
remaining field distinguishes those exact byte sequences. Neither a source
name nor an SPDX document identity proves the attached artifact's SHA-256.
Therefore SBOM replay refuses; no guessed binding or custom attestation is used.
Live execution hashes a bounded read and scans a private snapshot of those bytes.

## Fixtures and proofs

internal/grype/testdata/high.spdx.json is the populated Syft SPDX output;
clean.cdx.json is Syft output from an empty directory. not-sbom.json is ordinary
JSON without an SBOM format. high.allowlist.json names every measured advisory
with a reason and ruling. high.report.json retains real matches and source with
an explicit descriptor projection; clean.report.json is a minimal empty report.
The offline scanner is the Go test executable under the name grype, with exact
argv and snapshot-byte checks. Unit tests never consult a database or network.

Before implementation, `go test ./internal/grype -run
'^TestRule_GRYPE_SBOM_1$' -count=1 -v` failed each of its original six subtests:
`explicit SBOM request is unavailable`. After implementation those pass, as do
bounds, incompatible options, scanner parser refusal, cancellation and explicit
time checks. The CLI test initially expected --sbom in Go flag help; flag help
prints -sbom. Its corrected spelling assertion passes. Rulefloor initially
refused the lowercase rule tag; the required uppercase RULE tag fixed arming.
GRYPE-SBOM-1 uses the observed pre-implementation failure as a manual red proof.

The existing Decide function is unchanged. TestSBOMVerdictParity feeds the same
recorded findings and allowlist through existing directory and image paths.
Both exit 1 with the same vulnerability summary:

```text
check FAILED: grype-gate matches=5 fixable=0 severe=2 — take the fix or allowlist it by name with a reason: GHSA-35jh-r3h4-6jhm (lodash 4.17.20, fixed in 4.17.21); GHSA-r5fr-rjxr-66jc (lodash 4.17.20, fixed in 4.18.0)
```

## Real CLI evidence after implementation

The live executable used --sbom, --unpinned, the fixture directory as --repo,
and --as-of 2026-09-30T00:00:00Z. Only the allowlisted case supplied --allowlist.
These are real Grype runs, not the offline helper. Findings reflect the installed
2026-09-30 database; the committed reports freeze the deterministic unit proof.

```text
exit=1
check FAILED: grype-gate matches=5 fixable=5 severe=2 — take the fix or allowlist it by name with a reason: GHSA-29mw-wpgm-hmr9 (lodash 4.17.20, fixed in 4.17.21); GHSA-35jh-r3h4-6jhm (lodash 4.17.20, fixed in 4.17.21); GHSA-f23m-r3pf-42rh (lodash 4.17.20, fixed in 4.18.0); GHSA-r5fr-rjxr-66jc (lodash 4.17.20, fixed in 4.18.0); GHSA-xxjr-mmjv-4gpg (lodash 4.17.20, fixed in 4.17.23) — subject sbom:high.spdx.json sha256:6d2fddfc02a5ac02c34764438733f843787403e2c6e9f5881836381cd127b8b9; config: not applicable (sbom subject: no directory declaration to compare); coverage: not applicable (sbom subject: no gitignored tree to enumerate)
```

```text
exit=0
check OK: grype-gate matches=0 fixable=0 allowlisted=0 unfixable=0 severe=0 — subject sbom:clean.cdx.json sha256:80ffc18b2ad8d6cc55036f18cf53c42bac80caaf0971fa5bfd5b8ff5afc6ea01; config: not applicable (sbom subject: no directory declaration to compare); coverage: not applicable (sbom subject: no gitignored tree to enumerate)
```

```text
exit=1
check FAILED: grype-gate matches=5 fixable=0 severe=2 — take the fix or allowlist it by name with a reason: GHSA-35jh-r3h4-6jhm (lodash 4.17.20, fixed in 4.17.21); GHSA-r5fr-rjxr-66jc (lodash 4.17.20, fixed in 4.18.0) — subject sbom:high.spdx.json sha256:6d2fddfc02a5ac02c34764438733f843787403e2c6e9f5881836381cd127b8b9; config: not applicable (sbom subject: no directory declaration to compare); coverage: not applicable (sbom subject: no gitignored tree to enumerate)
```

```text
exit=2
CANNOT-EVALUATE: grype-gate: input is not an SPDX or CycloneDX SBOM — subject sbom:not-sbom.json sha256:9de67ac19f755da6f146769b2390f6d7cf0dd3b36d6bc26e5d5f36b397a21a98; config: not applicable (sbom subject: no directory declaration to compare); coverage: not applicable (sbom subject: no gitignored tree to enumerate)
```

