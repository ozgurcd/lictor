# LICTOR-8 — blocked dependent console output

Measured baseline: Lictor 099dfc3168de2271053ff31e5e79296085db9a83,
installed v0.4.3, main equal to origin/main. OSS source is
3718772b3820f8eea09a8e62f6436c060d28bd73, clean before and after the proof.
The owner's sole initial edit, PROJECT_DESC.md, stays untouched and unstaged:
SHA-256 86bbb4ecde57b6b3693d5b868407a5f264dd247cb1cd3d395574fdfc105a814f.

## Measurement and change

internal/witness/witness.go:203 gated blocked console output with
`dirty && o.All`. Its record write at :206 already retained NOT-RUN evidence,
elapsed zero and target exit 125. The fix changes only that guard to `o.All`.
The existing target-stream selection remains untouched. OSS
scripts/verify-all.sh:82–88 injects the reason into a normal recorder step,
which prints the blocked banner and reason on clean and dirty trees alike.

The baseline source comparison passed clean green and red, but failed the
blocked case. Relevant source stdout:

    ==> gate-witness: first
    ==> gate-witness: middle
    check FAILED: NOT-RUN middle: dependency first recorded exit=7
    ==> gate-witness: last

Lictor stderr omitted the middle banner and NOT-RUN line. Its record and exit
already matched. Clean stream routing, the repository context and stdout
summary are existing documented differences, not changes introduced here.

Existing console tests inspected include TestDirtyEchoSourceStreams
(dirty_echo_test.go:132), TestRunBytesUnchanged (:182),
TestRule_WITNESS_DIRTY_ECHO_1 (:50), TestDirtyEchoJSONTargetStream (:210),
TestModeAndFailureSemantics (conformance_test.go:202) and
TestDirtyRedLeavesRecordAndRunsTargets (edges_test.go:57). The mode test
discards console output and checks the blocked record. No existing assertion
was changed or weakened; the missing clean console contract gains a new test.

## Red, green and mutation

New internal/witness/all_echo_test.go binds WITNESS-ALL-NOT-RUN-ECHO-1.
The clean green/red/blocked/last fixture expects exactly:

    ==> gate-witness: green
    ==> gate-witness: red
    ==> gate-witness: blocked
    check FAILED: NOT-RUN blocked: dependency red recorded exit=7
    ==> gate-witness: last

Baseline v0.4.3 and restoration of `dirty && o.All` both produced:

    all_echo_test.go:21: plan-order console mismatch
    got:
    ==> gate-witness: green
    ==> gate-witness: red
    ==> gate-witness: last
    --- FAIL: TestRule_WITNESS_ALL_NOT_RUN_ECHO_1

With `if o.All`:

    --- PASS: TestRule_WITNESS_ALL_NOT_RUN_ECHO_1

The test also holds exit 1, minting, the exact stdout summary, blocked-target
nonexecution, record exit 125 and the independent last target. Rulefloor
arms hash 76118a6941a9 with mutation_observation; FLOOR and RED-PROOFS 8 → 9.
Existing bindings are unchanged. New test file assertion count 0 → 11,
counting explicit t.Fatal/t.Fatalf/t.Error/t.Errorf call sites including setup
guards (`rg -c 't\.(Fatal|Fatalf|Error|Errorf)\('`); existing test files unchanged.

## Frozen-clock comparison

Both source and port receive 2026-09-21T12:00:00Z: the existing test-binary
date shim feeds the source, --as-of feeds Lictor. No timestamp normalization.
Comparison ran before the version bump so both CLI binaries reported v0.4.3;
the old binary was copied from the installed release. This isolates behavior
from deliberately different runtime pins.

Corrected command, from the repository root with its Makefile cache environment:

    go test ./internal/witness -count=1 -timeout=120s -v -run '^(TestCleanAllSourceStreams|TestDirtyEchoSourceStreams|TestRunBytesUnchanged|TestStepBytesUnchanged|TestSourceConformance|TestRule_WITNESS_ALL_NOT_RUN_ECHO_1)$' -witness-source-script /Users/odemir/Development/identuum/identuum-idp-oss/scripts/gate-witness.sh -witness-all-script /Users/odemir/Development/identuum/identuum-idp-oss/scripts/verify-all.sh -witness-consumer /Users/odemir/Development/identuum/identuum-idp-oss -witness-cli /Users/odemir/Development/identuum/lictor/bin/lictor -witness-before-cli /Users/odemir/Development/identuum/lictor/.cache/lictor-v0.4.3

Six top-level tests and sixteen subtests ran (the named RUN/PASS lines), none
skipped, exit 0. Verbatim proof excerpts:

    clean green: records byte-equal; exit source=port=0; complete streams equal except declared routing, repository context and summary
    clean red: records byte-equal; exit source=port=1; complete streams equal except declared routing, repository context and summary
    clean not-run: records byte-equal; exit source=port=1; complete streams equal except declared routing, repository context and summary

Each dirty green/red/not-run case:

    stdout diff: empty (unfiltered full streams)
    stderr diff: empty (unfiltered full streams)

Exits respectively source=lictor=0, 1, 1. Full stdout includes the finalized
scratch record, so its equality also proves scratch-record equality. The
requested record remains byte-unchanged, SHA-256
ac1b631bfbf2e173957dbce967ef39e61a7bcac740f4fd840859bb4cf2d8c3ec.

    run dirty=false: stdout, stderr, record byte-identical; exit before=after=1
    run dirty=true: stdout, stderr, record byte-identical; exit before=after=1
    stepwise dirty=false: init, green/red/last steps and finalize stdout, stderr, records and exits byte-identical
    stepwise dirty=true: init, green/red/last steps and finalize stdout, stderr, records and exits byte-identical

The existing source-conformance suite also passed digest, commit, failed run,
blocked all-target run, stepwise and read-only OSS comparisons. OSS porcelain
was empty before and after, HEAD unchanged at 3718772.

The first Make invocation lost a dollar in its selector and ran zero tests;
another unquoted alternation was interpreted as shell pipelines (exit 127).
Neither counts as proof. The owner authorized the corrected direct invocation
and clarified that pre-execution invocation errors are corrected once and
reported, rather than treated as a measured-red STOP.

## Release census and limits

The charter's tagged literal census for v0.4.3 found exactly cmd/lictor/main.go,
schemas/lictor.version.v1.json and schemas/lictor.capabilities.v1.json.
All three hits are release constants and advance to v0.4.4 in the same commit.
README and PROJECT_SPEC current-release prose advance; historical notes and
proof provenance retain their original versions. RELEASE_NOTES names the fix.

make verify is required before the release commit and at the release head,
including rulefloor and wiki-check. Publication uses the existing release.yml
once from the authorized annotated tag; final remote/run/archive/formula and
installed-byte evidence is reported after those operations, not predicted here.
No new divergence, consumer change or record-format change is introduced.
Consumer adoption and the already-recorded tool-path/record-reader work remain
separately scoped; no external queue is written by this slice.
