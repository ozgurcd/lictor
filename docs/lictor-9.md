# LICTOR-9 — declared gate environment

Baseline: b7c1c62de1fe4b772b819b3d189e7e3adee919ae, main equal to origin/main
after fetch, annotated v0.4.4 points to the same commit. Only PROJECT_DESC.md
was initially modified. Its SHA-256 is
86bbb4ecde57b6b3693d5b868407a5f264dd247cb1cd3d395574fdfc105a814f;
the file is not edited, staged or committed. OSS was read-only at
6d994a79361f7ddb2a4e27586dcf049533db4413; no clean-tree condition was applied.

## Measurement and design

Baseline internal/executor/record.go:21-26 built the fixed Go environment plus
Grype cache controls. internal/executor/go.go:35-48 held the Go allowlist and
fixed overrides. cmd/lictor/witness.go:23-107 already parsed repeatable flags
and a caller-supplied argv plan. internal/witness/witness.go:272-289 wrote the
record header; :294-361 spooled output and retained evidence.

Repeatable witness run --env NAME is the smallest existing input surface:
consumers commit the flags in the Makefile invocation that already supplies
the plan. No new config file, discovery or plan grammar is needed. Stepwise
modes refuse declarations. The existing Go allowlist is shared with name
validation rather than copied. Every GO-prefixed name and fixed environment
name is reserved. Declared values are sampled once and added only to gate
processes. Set-empty and unset differ; the header records presence only.

The streaming filter replaces exact declared nonempty values before either
console output or temporary spooling. It handles arbitrary write boundaries,
overlapping values and truncated prefixes. Matching does not cover arbitrary
encodings or separately emitted secret substrings. The raw-byte output bound,
process exit and timeout remain unchanged. With no declarations, execution
uses the old output path and emits no environment lines.

Installed Achta measured v0.5.10 with module agreement pass before edits and
again before close. Item 0 therefore leaves all Achta pins unchanged:
Makefile:21, .github/workflows/verify.yml:16-17 and the live spec/README claims.

## Red and green

New cmd/lictor/environment_test.go, run against the saved v0.4.4 CLI with the
correctly quoted helper argv, failed in human, JSON and dirty-all modes:

    declared environment did not reach gate: exit=2
    --- FAIL: TestDeclaredEnvironment

Initial refusal cases against the original code also failed:

    missing named environment refusal
    --- FAIL: TestDeclaredEnvironmentRefusals

The implemented CLI passes the same fixture:

    declared set reached gate; undeclared and absent names absent; all output bytes exclude credential-shaped value
    --- PASS: TestDeclaredEnvironment
    --- PASS: TestDeclaredEnvironmentRefusals
    --- PASS: TestDeclaredEnvironmentAssignmentNeverEchoesValue

The child checks the actual DSN-shaped value, set-empty, absent and undeclared
variables before emitting its safe success line. It intentionally echoes the
synthetic value to both streams, including separate one-byte writes. Assertions
inspect complete stdout, stderr and record bytes without printing those bytes.
Removing the redactor assignment in RecordOutput produced, in each mode:

    declared value leaked in output bytes
    --- FAIL: TestDeclaredEnvironment

The assignment was restored and the same tests passed. Additional executor tests:

    --- PASS: TestEnvironmentControlledNames
    every two-chunk split, single-byte writes, truncated prefixes and overlapping values remain secret
    --- PASS: TestEnvironmentStreamSecrecy
    --- PASS: TestEnvironmentOutputFailure

A named refusal was also observed directly:

    environment name "PATH" refused: controlled by lictor

The first forwarding fixture after implementation used an unquoted dollar in
its test-selector argv and refused before execution. Quoting that argument
corrected the fixture; no production policy was relaxed. Documentation patches
also had context mismatches before successful application; those failed writes
made no changes.

## No-declaration comparison

Before the version bump, the saved v0.4.4 and changed CLI ran existing frozen-
clock fixtures. Command (repository-local Make cache environment):

    go test ./internal/witness -count=1 -timeout=120s -v -run '^(TestRunBytesUnchanged|TestStepBytesUnchanged)$' -witness-cli "$PWD/bin/lictor" -witness-before-cli "$PWD/.cache/lictor-v0.4.4"

Quoted results:

    stepwise dirty=false: init, green/red/last steps and finalize stdout, stderr, records and exits byte-identical
    stepwise dirty=true: init, green/red/last steps and finalize stdout, stderr, records and exits byte-identical
    run dirty=false: stdout, stderr, record byte-identical; exit before=after=1
    run dirty=true: stdout, stderr, record byte-identical; exit before=after=1
    PASS

Existing test files and assertions are unchanged. New assertion-call counts,
counting t.Fatal/t.Fatalf/t.Error/t.Errorf call sites including setup guards,
using rg -o over only the two new test files: cmd/lictor/environment_test.go
0 -> 13; internal/executor/environment_test.go 0 -> 11; total 0 -> 24.
No tests are weakened and RULE-FLOOR.md is unchanged.

## Docs and release

The documentation baseline search for --env, environment: and declared
environment returned no matches in README.md or PROJECT_SPEC.md. Both now
describe declarations, refusals, record presence and a Makefile example.
The charter's tagged version census returned exactly cmd/lictor/main.go and
the version/capabilities JSON schemas; all are release constants and move
together to v0.4.5. History remains unchanged. RELEASE_NOTES.md adds v0.4.5
in plain words. The local repository page's stale v0.4.3 lead is replaced by
the current release description, while its historical rows stay unchanged.

The serial make verify and release publication/archive verification results
belong to the final report, not a prediction in this committed measurement.
The fixture accepts -environment-cli ABS to repeat the proofs against the
published archive. No Homebrew, consumer pin or parent-wiki changes are made.
