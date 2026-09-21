# LICTOR-5 Amendment 1 — recorder measurements, 2026-09-21

Owner charter committed unchanged at 86d4a8a, on top of b910076:
`ce27930a385961529ff2eebad5bb78568488ca087c6f0a4821a012f41d674cf0`.
Source: workspace wiki e2e7b44 tools/gate-witness.sh, 1,039 lines;
OSS e642f92 scripts/verify-all.sh, 97 lines. Neither was edited. The committed
OSS GATE-RUN.txt was inspected for the gate-run.v1 field shape and tool lines;
its old plan is historical, not the current Makefile's 34-entry VERIFY_PLAN.

## Scope and behavior

The amended writer sentence is bound verbatim as WITNESS-ONE-WRITER-1:

> A recorder opens a record by TRUNCATING it, refuses a second concurrent writer on the same record after a bounded wait, and refuses a second stepwise session on a record already open; it does not claim to prevent interleaving, which the reader backstops.

The source's check_mode:545–550 is the reader backstop for interleaved records;
this amendment neither implements nor tests that judgement.

`witness`/`witness run` preserves fail-fast run behavior; `--all` preserves
verify-all's independent-target behavior, with explicit --requires ordering.
`init`, `step`, `finalize` use the same record and source-compatible lock/session
keys. Source lock/session exits 3/4 are retained as the amendment requests,
despite the charter's general 0/1/2 exit convention. Step keeps real target exits.

## Proofs

Initial stub implementation: all four initial test groups failed, naming the
missing plan refusal, writer, dirty/stale-session behavior and output record.
The restored implementation passes them. Three separate mutation observations:

- O_TRUNC changed to O_APPEND: `open did not truncate: gate: old` (FAIL).
- Live-lock refusal bypassed: `live lock refusal`, ExitCode:0 (FAIL).
- Live-session refusal changed to success: `session refusal`, ExitCode:0 (FAIL).
- Restored: `PASS: TestRule_WITNESS_ONE_WRITER_1`.

Rulefloor armed hash `4436b3cf02f6`, FLOOR 6, RED-PROOFS 6. Later strengthening
of the separate concurrency fixture did not change that bound test's hash;
rulefloor explicitly refused a no-op rehash. The five earlier hashes remain
9170df3c6ab6, 8f7d35b89d2e, 376572882f86, d63cb3513614 and e591453742f1.

Explicit `make witness-conformance` measured:

- Digest-tied fixture: source/port byte diff empty.
- Commit-tied fixture with cites: source/port byte diff empty.
- Init/step/finalize fixture: source/port byte diff empty.
- Failed-target and verify-all dependency-blocked records: source/port byte diff empty.
- OSS read-only two-target run: byte diff empty; HEAD e642f92f78e32b95242e83a73038d2fa34292fbe;
  porcelain before and after ` M GATE-RUN.txt\n`.
- 34 OSS plan entries accepted, 28 make and six direct; no targets executed in
  this parse census. Pipe-doctored plan refused before execution:
  `target tool-versions: unquoted shell metacharacter '|' refused`.

Both implementations receive the same fixed clock; the compared record bytes
are not rewritten or normalized. Header, evidence selection, tool preface,
Go package count, elapsed, target exit and finalization are included. The OSS
comparison exercises the recorder engine directly, not a CLI pin bypass:
the consumer's older declared version must still refuse a newer CLI.

The real concurrent-writer fixture holds its child until explicitly released,
so scheduling cannot turn the refusal into a success. With a one-second test
bound the contender exits 3 and leaves the record untouched; CLI default is
the source's 120 seconds. Dirty green and dirty red leave prior records intact;
independent targets still run with --all. A stale session and stale lock are
broken with their notices. Live sessions refuse a second init, exit 4.

Ceiling evidence (sizes exclude the summary prefix/newline):

- `output=9437184 bytes plus summary prefix; target exit=0 preserved; truncated=false`
- `output=68157440 bytes plus summary prefix; target exit=7 preserved; truncated=true`

The 9 MiB summary is asserted whole in the record. The 65 MiB target retains
exactly the first 64 MiB of output, records truncation, and finalizes red because
its actual exit is 7. Judge adapters retain their prior 8 MiB refusal cap.

## Tool-line audit and queue item

OSS ci-witness ParseRecord (judge.go:190–225) selects gate, plan, target, result,
ci-run and tree fields, not tool lines. Toolchain-parity collect (main.go:81+)
reads workflow pins, go.mod, binaries and script digests, not recorded tool lines.
Witness-parity (Makefile:1325+) hashes its witness recipe, not the record.
The wiki's own script check likewise does not consume tool lines, but the wiki
calls Achta witness check for its e2e record (Makefile:608), whose
internal/witness/witness.go:368 recognizes `tool: ` as informational syntax.
It does not extract the absolute binary path. Conservatively following brief
item 5's any-reader branch, all tool lines are preserved byte-for-byte.

Queue item for the next authorized workspace-wiki slice:
“gate-run.v1 tool lines still disclose absolute executable paths. Achta's
witness parser recognizes tool: as an informational line, so LICTOR-5 retained
the source shape under its any-reader rule. Coordinate removal against that
reader contract before stripping paths. — OSS committed GATE-RUN.txt;
Achta internal/witness/witness.go:368; measured 2026-09-21.”

No external queue was written. The brief's explicit keep-and-report branch
takes precedence over the charter's stop-if-any-reader wording.

## Tests and limits

New test files; before counts are zero. After counts are static occurrences of
t.Fatal/t.Fatalf/t.Error/t.Errorf, including fixture helper assertions, not
executed case counts:

- cmd/lictor/witness_test.go: 0 → 2, plus shared schema validation.
- internal/witness/witness_test.go: 0 → 27.
- internal/witness/conformance_test.go: 0 → 36.
- internal/witness/edges_test.go: 0 → 8.

No existing test or assertion changed. The two source/consumer conformance
tests explicitly skip in default tests; the explicit conformance target ran
them with the authorized read-only paths. Unit tests have no sibling dependency.
No provider, vulnerability or evidence judgement was weakened.

The outside-tree source pathspec yields EMPTY-TREE, not a content digest. The
port reproduces it for compatibility; this external record is diagnostic only.
Source run stops at the first red despite its dirty-tree prose saying every
target runs. --all is the explicit verify-all counterpart. The six direct OSS
entries plus 28 make entries count is correct, but grype-scan's make invocation
does not contain --no-print-directory as the charter's shorthand implies.

## Consumer switch / open and adjacent

Move every affected LICTOR_VERSION and OSS LICTOR_SHA256 together. Replace
verify-all.sh with witness --all plus its dependency; preserve fail-fast CI and
integration drivers. Adapt GATE_RECORD_DRIVER/non-minting paths to the selected
record path and the new argv interface. CI inputs for cites/commit tie/sibling
pins become explicit flags. Pins still fail closed throughout the release window.

Keep all four gate-witness copies and their mirror/digest guards until the
reader migration to Achta. This execute half cannot promise that a record was
not written by two interleaved stepwise runs; it does not judge arbitrary
records, repeated finalizations or hand-edited plans. That is the next row's
boundary. Outside-tree digest repair and tool-path removal remain adjacent.
