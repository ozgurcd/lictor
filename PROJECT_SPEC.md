# Lictor specification

## 1. Purpose

Lictor executes Identuum's gate programs and applies the house policy. It replaces
copies and sibling-path calls with one versioned executable. Generic tools retain
their judgements; Lictor consumes their answers. Its first command ports the OSS
Grype judge at `1cbe9f6c1df8dee83f8ba93e66217cf170b454a7` (1,659 lines).
The owner-controlled [PROJECT_DESC.md](PROJECT_DESC.md) is the binding charter.

## 2. Boundary

Lictor is Identuum-specific. A judgement useful to an unrelated workspace belongs
in the appropriate generic tool, never a parallel implementation here. Lictor may
execute fixed tool invocations, with context cancellation, deadlines, an explicit
environment allowlist and bounded output. It never invokes a shell, accepts an
arbitrary executable command, mutates Git, reads local credential files, or makes
network requests itself. External tools retain their documented execution effects.
The consumer Makefile remains responsible for any witness commit.

Knowledge belongs only in this repository's co-versioned wiki. Every commit adds
its local ledger row and log entry. Parent-wiki adoption and consumer pin changes
belong to a later authorized slice.

## 3. Command contracts

`version` and `capabilities` are repository-independent. Each accepts `--json`.
`grype` takes `--repo ABSOLUTE_PATH`; omission selects the process working directory.
The CLI supplies the selected directory explicitly to the judge; a saved directory
report naming another tree is refused. No sibling discovery or mutable current
repository exists. Any future sibling input must be `--sibling NAME=ABSOLUTE_PATH`.

`grype` also accepts `-scan`, `-inventory`, `-allowlist`, `-coverage-only`, `--as-of`
and `--json`. `-root` has been replaced by `--repo`. Explicit relative scan and
inventory filenames resolve from the caller's working directory; relative allowlist
paths resolve from the selected repository. The default is `grype-allowlist.json`.
A missing allowlist means no entries, as in the source judge.

Exit 0 means pass, 1 means an evaluated failure, and 2 means cannot-evaluate.
The human stdout evidence line stays byte-identical to the source judge, including
`grype-gate` and the directory base-name subject. Separate stderr context names the
absolute selected repository and evaluation time, including on refusals; it is not
part of the legacy evidence line. JSON emits one `lictor.grype.v1` document: its
`evaluation.reason` is exactly the human evidence text, and it carries `repository`,
`as_of`, `evaluation.outcome` and `exit_code`. Schemas live under `schemas/`.
Repository-independent command errors use `lictor.error.v1`.

The CLI samples its clock once per run. `--as-of RFC3339` replaces that sample and
is normalized to UTC. Lictor uses it only for suppression re-check dates. A lapsed
or undated suppression remains a failure. Replay fixes the scan, matching inventory,
allowlist, declaration, selected tree/ignored paths and `--as-of`; it does not
re-create a past scanner database. Grype's database-age clock remains inside Grype.
The applied-configuration predicate still requires the declared database-age
settings, exclusions and ignores. Image subjects retain both NOT APPLICABLE
predicates; their vulnerability verdict remains the gate.

Live execution is exactly `grype dir:. --output json=REPORT --output
cyclonedx-json=INVENTORY` in the selected tree. Scanner work has a five-minute bound;
Git ignored-path enumeration has a 30-second bound. Each stdout/stderr stream is
limited to 8 MiB, and report reads to regular files of at most 64 MiB. Oversize,
missing executable, cancellation and unreadable input are refusals. Temporary
scanner reports are removed after judgement. Git receives optional-lock suppression
and fsmonitor disabling. The allowlisted environment contains PATH, HOME, TMPDIR,
XDG_CACHE_HOME, GRYPE_DB_CACHE_DIR, GRYPE_DB_AUTO_UPDATE and
GRYPE_CHECK_FOR_APP_UPDATE; it adds GIT_OPTIONAL_LOCKS=0 and LC_ALL=C. No scanner
policy override or credential variable is inherited by this adapter.

Go 1.27.1 is the creation floor. Runtime implementation uses the standard library.
The sole direct test dependency, jsonschema/v6, validates published JSON Schema;
the standard library parses JSON but does not validate JSON Schema. No database,
ORM, UUID allocator, HTTP client or generic command runner is introduced.
CLI dispatch is in cmd/lictor, the policy in internal/grype, and fixed process
adapters in internal/executor. Adapters import no command package.

`green --repo ABSOLUTE_PATH [--json]` ports only repo-green-gate's check mode.
Omitted --repo selects cwd. The policy in internal/green stops at the first red:
`gofmt -l .`, `go build ./...`, `go vet ./...`, then
`go test ./... -count=1 -timeout=120s`. It first obtains `go version` in the
selected directory. Missing Go or go.mod is CANNOT-EVALUATE (exit 2, replacing
the source script's 3); an evaluated floor failure is NOT-GREEN (1); all four
steps passing is GREEN (0). Hook input, bypass and --selftest are not commands.
Human stdout names the outcome, original cause phrase, repository base name and
Go version (or explicitly unavailable). JSON is lictor.green.v1: outcome, cause,
subject, go_version, excerpt and exit_code. No home path appears in the evidence
line. Tool diagnostics may contain source paths; they are bounded, not redacted.

Go execution uses fixed argv in internal/executor, no shell. Each floor step
has a five-minute process bound; the inherited per-package test timeout remains
120 seconds. Version probing has a 30-second bound. Combined stdout/stderr is
limited to 8 MiB, then excerpted to twelve lines and at most 16 KiB; test failures
retain only FAIL headings and test-file lines, as the source does. The excerpt
is printed on stderr and included in JSON. Cancellation, output overflow or
an unstartable tool refuses with exit 2, never a green verdict.

The separate Go environment allowlist is PATH, HOME, TMPDIR, XDG_CACHE_HOME,
GOCACHE, GOMODCACHE, GOPATH, GOROOT, GOENV, GOTOOLCHAIN, GOWORK, GOFLAGS, GOPROXY,
GOSUMDB, GOPRIVATE, GONOPROXY, GONOSUMDB, GOOS, GOARCH, CGO_ENABLED, CC, CXX,
SDKROOT and MACOSX_DEPLOYMENT_TARGET, with GIT_OPTIONAL_LOCKS=0 and LC_ALL=C.
Go retains its own configuration and execution effects, including build caches
and module downloads; Lictor neither reads credentials nor confines consumer
tests. Offline fixture tests set proxy/sumdb off and use the local toolchain.
The Grype adapter's environment and policy are unchanged.

`make verify` runs toolchain-check, format-check, build, test, race, vet,
staticcheck, govulncheck, tidy-check, rulefloor and wiki-check, serially. Default
tests are offline fixtures. `make grype-scan REPO=/absolute/path` is an explicit
live scan outside that default plan. Tool pins are Go from go.mod, staticcheck
v0.8.1, govulncheck v1.7.0, rulefloor v0.9.1 and Achta v0.5.10.
Make keeps its Go module cache in `.git/lictor-cache/go-mod`, so dependency test
sources are not part of the repository source census. Other build caches and
temporary files are under `.cache/`. This development layout requires a normal
Git checkout; all these paths stay within this repository.

The release workflow requires an owner-authorized annotated vX.Y.Z tag matching
`lictor version --json`. It tests source and emits reproducible Darwin/Linux,
arm64/amd64 archives using CGO_ENABLED=0, trimpath, buildvcs and stripped symbols;
tar order, ownership and mtime are fixed and gzip omits timestamps. checksums.txt
is checked before publication with RELEASE_NOTES.md. No Windows build is produced.
A later owner-authorized release updates only Formula/lictor.rb in the Homebrew
tap, using public release URLs and the published checksums without a token,
then verifies installed bytes. The repository and releases are public since
2026-09-20. LICTOR-0 created no tag, release or tap change.

## 4. Migration inventory

Rows 1 and 2 are implemented. Other commands are proposed, not capabilities. The
workspace retirement ledger remains the inventory owner; this snapshot records
the charter's migration order with the measured Grype source count.

| Order | Program | Source | Lines | Rules bound | Why not achta | lictor command (proposed) |
|---|---|---|---|---|---|---|
| 1 | grype-gate | identuum-idp-oss `tools/grype-gate` | 1659 | GRYPE-FIXABLE-FAILS-1, GRYPE-SUBJECT-1 | vulnerability policy, outside achta's boundary by its spec; three consumers today | `lictor grype --repo` |
| 2 | repo-green-gate | wiki `tools/repo-green-gate.sh` | 366 | GREEN-FLOOR-1 (created by the port) | executor (build/vet/gofmt/test now); achta declares no shell execution | `lictor green --repo` |
| 3 | clockfuse-gate | wiki `tools/clockfuse-gate.sh` (+ OSS `tools/clockfuse`, 2048, test-policy analyzer) | 169 | none | test-policy analyzer | `lictor clockfuse --repo` |
| 4 | rulefloor-install-gate | wiki `tools/rulefloor-install-gate.sh`, mirrored into OSS and ui `scripts/` | 196 | CI-LOCAL-PARITY-1 pins its digest | achta `toolchain check` covers the pin half only; the workflow scan for a second install route is policy | `lictor rulefloor-install --repo` |
| 5 | gate-witness (run half) | wiki `tools/gate-witness.sh`, mirrored ×4 | 1039 | WITNESS-CLEAN-HEAD-1, WITNESS-ONE-RUN-PER-RECORD-1 name it; six more read its records | the RECORD half is achta's (`witness init/step/finalize/check`, ledger row 19, order 2, LAST); the RUN half executes targets | `lictor witness run --repo` calling achta for the record |
| 6 | toolchain-parity | OSS `tools/toolchain-parity` | 514 | CI-LOCAL-PARITY-1 | probes the INSTALLED binary; achta compares declarations only | `lictor toolchain --repo` |
| 7 | mint policy | OSS `tools/mint-reachability` | 1756 | MINT-REACHABILITY-1, MINT-RECORD-AUTHORITY-1, TOOLS-NO-REACH-1 | the CLASSIFICATION is generic and belongs in achta (`reachability classify` extended to classify every repository a gate-run record pins, from the record, in one verdict); the AUTHORITY policy (which record is the mint, absent/red/foreign-headed never satisfies, the Go build-closure proof of declared no-reach) is Identuum's | `lictor mint decide --repo --sibling`, after the achta feature |
| 8 | ci-witness, witness-earns, ledger-diff-gate | OSS `tools/*` | 505, 282, 1147 | CI-RECORD-HONEST-1; WITNESS-EARNS-ITS-CYCLE-1; LEDGER-DIFF-RECONCILED-1, LEDGER-REBASE-DERIVES-BASE-1 | each has an achta verb that covers most of it and a named gap (ledger notes under the GO table); decide per program: close the achta gap, or move the policy remainder here | per program |
| 9 | product analyzers | wiki `tools/route-parity-gate.sh`, `tessera-gate.sh`, `inert-parameter-gate.sh`, `emailed-link-gate.sh` | 286, 168, 244, 169 | none | cross-repository product contracts (ui against oss); nothing generic in them | `lictor analyze <name> --repo --sibling` |
| — | count-claim-check, close-condition-check, ledger-claim-check | wiki `tools/` | 322, 167, 205 | none | wiki-text judgements; achta spec §13 names count-claim a future component — stay in the wiki until achta takes them | not lictor |
| — | api-docgen, devseed, notrun, integration-witness | OSS `tools/` | 2886, 671, 838, 475 | OPENAPI-CHECKED-IN-CURRENT-1; —; —; INTEGRATION-GATE-1 | repository-local by nature (generate the repo's own spec, seed its own dev appliance, derive its own tagged-vet plan, run its own integration profile) | stay in identuum-idp-oss |
| — | gograph-first-hook.sh, `achta hook cd` | wiki `tools/`, achta | 558 | none | workspace rule enforcement for the coding agent's harness, not a repository gate | stay |
| — | witness-mint-test.sh | OSS `scripts/` | — | none | the state-based proof of the mint policy's Makefile wiring; moves with row 7 into lictor's own tests | follows row 7 |

## 5. Consumer adoption

Switch one consumer per slice, after an explicitly authorized Lictor release.
Declare one LICTOR_VERSION and LICTOR_SHA256 and one derived CI download route;
assert the installed version. Use public release URLs and published checksums;
no private-release access or token is required.
Replace the existing program invocation with `lictor COMMAND --repo "$(CURDIR)"`,
retain its evidence line, and replace the copied-file digest with a version pin in
toolchain parity. Remove that consumer's copy or sibling-path call in the same
slice. The next wiki slice records the retirement; the last consumer retires the
source program. A Legattus borrowed commit tree without siblings must answer for
facts contained in the commit rather than a missing sibling path.

## 6. Non-goals

No arbitrary `lictor exec`, wiki writer, replacement Make system, product runtime,
product migrations, or duplication of a generic tool's judgement. Future durable
queryable state requires its own case first; if authorized, use SQLite without
CGO, native SQL and explicit migrations, with canonical content identities.
