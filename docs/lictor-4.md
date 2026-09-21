# LICTOR-4 (Amendment 1) implementation measurements

2026-09-21. Charter commit d9db9b1, parent 274f82f; charter SHA-256
cd04f8478605bef875a528883018bda2d5e0ddd882fde34e4baf6099dc240768.

These consumer measurements used the unreleased development code while its
version was still v0.2.0, matching consumer pins. Release v0.3.0 must refuse those
unchanged pins. This distinction is deliberate and is not a production v0.2.0
release or a claim that installed v0.3.0 passes a v0.2.0 pin.

## Before and after

- Missing and mismatched pins formerly accepted a saved image report (exit 0);
  now both refuse with exit 2, empty stdout and the exact owner diagnostic.
  Equal runs; absent plus --unpinned runs with an explicit warning; mismatch
  plus --unpinned still refuses. Every repository command checks before execution.
- GOFLAGS, GOWORK and GOENV were inherited. Now GOFLAGS is omitted, and
  GOWORK=off and GOENV=off are appended; every other allowlisted value, including
  GOTOOLCHAIN, remains. Grype's allowlist has no diff.
- Old green fixture line:
  `GREEN: builds, vets and tests clean — subject lictor-4a1-green-fixture; go version go1.27.1 darwin/arm64`
- New line:
  `GREEN: builds, vets and tests clean — subject lictor-4a1-green-fixture; go1.27.1`
- The red file hidden by GOFLAGS=-tags=nonexistent formerly passed GREEN;
  now it returns NOT-GREEN / go build FAILED. The clean fixture still passes.

## Clockfuse source versus port

The source is wiki/tools/clockfuse-gate.sh at wiki e2e7b44, byte-unchanged from
the requested 662585b source reference. Optional make clockfuse-conformance
compares stdout after replacing only the remedy command. The five cases:

| Source fixture | Source exit | Port exit |
|---|---|---|
| identical findings | 0 | 0 |
| line-number churn | 0 | 0 |
| fewer findings | 0 | 0 |
| new finding-class | 1 | 1 |
| count rise | 1 | 1 |

Pass text: `clockfuse-gate: no findings above snapshot.`
New-class text: `  NEW FINDING  internal/new/c_test.go|BarOptions|Clock — 1 now vs 0 in snapshot`
Count-rise text: `  NEW FINDING  internal/service/a_test.go|FooConfig|Now — 3 now vs 2 in snapshot`

Neutralizing the comparison yielded `want 1 got 0` for both red fixtures.
Scratch tests prove check-mode bytes unchanged, the snapshot's four exact header
lines, and source-exact count lines. The CLI fixture analyzer exits 1 on purpose;
that exit is tolerated, while missing tools, cancellation and unavailable snapshot
refuse. No consumer snapshot was regenerated. On saved live raw analyzer output,
the source normalizer and port agree exactly: OSS 14 classes, CE 28. OSS's
snapshot has 15, with one removed class; CE's has 28 unchanged.

## Route delegation

Installed Achta v0.5.10, source de3395b. Exact patterns are in PROJECT_SPEC.md.
OSS e642f92 selects three calls, each printing:

```text
declared-route check: pass; 2 workflow(s), 0 violation(s)
declared-route check: refused — line-level scope inference; required scope is resolved from YAML mapping nodes
declared-route check: refused — default route, banned-pattern, required-key, and scope vocabulary
declared-route check: pass; 2 workflow(s), 0 violation(s)
declared-route check: refused — line-level scope inference; required scope is resolved from YAML mapping nodes
declared-route check: refused — default route, banned-pattern, required-key, and scope vocabulary
declared-route check: pass; 2 workflow(s), 0 violation(s)
declared-route check: refused — line-level scope inference; required scope is resolved from YAML mapping nodes
declared-route check: refused — default route, banned-pattern, required-key, and scope vocabulary
```

CE 56a6675:

```text
rulefloor set skipped: no install line
lictor set skipped: no install line
```

UI 4371909:

```text
declared-route check: pass; 2 workflow(s), 0 violation(s)
declared-route check: refused — line-level scope inference; required scope is resolved from YAML mapping nodes
declared-route check: refused — default route, banned-pattern, required-key, and scope vocabulary
lictor set skipped: no install line
```

The brief incorrectly says CE installs Rulefloor: no matching workflow line
exists. The owner's install-line rule therefore skips both sets. Achta's native
`refused` explanatory lines above describe inference it declines, not a failed
route judgement; its check exits 0 and status is pass. No wording is reinterpreted.

The go-install and literal-release fixture each exit 1 with one native banned
pattern violation per Lictor key call. Fixtures use --achta-workspace pointing at
Lictor because Achta discovers multiple ancestor workspaces otherwise. This
explicit forwarding flag is an implementation addition; default argv is unchanged.
An isolated consumer with no wiki workspace will still refuse in Achta.

## Tests touched and assertion counts

Counts are static t.Fatal/t.Fatalf/t.Error/t.Errorf call sites, not runtime case
counts. Existing rule-bound tests remain unchanged.

- cmd/lictor/commands_test.go: 0 → 9.
- cmd/lictor/green_test.go: 7 → 7.
- cmd/lictor/main_test.go: 23 → 23.
- cmd/lictor/pin_test.go: 0 → 15.
- internal/clockfuse/clockfuse_test.go: 0 → 18.
- internal/executor/go_test.go: 3 → 4.
- internal/green/environment_test.go: 0 → 3.
- internal/route/route_test.go: 0 → 12.

The old CLI green assertion required `strings.Contains(human, got.GoVersion)`.
It now requires `strings.HasSuffix(human, "; "+strings.Fields(got.GoVersion)[2])`.
Existing CLI fixtures now declare a matching pin; existing expected verdicts
are unchanged. All new tests are offline fixtures.

## CE Grype evidence, before = after byte-for-byte

```text
check OK: grype-gate matches=0 fixable=0 allowlisted=0 unfixable=0 severe=0 — subject directory:identuum-idp-ce; config applied (3 exclude(s), db max-allowed-built-age 120h, 1 declared ignore(s)); re-check: 1 declared ignore(s), none lapsed on 2026-09-21; coverage: inventory 64 component(s) at 3 location(s), none under an ignored path (8 ignored path(s) read from git)
```

## Audit

================================================================================
GOGRAPH AGENT SESSION AUDIT
================================================================================
Session ID      : LICTOR4A1_20260921_151955
Status          : In Progress
Created At      : 2026-09-21T15:19:55+01:00
Ended At        : 2026-09-21T15:39:39+01:00
Duration        : 19m44s

━━━ METRICS ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total Commands  : 27
Successful      : 20
Failed          : 7
Success Rate    : 74.1%

━━━ COMPLIANCE SCORE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Plan Rule Run   : true (Weight: 35%)
Review Rule Run : true (Weight: 35%)
Composability   : 20.8% (Weight: 30%)

Overall Score   : 76.2%
Compliance Grade: C (Needs Improvement)

━━━ RECOMMENDATIONS ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Perfect! AI agent followed all core compliance and efficiency workflow rules.
================================================================================

## Open and adjacent

Consumer pin/route/snapshot changes are next-slice work, not this release. Achta
needs a workspace layout even for YAML-only route checks; no generic judgement
or layout is fabricated here. Malformed snapshot counts refuse as unjudgeable
input. Analyzer nonzero status is deliberately tolerated, matching the source;
the independent deadline gate must remain in the consumer. No other script
modes, YAML judgements or external queue edits were added.
