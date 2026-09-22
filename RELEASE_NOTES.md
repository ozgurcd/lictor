# v0.4.2 — 2026-09-22

Dirty `witness --all` now reproduces verify-all's complete diagnostic record:
the exact NOT MINTING stderr notice, target output and finalized scratch record
on stdout, and the exact NOT MINTED notice. The committed record is untouched;
green, red and dependency NOT-RUN plans retain their outcomes. Full unfiltered
stdout/stderr comparisons against the source pass for all three plans.
Default fail-fast run is byte-identical on clean and dirty fixtures. JSON remains
lictor.witness.v1 with target diagnostics on stderr; check is not implemented.
WITNESS-DIRTY-ECHO-1 brings the armed floor and observed red proofs to seven.
No consumer is switched by this release; the existing two declared source
divergences remain unchanged.

# v0.4.1 — 2026-09-21

Clockfuse now preserves NBSP, CR, VT and FF inside finding paths, matching the
source awk field splitting instead of truncating those paths as Unicode whitespace.

Route selection now refuses workflow file or directory symlinks that escape the
selected repository, while preserving policy selection and Achta's judgements.

Witness now records NUL or invalid UTF-8 output with one deterministic
`binary-output:` line, omits that target's tool/evidence lines, and preserves
its real exit code and the normal elapsed/target/finalization lines.

The chosen NUL line is `binary-output: <target> contains NUL; evidence omitted`;
invalid UTF-8 uses `contains invalid UTF-8` instead. Invalid label/citation text
is rejected before opening a record. Captured text output is unchanged.
This is the first explicitly ruled recorder divergence: the source grep emits
`Binary file <random temporary path> matches`, while old Lictor copied raw NUL.
Neither belongs in a reproducible text record. The owner's consumer premise is
that no current target emits NUL; the six source record conformances remain
byte-identical. Those six fixtures are not an exhaustive capture of every target.

A second declared divergence is already shipped: malformed/overflowing
clockfuse snapshot counts refuse with exit 2 instead of the source's exit 0
after `integer expression expected`. It is unchanged in this release.
Pin grammar, ANSI-blind evidence matching, duration-shaped prose counting and
colon-path normalization limitations remain unchanged. No new command, rule row
or schema version; the floor remains six armed rules. Archives remain
Darwin/Linux on arm64/amd64. Consumer pin updates are separate work.

# v0.4.0 — 2026-09-21

Adds witness's execute half: argv plans, run/init/step/finalize, explicit all-target
and dependency handling, source-shaped gate-run.v1 records, dirty-work no-mint,
bounded per-record locks and stepwise-session refusals. The six-rule floor adds
mutation-proved WITNESS-ONE-WRITER-1. Capabilities now names seven commands.

Recorder output streams under a 64 MiB per-target ceiling; truncation is explicit
and retains the target's real exit. Other command limits and verdicts are unchanged.
Digest and commit-tied fixture records are byte-identical to the source, as is
the outside-tree OSS diagnostic record. Its inherited EMPTY-TREE digest is not
a usable witness. Tool lines remain unchanged because an existing reader
recognizes them. No record reader, selftest or mirror checker is ported.

Consumers must move pins/checksums and choose run versus --all deliberately;
their source copies remain until the reader migration. Per-invocation locks
cannot promise that separate step calls never interleave. Archives remain
Darwin/Linux on arm64/amd64.

# v0.3.0 — 2026-09-21

Every repository command now refuses an absent or mismatched LICTOR_VERSION;
--unpinned permits absence only. JSON names the declared version and pin status.
Go execution drops GOFLAGS and fixes GOWORK=off and GOENV=off. Green human
output ends in the version token; JSON retains full Go version and platform.
Grype policy and its recorded evidence line are unchanged.

Adds clockfuse check and explicit snapshot regeneration, preserving the source
normalization and count policy while keeping the analyzer in the consumer.
Adds route: installation-line selection and the Identuum patterns delegated to
Achta, with explicit skipped sets and native evidence preserved. The optional
--achta-workspace resolves Achta's existing discovery ambiguity without adding
workspace discovery or YAML judgement to Lictor. Both new commands have schemas
and mutation-tested rules; the floor is five armed rules. Consumer switches,
Achta runner prerequisites and source-script retirement remain separate work.
Archives remain Darwin/Linux on arm64/amd64.

# v0.2.0 — 2026-09-21

Adds `lictor green --repo ABSOLUTE_PATH [--json]`: the Identuum gofmt, build,
vet and test floor, with the inherited 120-second test timeout and first-red
behavior. Missing Go or go.mod refuses separately from a failing tree. Human
evidence names the repository base name and Go version; lictor.green.v1 publishes
the same result and bounded diagnostic excerpt. Eight source fixture cases and
GREEN-FLOOR-1 bind the port. No hook or bypass is ported; Grype policy is unchanged.
Archives remain Darwin/Linux on arm64/amd64, publicly downloadable with published
checksums. Consumer switches are separate work.

# v0.1.0 — 2026-09-19

Initial Identuum gate executor: the OSS Grype policy and tests, explicit repository
selection, saved-report replay with as_of, versioned JSON and bounded tool calls.
The source vulnerability and subject verdicts, including dated suppressions and
image NOT APPLICABLE predicates, are retained. No consumer migration is included.
Release archives target Darwin/Linux on arm64/amd64; no Windows artifacts.
