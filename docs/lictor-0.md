# LICTOR-0 measurements

Source: identuum-idp-oss `1cbe9f6c1df8dee83f8ba93e66217cf170b454a7`, seven files, 1,659 lines.

The owner charter is unchanged at SHA-256 `dca5e93276ff8bfb44090ad39ca0de3e5188b01e441dcf86d4834f9e9ee4c71e`.

## Port and assertions

Counts are static `testing.T` failure-call sites (Fatal/Fatalf/Error/Errorf/Fail/FailNow), including fixture preconditions; loops are not expanded. Original copies were checked against the source SHA-256 manifest.
- coverage_test.go: 35 source → 35 ported.
- decide_test.go: 18 source → 18 ported.
- subject_test.go: 24 source → 24 ported.
- New asof_test.go: 1 site, exercised at two dates; port_test.go helper: 0.
- New cmd/lictor/main_test.go: 23 sites; internal/executor/executor_test.go: 9.
- Original total: 77 → 77. New tests: 33 additional sites.

Port adaptations: package main becomes grype; CLI flags move to cmd/lictor and --repo replaces -root. The fixture helper maps original arguments to Options. Existing direct date tests use a fixed date and a saved source configuration fixture instead of the host clock and source repository file. The Makefile wiring assertion changes `tools/grype-gate` to `./cmd/lictor`; grype-scan remains an explicit live target. All vulnerability, applied-configuration, coverage and subject expectations remain intact. The rule text retains its original -root spelling; the CLI spelling is now --repo.

## Red observations

Before the override existed:
```text
flag provided but not defined: -as-of
--- FAIL: TestAsOfOverridesSuppressionDate
```
After the port:
```text
--- PASS: TestAsOfOverridesSuppressionDate
--- PASS: TestCoverage_LapsedSuppressionIsAFinding
```

Both migrated rules were re-watched with mutations, then restored byte-identical:
```text
    decide_test.go:49: a finding with a published fix must FAIL; got pass with "check OK: grype-gate matches=2 fixable=0 allowlisted=0 unfixable=2 severe=0"
--- FAIL: TestRule_GRYPE_FIXABLE_FAILS_1 (0.00s)
    subject_test.go:116: an image with one unfixable Low finding must pass; exit 1: check FAILED: grype-gate config: what grype applied is not the committed .grype.yaml — db.max-allowed-built-age declared  is not a duration; db.validate-age effective=true declared=false
--- FAIL: TestRule_GRYPE_SUBJECT_1 (0.11s)
```

## Live comparison

All four invocations exited 0. Grype 0.119.0 used a local copy of the existing v6.1.9 database (reported built 2026-09-19T06:27:50Z). Both commands in each pair received the same cache path, with automatic database/application update checks disabled for this measurement; database age validation and the declared 120h threshold stayed enabled and unchanged. This fixes the external database input and keeps scanner cache writes inside Lictor. Go caches and temporary directories were also redirected inside Lictor. No --as-of override was used for these live comparisons.

### identuum-idp-oss — source

`make grype-scan`

Porcelain before and after: `' M GATE-RUN.txt\n'` → `' M GATE-RUN.txt\n'`.

```text
check OK: grype-gate matches=0 fixable=0 allowlisted=0 unfixable=0 severe=0 — subject directory:identuum-idp-oss; config applied (3 exclude(s), db max-allowed-built-age 120h, 1 declared ignore(s)); re-check: 1 declared ignore(s), none lapsed on 2026-09-19; coverage: inventory 78 component(s) at 3 location(s), none under an ignored path (12 ignored path(s) read from git)
```

### identuum-idp-oss — lictor

`/Users/odemir/Development/identuum/lictor/bin/lictor grype --repo /Users/odemir/Development/identuum/identuum-idp-oss`

Porcelain before and after: `' M GATE-RUN.txt\n'` → `' M GATE-RUN.txt\n'`.

```text
check OK: grype-gate matches=0 fixable=0 allowlisted=0 unfixable=0 severe=0 — subject directory:identuum-idp-oss; config applied (3 exclude(s), db max-allowed-built-age 120h, 1 declared ignore(s)); re-check: 1 declared ignore(s), none lapsed on 2026-09-19; coverage: inventory 78 component(s) at 3 location(s), none under an ignored path (12 ignored path(s) read from git)
```

### identuum-idp-ce — source

`make grype`

Porcelain before and after: `''` → `''`.

```text
check OK: grype-gate matches=0 fixable=0 allowlisted=0 unfixable=0 severe=0 — subject directory:identuum-idp-ce; config applied (3 exclude(s), db max-allowed-built-age 120h, 1 declared ignore(s)); re-check: 1 declared ignore(s), none lapsed on 2026-09-19; coverage: inventory 64 component(s) at 3 location(s), none under an ignored path (8 ignored path(s) read from git)
```

### identuum-idp-ce — lictor

`/Users/odemir/Development/identuum/lictor/bin/lictor grype --repo /Users/odemir/Development/identuum/identuum-idp-ce`

Porcelain before and after: `''` → `''`.

```text
check OK: grype-gate matches=0 fixable=0 allowlisted=0 unfixable=0 severe=0 — subject directory:identuum-idp-ce; config applied (3 exclude(s), db max-allowed-built-age 120h, 1 declared ignore(s)); re-check: 1 declared ignore(s), none lapsed on 2026-09-19; coverage: inventory 64 component(s) at 3 location(s), none under an ignored path (8 ignored path(s) read from git)
```

Both evidence-line pairs are byte-identical. OSS make dry-run required no binary rebuild; its held GATE-RUN.txt modification stayed held. CE stayed clean. No consumer was migrated.

## Validation

Pre-commit `make verify` exited 0: toolchain checks, formatting, build, tests, race tests, vet, staticcheck, govulncheck, tidy-check, both Rulefloor bindings and the local wiki check passed.
```text
No vulnerabilities found.
PASS GRYPE-FIXABLE-FAILS-1 (internal/grype/decide_test.go @ unit)
PASS GRYPE-SUBJECT-1 (internal/grype/subject_test.go @ unit)
check OK: 2 rows (2 armed), FLOOR 2, RED-PROOFS 2 (measured 2)
```

Achta reports freshness pass with one co-versioned page skipped, and derive unchanged with zero opted-in blocks. That is its mechanical coverage, not a claim that it verifies the prose. Both workflow YAML files parsed; the two Linux validation-tool archive checksums and archive entry names were verified against downloaded release assets. Release publication was not exercised or authorized.

`lictor capabilities --json`:
```json
{"commands":["version","capabilities","grype"],"exit_codes":{"cannot_evaluate":2,"fail":1,"pass":0},"machine_interfaces":["lictor.version.v1","lictor.capabilities.v1","lictor.grype.v1"],"replay":"grype -scan JSON -inventory JSON --as-of RFC3339; unchanged repository inputs required","repository_selection":"--repo absolute path; default current working directory","schema_version":"lictor.capabilities.v1","version":"v0.1.0"}
```

## Development corrections and adjacent findings

- The first test selector matched no tests; it was corrected before the recorded red proof.
- A partially applied mechanical port produced a duplicate helper declaration; corrected before tests passed.
- The new expiry fixture first used its re-check day, which is inclusive; corrected to the next day without changing an existing fixture or assertion.
- Source-specific Makefile wiring assertions required the documented path adaptation.
- The executor output-limit test caught promoted bytes.Buffer.ReadFrom bypassing Write; a private buffer field removed that bypass. The process-level overflow and cancellation tests now pass.
- Achta required wiki/platform/decisions.md beyond the minimum scaffold list; added the local decision index.
- Initial hand-written unarmed ledger placeholders were rejected. The ledger was initialized through rulefloor declare and arm, with fresh observed mutation proofs.
- Rulefloor initially counted 254 orphan tags in downloaded OSS dependency tests under the ignored .cache/go-mod directory. The module cache now lives under .git/lictor-cache/go-mod; the unchanged full check passes. Generic cache/source selection remains adjacent Rulefloor work; no exception or generic-tool fix was added here.
- Gograph session creation before a local graph found an older ancestor session. It was left alone. Building the local graph allowed a separate Lictor session. Failed exploratory symbol/intention requests were corrected; no sibling graph was built.
- The charter still describes the retirement ledger as carrying 1,638 lines; the ledger row read during this restart now says 1,659. PROJECT_DESC.md was not edited.
- The source LICENSE is AllRightsReserved, not an open-source license. The owner explicitly authorized this port; no license policy or release permission was changed.

Closing Gograph audit: session LICTOR0_20260919_131228 completed; 30 commands, 24 successful, 6 failed; Plan Rule Run true; Review Rule Run true; score 83.8%, grade B. Final review attempts exposed unsupported source --package syntax and a case-insensitive ambiguity between Run and run even with a package-qualified selector; exact file reads completed that review. The final report records committed-head validation.

## Close owed outside this repository

A later authorized workspace-wiki slice owns the Lictor repository page and adoption ledger. Consumer switching, retirement of copies, version pins, private-release access and Homebrew installation wait for an authorized release and one consumer slice at a time. No parent-wiki entry, consumer edit, tag, release or tap edit was made here.
