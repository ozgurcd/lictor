# LICTOR-7 — dirty all-target record echo

Charter-first commit: 2a7cd87, parent 471544f. Owner-controlled charter SHA-256:
`a3150eefa8768613b0a1c3c753fefd9c96346cc6d474ac748a2da3769061a887`.
No charter bytes were edited by this slice.

Source: OSS cd79b68bee33cfead1cd6612aa055131b9c5f95e,
scripts/verify-all.sh and cmd/identuum-idp/verify_all_test.go:88–110.
Exact source lines:

    GATE-WITNESS NOT MINTING: dirty work; $record remains untouched
    GATE-WITNESS NOT MINTED: dirty work; $record is untouched

Before implementation, TestRule_WITNESS_DIRTY_ECHO_1 failed green, red and
not-run: `target lines=0 want=3`, `last target absent`, `NOT MINTED absent`
(the exact required notice), and stderr mismatch. Full CLI comparison also
failed both streams for each plan. Record preservation already worked.
The early red-run log mistakenly printed an unconditional success-looking
summary after its failures; that summary is now conditional on !t.Failed().
The FAIL result and individual failed assertions were never suppressed.

After implementation, the four reproduced assertions are: exact NOT MINTED
present; exactly three target lines for the three-entry plan; last target
present; committed record unchanged. Each holds for green, exit-7 and
dependency NOT-RUN 125 plans. The source and Lictor use the same fixture,
commands and explicit clock (2026-09-21T12:00:00Z); a test-binary date shim
supplies that clock to the source, with no timestamp/output normalization.

For each of green, red and not-run, verbatim comparison proof:

    stdout diff: empty (unfiltered full streams)
    stderr diff: empty (unfiltered full streams)

Exits respectively:

    exit: source=0 lictor=0
    exit: source=1 lictor=1
    exit: source=1 lictor=1

Each plan's committed record proof:

    record sha256: before=ac1b631bfbf2e173957dbce967ef39e61a7bcac740f4fd840859bb4cf2d8c3ec after-source=ac1b631bfbf2e173957dbce967ef39e61a7bcac740f4fd840859bb4cf2d8c3ec after-lictor=ac1b631bfbf2e173957dbce967ef39e61a7bcac740f4fd840859bb4cf2d8c3ec

Fail-fast compatibility, measured with saved pre-change and patched CLIs
while both still reported v0.4.1 (before the authorized version bump):

    run dirty=false: stdout, stderr, record byte-identical; exit before=after=1
    run dirty=true: stdout, stderr, record byte-identical; exit before=after=1

Only test file added: internal/witness/dirty_echo_test.go. Assertion-count
predicate: explicit t.Fatal/t.Fatalf/t.Error/t.Errorf call sites in that file,
including fixture setup guards: 0 → 23 (9 Fatal, 4 Fatalf, 10 Errorf).
Four top-level tests were added. Amendment 1 changes only the old --all notice
literal in TestDirtyRedLeavesRecordAndRunsTargets (internal/witness/edges_test.go);
all its other assertions remain byte-identical. No other old-phrase test was
found by the literal census. The new floor
binding is fd6b558b2b7f, manual_observation of the pre-fix red; existing six
hashes are unchanged. FLOOR 6 → 7, RED-PROOFS 6 → 7. Capabilities remain the
seven commands version, capabilities, grype, green, clockfuse, route, witness.
No change to schemas/lictor.witness.v1.json and no check implementation.

Source conformance command:

    make witness-conformance SOURCE_SCRIPT=/absolute/OSS/scripts/gate-witness.sh ALL_SCRIPT=/absolute/OSS/scripts/verify-all.sh CONSUMER=/absolute/OSS TEST_ARGS='-run TestSourceConformance\|TestDirtyEchoSourceStreams -witness-cli /absolute/lictor/bin/lictor'

This optional source proof has external read-only inputs; the bound rule and
default verification use local offline fixtures. OSS porcelain before/after
the read-only record proof: ` M GATE-RUN.txt`, with the same head cd79b68.

OPEN AND ADJACENT: the consumer slice must switch OSS verify, delete its
verify-all.sh, remove verify-check.sh's branch, and change the driver-string
assertion in verify_all_test.go. The scripts' record-judging half remains out
of scope. No external queue edit is authorized here. The existing diagnostic
EMPTY-TREE and binary-output divergence contracts are unchanged.
