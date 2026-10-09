# Lictor

Identuum-specific gate execution. Current release: v0.4.5.
The binding charter is [PROJECT_DESC.md](PROJECT_DESC.md); implemented contracts,
boundaries and the migration plan are in [PROJECT_SPEC.md](PROJECT_SPEC.md).

## Install

```sh
brew install ozgurcd/tap/lictor
lictor version --json
```

The repository and release downloads are public; no GitHub token is required.
The [releases page](https://github.com/ozgurcd/lictor/releases) publishes archives
for Darwin/Linux on arm64/amd64 and a `checksums.txt` file for each release.
For CI, pin `LICTOR_VERSION` and the matching platform archive's published
`LICTOR_SHA256`, download from the public release URL, verify that checksum and
assert `lictor version --json` before running a gate.

## Build and use

```sh
make build
./bin/lictor version
./bin/lictor capabilities --json
./bin/lictor green --repo /absolute/path/to/repository
./bin/lictor grype --repo /absolute/path/to/repository
./bin/lictor grype --repo /absolute/path/to/repository --sbom production.spdx.json --as-of 2026-09-30T00:00:00Z
./bin/lictor clockfuse --repo /absolute/path/to/repository
./bin/lictor route --repo /absolute/path/to/repository
./bin/lictor grype --repo /absolute/path/to/repository -scan report.json -inventory inventory.cdx.json --as-of 2026-09-19T00:00:00Z --json
```

Every repository command requires a matching LICTOR_VERSION in the consumer's
.github/workflows/ci.yml. Absence or mismatch refuses with exit 2 and one stderr
line, with no stdout. `--unpinned` permits deliberate non-consumer use with no
declaration; it never bypasses a mismatch. JSON includes declared_version and
pinned. The released v0.4.5 intentionally refuses consumers pinned to another version.

The source judge's stdout evidence line is preserved. Human stderr separately
names the selected repository and as_of; JSON carries the same evidence as
`evaluation.reason`. Exit codes are 0 pass, 1 fail, 2 cannot-evaluate.
Without --repo, the selected repository is cwd. Saved directory reports must
name that selected tree. --as-of governs Lictor's suppression dates, never the
external scanner's clock. Replays require unchanged report, inventory, allowlist,
declaration and ignored-path inputs. Run `lictor grype --help` for all flags.

Live scans require Grype at the consumer's declared GRYPE_VERSION (source baseline
0.119.0) and Git. Lictor leaves toolchain-parity ownership with the consumer.
Executors have deadlines and output bounds; missing tools never pass.

`grype --sbom FILE` scans an SPDX or CycloneDX SBOM through Grype. Evidence names
the file by basename and SHA-256 of the exact bytes scanned. Directory
configuration comparison and ignored-tree coverage are explicitly not applicable;
the existing vulnerability and allowlist policy still applies, including failure
for an allowlisted High. Inputs must be readable regular files no larger than
64 MiB. Invalid input or an unavailable scanner exits 2. SBOM replay with -scan
is refused because Grype reports do not bind the exact input bytes; omit -scan
to run a fresh scan. See [the measurement and fixtures](docs/sbom.md).

`green` runs gofmt, build, vet and tests in that order, stopping at the first red.
The test invocation keeps `-count=1 -timeout=120s`. Human output names the subject
by base name and the Go version token, without platform. JSON retains the full
Go version and platform. GOFLAGS is dropped; GOWORK and GOENV are set to off. Exit 0 is GREEN, 1 NOT-GREEN, 2 CANNOT-EVALUATE
(including a missing Go toolchain or go.mod). `--json` emits lictor.green.v1 with
the same cause, subject, version and exit code plus the failing output excerpt.
The excerpt is also on stderr: at most twelve lines and 16 KiB. Go test excerpts
keep only FAIL headings and test-file lines. `lictor green --help` lists flags;
omitting --repo selects cwd. No hook, bypass or selftest command is included.
See PROJECT_SPEC.md for subprocess bounds and the Go environment allowlist.

`clockfuse` checks the consumer-owned analyzer against .clockfuse-snapshot.
Only new classes or rising counts fail; line changes and removals pass. The
explicit `--snapshot` form regenerates that file; default check mode writes
nothing. Missing analyzer/snapshot refuses. No selftest or stdin mode is shipped.

`route` selects house policy only where a workflow has a non-comment install
line, delegates each judgement to Achta, and names every skipped set. It checks
both Lictor keys in separate Achta calls. A declared ACHTA_VERSION must match;
otherwise the installed version is printed. `--achta-workspace ABS` explicitly
selects Achta's workspace when discovery is ambiguous. Achta still requires its
wiki layout; Lictor does not fabricate one on an isolated runner.

## Recorded execution

`lictor witness --repo ABS --record GATE-RUN.txt --label TEXT -- name=argv...`
runs a caller-declared gate plan and records gate-run.v1. POSIX quotes group
arguments; there is no expansion or shell. Unquoted shell metacharacters refuse
the entire plan before execution. Direct shell executables are refused; put
shell-dependent work in a make target.

To pass a gate's environment inputs, commit repeatable `--env NAME` flags in
the consumer's existing Makefile invocation. Declare names only, never values:

```make
verify-database:
	lictor witness run --repo "$(CURDIR)" --record GATE-RUN.txt --label database --env IDENTUUM_IDP_TEST_DATABASE_URL -- database="$(MAKE) --no-print-directory database-test"
```

The caller supplies the value through its environment. Lictor reads each declared
name once, forwards it only when set (including an empty value), and records
`environment: IDENTUUM_IDP_TEST_DATABASE_URL present` or `absent`. No other
ambient variables are added to the existing fixed allowlist. Duplicate names,
names outside `[A-Za-z_][A-Za-z0-9_]*`, every `GO*` name, and names already
controlled by Lictor are refused by name before execution or record writes.
Controlled names include PATH, HOME, CGO_ENABLED, CC, LC_ALL and the Grype cache
controls; see the full allowlist in PROJECT_SPEC.md. An accidental NAME=value
is refused without echoing the value. The flag is supported by `witness run`
only; stepwise modes refuse declarations.

Declared values of 8 bytes or more in target output become `[redacted]` before console
output or temporary spooling, including values split across output writes.
Partial value prefixes at the output ceiling are also hidden. Redaction matches
declared bytes, not arbitrary transformations or separately printed substrings.
Shorter values pass through; their record line says
`environment: NAME present (short value, not redacted)`, including empty values.
Target exits, timeouts and output bounds
remain unchanged. With no declarations, console and record bytes retain the
previous behavior.

The tag-triggered release updates and reads back Homebrew's `Formula/lictor.rb`
using the four published archive checksums. Its tap step fails closed naming
`HOMEBREW_TAP_GITHUB_TOKEN` if that required secret is absent.

Default `run` stops at the first failure. `--all` preserves verify-all's behavior:
attempt independent targets, and use repeated `--requires a:b` to record a
blocked dependent as NOT-RUN with exit 125. Dependencies must name earlier
planned targets. `init` accepts plan names, `step` takes one name=argv entry,
and `finalize` closes the same record. A missing target cannot finalize green.

Dirty work runs without replacing an in-tree record. Human `--all` reproduces
verify-all's streams: one NOT MINTING stderr notice, target output on stdout,
then the complete scratch record and the exact NOT MINTED stdout notice. It
does not add repository context to those source-compatible streams; JSON still
names the repository and routes target output to stderr. Default fail-fast
`run` retains its existing diagnostics and does not echo the scratch record.
Every GATE-RUN*.txt is excluded from the dirty-work decision.
The source's per-record
lock defaults to 120 seconds (`--lock-wait`); refusal is exit 3. A live stepwise
session refuses run/init with exit 4. Steps preserve the target's exit code;
run/finalize return 0 green, 1 red, or 2 cannot-evaluate. Sessions and locks use
the source-compatible physical-path keys under /tmp. Separate steps can still
interleave; only the external reader can judge such a record.

Except for dirty human `--all` above, targets stream combined output to stderr
and a temporary spool, capped at
64 MiB per target. Overflow is drained, recorded with `truncated:`, and does not
replace the target's real exit. Ordinary non-summary output remains on stderr;
the record retains the source's evidence selection for text output. Captured
NUL output produces `binary-output: NAME contains NUL; evidence omitted`;
invalid UTF-8 uses `contains invalid UTF-8` instead. Both suppress all tool and
evidence lines for that target and keep its real exit, elapsed and final verdict.
The raw diagnostic stream is unchanged except for declared-value redaction.
Labels/citations must be valid UTF-8
without NUL before opening a record. This deliberately replaces the source's
random-path binary-match message with a reproducible text line.
`--timeout` defaults to 30 minutes per target; other commands retain their caps.
`--cites`, `--tie commit`, and repeated `--sibling NAME=ABS` replace the source's
corresponding environment inputs explicitly. `--as-of RFC3339` freezes the record
clock for replay. `--json` emits one lictor.witness.v1 result.

An outside-tree record preserves the source's EMPTY-TREE digest limitation;
it is diagnostic only, not an attestation a reader can accept. `check`,
`--selftest`, and `--sync-check` are not ported. Consumer pins and the four script
copies remain until their separately authorized switch and reader migration.

## Validation

```sh
make verify
make wiki-check
```

Go 1.27.2, staticcheck v0.8.1 and jq must be on PATH, with
achta v0.5.15 or later within v0.5 and rulefloor v0.9.1 or later within v0.9.
Both tools must report `version_agreement == pass`; prereleases are refused.
CI pins exact versions and archive checksums. A new minor line or a required
feature needs a deliberate pin change. The compatibility selftest runs in `make verify`.
Make installs pinned govulncheck v1.7.0 under .cache/tools. Dependency download and
vulnerability-DB access belong to developer validation; unit tests are offline.
The runtime has no third-party dependencies. JSON Schema validation is test-only.
The published schemas are in schemas/; port measurements are in docs/lictor-0.md.
Build from a normal Git checkout: Make stores downloaded module sources under
.git/lictor-cache/go-mod and other build artifacts under .cache/.

## Releases

An owner-authorized annotated version tag triggers .github/workflows/release.yml;
its dispatch input can publish an existing immutable tag. Version, archives,
checksums, release notes and Formula/lictor.rb must agree. Consumer migrations
remain separate work after all affected pins and checksums move together.

## Shared CI toolchain bumps

After checkout, use `ozgurcd/lictor/.github/actions/go-toolchain@<full-commit-SHA>`.
Its only input, `go-version-file`, defaults to the caller's `go.mod`.
For a Go bump, change the `go` directive in each repository's `go.mod`.
For a Staticcheck bump, change the source, patch checksums and version assertion
once in `.github/actions/go-toolchain/action.yml`, verify Lictor CI, then update
the action SHA in each consumer. The binary cache includes the action's complete
recipe, Go version, runner OS and architecture; both versions are checked on hits
and misses. Other tools remain pinned by each consumer. Local Makefile tool
version checks must agree with the selected Staticcheck version.
