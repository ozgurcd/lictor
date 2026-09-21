# LICTOR-3 measurements — 2026-09-21

Charter commit 22c1aba preserves PROJECT_DESC.md SHA-256
c5cf54faa3b45e55e197fca70efceece8413e11cda27937552afc102127d09f3.
Source: wiki 662585b, tools/repo-green-gate.sh, 366 lines. Only green_check and
the eight non-hook expect cases are ported. Lines 207–296 also contain hook
test machinery; it is excluded as the brief requires.

Red before implementation: TestGreenCommandContract's nomod, green and red
subtests failed with `CANNOT-EVALUATE: unknown command: green`. After the port,
all three pass, checking outcomes, causes, JSON Schema, base-name identity,
Go version and stderr/JSON excerpt parity. The GREEN-FLOOR-1 mutation skipped
all step failures; nocompile, testnocompile, testfail, vetfail and gofmtfail each
failed with `exit: want 1, got ... Outcome:GREEN ... ExitCode:0`. Restored afterward.

## Source versus port

Each case runs over a t.TempDir fixture module. The no-go port uses a Tools
implementation returning exec.ErrNotFound; PATH is not changed for the port.
Only the comparison source process uses the source script's historical PATH
fixture. Normal tests are offline and need no sibling. The optional
`-green-source-script=/absolute/path` flag belongs to internal/green's test binary,
not to Lictor or to every package's test binary.

| Fixture | Source exit | Port exit | Shared outcome and cause |
|---|---|---|---|
| green | 0 | 0 | GREEN; builds, vets and tests clean |
| no-go | 3 | 2 | CANNOT-EVALUATE; no 'go' on PATH |
| nocompile | 1 | 1 | NOT-GREEN; go build FAILED |
| testnocompile | 1 | 1 | NOT-GREEN; go vet FAILED |
| testfail | 1 | 1 | NOT-GREEN; go test FAILED |
| vetfail | 1 | 1 | NOT-GREEN; go vet FAILED |
| gofmtfail | 1 | 1 | NOT-GREEN; gofmt FAILED — unformatted files |
| nomod | 3 | 2 | CANNOT-EVALUATE; not a Go module |

Source: eight cases each check exit and output match (16 predicate checks).
The port retains both and checks identity and red diagnostics. Static test
failure callsites (`t.Fatal`, `t.Fatalf`, `t.Error`, `t.Errorf`, including setup
guards; not a coverage metric) in new files are cmd/lictor/green_test.go 0→7,
internal/green/green_test.go 0→15, internal/executor/go_test.go 0→3. Existing test
files have no changes. Additional tests pin order, early stop, timeout refusal,
gofmt output refusal, excerpt bounds/filter, environment isolation and capture
overflow. No timeout or formatting floor was relaxed.

## Consumer evidence

`GREEN: builds, vets and tests clean — subject identuum-idp-oss; go version go1.27.1 darwin/arm64`

`GREEN: builds, vets and tests clean — subject identuum-idp-ce; go version go1.27.1 darwin/arm64`

Both exit 0. OSS was at 63ee215 with ` M GATE-RUN.txt`; CE at 128dc78 was clean.
Heads and porcelain were unchanged afterward. Build/module/temp caches for these
runs were explicitly selected inside Lictor. No consumer source, gate, workflow,
pin, test, hook or record was edited. internal/grype has no diff from 36c670e.

The full local validation, release workflow, published checksums, tap commit and
installed proofs are reported from their completed runs, not inferred here.
