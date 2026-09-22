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
executable command outside an explicit recorded witness plan, mutates Git, reads local credential files, or makes
network requests itself. External tools retain their documented execution effects.
The consumer Makefile remains responsible for any witness commit.

Knowledge belongs only in this repository's co-versioned wiki. Every commit adds
its local ledger row and log entry. Parent-wiki adoption and consumer pin changes
belong to a later authorized slice.

## 3. Command contracts

`version` and `capabilities` are repository-independent. Each accepts `--json`.
The release is v0.4.2; capabilities names version, capabilities, grype, green,
clockfuse, route and witness. Every repository command checks the first ci.yml line
matching `^  LICTOR_VERSION: (v[0-9][0-9.]*)`. A matching pin runs; a mismatch
refuses with exit 2, one stderr line naming the declared version and the public
brew install command, and no stdout (including under --json). Missing ci.yml or
no matching declaration also refuses, naming the file and `--unpinned` escape
for deliberate non-consumer use. `--unpinned` permits absence only, never a
mismatch, and prints one unpinned diagnostic. Unreadable/non-regular/oversize
ci.yml refuses even with --unpinned; declaration reads are bounded to 1 MiB.
This is the owner's exact line convention, not a YAML judgement. Help/version/
capabilities do not require a pin. Command JSON adds declared_version (string
or null) and pinned (bool); pin refusals deliberately emit no JSON document.
Own `make grype-scan` passes --unpinned, still enforcing any declaration.
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
subject, go_version, platform, excerpt and exit_code, plus pin metadata.
Human evidence ends in the version token alone (`go1.27.1`); JSON retains
`go version go1.27.1 darwin/arm64` and separately `platform: darwin/arm64`. No home path appears in the evidence
line. Tool diagnostics may contain source paths; they are bounded, not redacted.

Go execution uses fixed argv in internal/executor, no shell. Each floor step
has a five-minute process bound; the inherited per-package test timeout remains
120 seconds. Version probing has a 30-second bound. Combined stdout/stderr is
limited to 8 MiB, then excerpted to twelve lines and at most 16 KiB; test failures
retain only FAIL headings and test-file lines, as the source does. The excerpt
is printed on stderr and included in JSON. Cancellation, output overflow or
an unstartable tool refuses with exit 2, never a green verdict.

The separate Go environment allowlist is PATH, HOME, TMPDIR, XDG_CACHE_HOME,
GOCACHE, GOMODCACHE, GOPATH, GOROOT, GOTOOLCHAIN, GOPROXY,
GOSUMDB, GOPRIVATE, GONOPROXY, GONOSUMDB, GOOS, GOARCH, CGO_ENABLED, CC, CXX,
SDKROOT and MACOSX_DEPLOYMENT_TARGET, with GIT_OPTIONAL_LOCKS=0, LC_ALL=C,
GOWORK=off and GOENV=off. GOFLAGS is not inherited. The selected repository is
the module; ambient workspace/config files and build tags cannot hide a red.
GOTOOLCHAIN stays because the actual version is reported.
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

## Clockfuse and installation routes

`clockfuse --repo ABS [--snapshot] [--unpinned] [--json]` executes precisely
`go run ./tools/clockfuse .` in the selected repository, with the Go environment
above and a five-minute/8 MiB bound. Analyzer exit status is tolerated as in the
source wrapper; process startup errors, cancellation and overflow refuse.
The analyzer and the separate deadline gate remain consumer-owned. Missing
`tools/clockfuse` or `.clockfuse-snapshot` refuses by name with exit 2 (source 3).
Human clockfuse and route execution names the selected repository on stderr,
leaving native stdout evidence unchanged. An explicitly selected Achta workspace
is also printed there. Pin refusals remain exactly one stderr line.
Check mode writes nothing. Normalization preserves the source grep filter,
POSIX sed substitution, sort-before-count order and awk first-field behavior,
including unmatched filtered lines. Field splitting uses the source awk's
space/tab separators, preserving NBSP, CR, VT and FF inside paths. Comparison permits removals and line-number
changes; new classes and rising counts fail. Snapshot count input must be an
integer; malformed counts refuse rather than supplying a made-up baseline.

Explicit --snapshot writes only .clockfuse-snapshot, with four source-shaped
header lines naming `lictor clockfuse --snapshot`, the Git short head (or
untracked), format and regeneration instruction. Finding lines are source-exact.
Non-regular snapshot destinations refuse. Snapshot reads are bounded to 8 MiB.
Human check text is source-exact except the remedy command, now
`lictor clockfuse --repo <path> --snapshot`. JSON is lictor.clockfuse.v1, carrying
that reason, finding lines, subject, write flag, outcome, exit and pin metadata.
There is no stdin/selftest/sync command. CLOCKFUSE-SNAPSHOT-1 binds five fixtures.

`route --repo ABS [--achta-workspace ABS] [--unpinned] [--json]` scans .yml/.yaml
workflow files for non-comment lines mentioning a tool plus `install`,
`releases/download`, `archive/refs/tags`, `brew` or `go install`. This literal
selection does not parse or judge YAML. A tool with no install line is explicitly
skipped. Workflow directory enumeration and file reads are confined beneath
the selected repository with os.Root, preserving sorted enumeration. Outside
symlinks refuse before delegation. Reads are bounded to 1 MiB each. For each selected policy,
Lictor executes `achta declared-route check --dir <repo>/.github/workflows`
with `--route-cardinality per-file-any`, one `--route-pattern`, that same
`--required-route-pattern`, the following `--ban-pattern` values, the named
`--required-key`, and `--required-scope /env`.

Rulefloor pattern (exactly the wiki Makefile policy):

```text
rulefloor/archive/refs/tags/[$][{]RULEFLOOR_VERSION[}]
```

Rulefloor bans, in order:

```text
brew install[^#]*rulefloor
go install[^#]*rulefloor
rulefloor/archive/refs/tags/v[0-9]
-X main[.]version=[$][{]RULEFLOOR_VERSION[^}]
-X main[.]version=[$][{][^R]
-X main[.]version=[$][^{]
-X main[.]version=[^$]
```

Required key: RULEFLOOR_VERSION. Lictor pattern:

```text
lictor/releases/download/[$][{]LICTOR_VERSION[}]/lictor_[$][{]LICTOR_VERSION#v[}]_
```

Lictor bans, in order:

```text
brew install[^#]*lictor
go install[^#]*lictor
lictor/releases/download/v[0-9]
```

Achta accepts one required key per call, so Lictor invokes this set twice, once
for LICTOR_VERSION and once for LICTOR_SHA256. Both scope to /env. No
required-key-only mode is invented. The runtime self-check owns missing pins.

Achta version is obtained with fixed `achta version --json`, requiring its
version agreement. A declared ACHTA_VERSION uses the same exact ci.yml line
convention and must match; stderr states declared and installed versions, or
installed alone when undeclared. Missing Achta refuses by name. Version/check
bounds are 30/60 seconds, each output stream 8 MiB. No retries or shell.
Human native stdout is passed through unchanged; each native exit 0/1/2 is
preserved, with cannot-evaluate taking precedence in the aggregate. JSON adds
--json to native calls, retains each document and exit, and wraps selected/skipped
sets in lictor.route.v1. Unsupported or exit-inconsistent documents refuse.
INSTALL-ROUTE-1 binds selection, complete key sets and native-result preservation.

Achta currently requires a wiki workspace even for this YAML-only judgement.
If discovery is ambiguous, --achta-workspace passes the explicit absolute
workspace to Achta; Lictor neither discovers one nor manufactures one. An
isolated runner without that layout still refuses in Achta. This is an adjacent
Achta requirement for the consumer slice, not a reason to duplicate judgement.

## Recorded execution

`witness [run|init|step|finalize]` implements only the source recorder's execute
half. Default run stops at the first red; --all attempts independent targets as
verify-all does. Plan and --requires validation completes before locks, record
writes or target execution. Dependencies name earlier planned entries only;
blocked dependents record source-shaped NOT-RUN evidence and exit 125.
POSIX quote removal supplies argv without expansion. Unquoted pipe, ampersand,
semicolon, redirection, dollar, backtick, parentheses and newline refuse by
target name. Direct shell executables refuse; make retains its own execution.

Opening run/init truncates. Run excludes its selected record and GATE-RUN*.txt
when deciding whether work is dirty; a dirty in-tree run writes only a temporary
record and prints NOT MINTED. Under dirty human --all, stderr is exactly
`GATE-WITNESS NOT MINTING: dirty work; <record> remains untouched` before
execution; target banners and combined output go to stdout, followed by the
complete finalized scratch record and
`GATE-WITNESS NOT MINTED: dirty work; <record> is untouched`. This mode omits
the extra repository context line to preserve the source streams exactly.
JSON still names the repository and carries one lictor.witness.v1 result,
with target output on stderr. Default fail-fast run keeps its previous output.
The requested record is never opened for writing on the dirty path; the overall
exit remains 0/1, with dependent targets recorded as NOT-RUN 125. The rule
WITNESS-DIRTY-ECHO-1 binds the four source assertions over green, red and
dependency-blocked fixtures. Stepwise init deliberately retains the source's
CI behavior, permitting dirty work. Init owns a session through its parent PID;
run and init refuse a live session, step/finalize do not claim session ownership.
Finalization closes the session even when red. Missing records refuse. Symlink
or non-regular record destinations refuse before opening.

Locks and sessions share the script's canonical physical directory plus record
basename, SHA-256 keyed as /tmp/gate-witness-<key>.lock and .session. Lock wait
defaults to 120 seconds, polling once per second, with an explicit --lock-wait
override. Dead holders/sessions are broken with the source notice. Catchable
signals cancel CLI execution and deferred cleanup releases the acquired lock.
Writer refusal exits match the source: lock 3, session 4; step keeps target exit.
These are explicit exceptions to the general command exit contract; run/finalize
otherwise use 0/1/2. Per-invocation exclusion is not a claim that independent
step/finalize calls cannot interleave; check's judgement remains outside Lictor.

Each target has a 30-minute default deadline, explicitly configurable by
--timeout. Except for dirty human --all above, combined stdout/stderr streams
to stderr and a temporary file; only
the first 64 MiB are retained and mirrored. Excess is drained, not buffered;
`truncated: NAME output exceeded 67108864 bytes; retained=67108864 total=N`
is appended separately, preserving the real process exit. Startup, cancellation,
timeout or output I/O failure refuses and leaves incomplete evidence. Judge
adapters retain their 8 MiB fail-closed caps. Evidence lines allow 65 MiB per
line while streaming the record, preserving carriage returns and final lines
without a newline. Captured output is inspected before copying any line.
If that captured prefix contains NUL, the record contains exactly
`binary-output: NAME contains NUL; evidence omitted` and no tool or evidence
lines for the target, including no Go package count. Invalid UTF-8 without NUL
uses `binary-output: NAME contains invalid UTF-8; evidence omitted`. NUL takes
precedence if both occur. Neither case refuses: elapsed, target exit,
truncation if applicable and finalization retain their normal semantics.
The raw diagnostic stream remains unchanged; only the text record is filtered.
Label/citation input must be valid UTF-8 without NUL before opening the record.
Other target names are already ASCII-constrained; Git identities are hashes.
For valid text only the original EVIDENCE_RE matches, tool lines and Go package
counts are retained; the spool is removed at command completion.

Execution inherits the existing Go allowlist plus the three Grype cache/update
variables. No ambient GATE_WITNESS_* changes the record: --cites, --tie and
--sibling make those inputs explicit. --as-of freezes header/final/elapsed clock
reads for replay; otherwise a supplied clock is sampled at each source-equivalent
point. lictor.witness.v1 carries repository, record, reason, exit, whether a record
was written, and pin metadata. Target output always goes to stderr under JSON.

The source's outside-tree exclusion yields EMPTY-TREE. The port preserves that
record byte contract; such a record is diagnostic, not a usable witness.
Tool lines are preserved: Achta's allowedInformational parser recognizes tool:
when the wiki calls witness check. The absolute-path cleanup remains adjacent.
Source conformance is explicit `make witness-conformance SOURCE_SCRIPT=ABS
ALL_SCRIPT=ABS CONSUMER=ABS`; unit tests require no sibling checkout. The conformance fixture
supplies a fixed clock to both implementations, not timestamp normalization.

## 4. Migration inventory

Rows 1 through 4 and the execute half of row 5 are implemented. Later commands are proposed, not capabilities. The
workspace retirement ledger remains the inventory owner; this snapshot records
the charter's migration order with the measured Grype source count.

| Order | Program | Source | Lines | Rules bound | Why not achta | lictor command (proposed) |
|---|---|---|---|---|---|---|
| 1 | grype-gate | identuum-idp-oss `tools/grype-gate` | 1659 | GRYPE-FIXABLE-FAILS-1, GRYPE-SUBJECT-1 | vulnerability policy, outside achta's boundary by its spec; three consumers today | `lictor grype --repo` |
| 2 | repo-green-gate | wiki `tools/repo-green-gate.sh` | 366 | GREEN-FLOOR-1 (created by the port) | executor (build/vet/gofmt/test now); achta declares no shell execution | `lictor green --repo` |
| 3 | clockfuse-gate | wiki `tools/clockfuse-gate.sh` (+ OSS `tools/clockfuse`, 2048, test-policy analyzer) | 169 | CLOCKFUSE-SNAPSHOT-1 | Identuum snapshot arithmetic; analyzer remains consumer-owned | `lictor clockfuse --repo [--snapshot]` |
| 4 | rulefloor-install-gate | wiki `tools/rulefloor-install-gate.sh`, mirrored into OSS and ui `scripts/` | 196 | INSTALL-ROUTE-1 | Identuum pattern selection; Achta retains all YAML judgement | `lictor route --repo` |
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

Switch the consumers in a separately authorized consumer slice after release.
Until their pins advance, v0.4.2 intentionally refuses older version declarations.
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
