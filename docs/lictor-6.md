# LICTOR-6 — adversarial parser measurements

Measured 2026-09-21 at baseline 825edc8 (v0.4.0), with charter SHA-256
ce27930a385961529ff2eebad5bb78568488ca087c6f0a4821a012f41d674cf0.
Amendment 1 charter SHA-256 (committed byte-unchanged):
6039414b404c8ebfa7cab6d69a066aa5679de869651a788eb80abcad48ee941a.
Fixtures and logs are in the four new adversary_test.go files, binary_test.go and the local
ignored .work/lictor-6 directory. No consumer or source script was edited.

“Right” below means right under the current contract, not a claim that a
text matcher validates YAML or proves that arbitrary output came from Go.
LIMITATION means the source has the same behavior and was left unchanged.
Source pipeline measurements use LC_ALL=C. Plan measurements execute only
the source's name-validation loop, not arbitrary source command strings.

## Pin reader

Source: OSS Makefile LICTOR_ASSERT, first result of
`sed -nE 's/^  LICTOR_VERSION:[[:space:]]*(v[0-9][0-9.]*).*/\1/p'`.
The charter intentionally specifies one literal space in Lictor's pattern.

| Input | Lictor result | Assessment / source measurement |
|---|---|---|
| `# LICTOR_VERSION: v0.4.0` | no declaration | Right; source returns no match. |
| Job env, six-space indentation | no declaration | Right; source returns no match. |
| v0.4.0 then v99.0.0 | v0.4.0 | LIMITATION: source also returns first `v0.4.0`; neither judges duplicate YAML keys. |
| `"v0.4.0"` | no declaration | LIMITATION: source returns no match; quoted YAML is outside the line convention. |
| v0.4.0 with trailing spaces | v0.4.0 | Right; source `v0.4.0`. |
| Unclosed `[` before matching line | v0.4.0 | LIMITATION: source `v0.4.0`; YAML validity belongs to the YAML judge. |
| ci.yml directory | `ci.yml is not a regular file` | Right, refusal. |
| ci.yml symlink outside repository | `path escapes from parent` | Right, refusal. |
| 1 MiB + 1 byte file | `ci.yml exceeds 1 MiB` | Right, bounded read. |

No pin fix. Absence remains a refusal unless explicit --unpinned; no mismatch
bypass was introduced.

## Plan

Source: OSS scripts/verify-all.sh name-validation loop. POSIX argv splitting
and pre-execution shell refusal are explicit port rulings, not source features.

| Input | Lictor result | Assessment / source measurement |
|---|---|---|
| Newline in name | `invalid or duplicate plan entry` | Right; source exit 2, same diagnostic. |
| NUL in name | `invalid or duplicate plan entry` | Right; Go API fixture. OS argv cannot carry NUL, so no shell-argv equivalent is claimed. |
| No `=` | `invalid or duplicate plan entry: echo` | Right; source exit 2, same diagnostic. |
| `test=` | `target test: empty argv` | Right under argv contract; source name loop accepts, leaving shell execution later. |
| `test=echo 'x` | `target test: unterminated ' quote` | Right under argv contract; source name loop accepts, shell would parse later. |
| Exactly 1,048,576-byte entry | one entry parsed | Right; source loop also reports `accepted 1048576 bytes`. This proves parsing, not host exec argument capacity. |
| `Test` and `test` | two entries parsed | Right; source `accepted names: Test test`. Names are case-sensitive. |
| Two `test` names | `invalid or duplicate plan entry: test` | Right; source exit 2, same diagnostic. |

No plan fix. Error text retains hostile name bytes as before.

## Evidence and package matchers

Source: wiki/tools/gate-witness.sh EVIDENCE_RE and package grep at lines
134 and 200–205. Results below quote selected evidence bytes and package count.

| Input | Lictor / source result | Assessment |
|---|---|---|
| `ok this is prose` | empty / 0, both | Right. |
| `ok everything 1s later was wrong` | empty / 1, both | LIMITATION: source also counts duration-shaped prose. |
| ANSI-prefixed `check OK:` and `ok pkg 0.1s` | empty / 0, both | LIMITATION: source also misses the anchored colored lines. No color stripping added. |
| CRLF summaries | `check OK: yes\r\n` / 1, both | Right; CR byte preserved. |
| Invalid UTF-8 byte FF in summary/package | `check OK: \xff\n` / 1, both | Right for the measured C locale; raw byte preserved. |

Additional NUL probe: the source grep prints
`Binary file (standard input) matches` when fed stdin; its real recorder uses
a randomly named temporary file instead. Lictor's line matcher retains the
matching raw NUL. This is a separate observed record-fidelity conflict, not
one of the passing UTF-8 cases. Amendment 1 resolves the conflict: this source
behavior is a defect and the recorder deliberately diverges. The matcher itself
is unchanged; the recorder now checks captured output before copying any line.

The full init/step/finalize comparison also reproduces this, beyond the
isolated matcher probe. Both recorders execute the same fixed printf fixture
and every stage exits 0. Their actual record evidence is:

```text
source: evidence: [probe] Binary file /Users/odemir/Development/identuum/lictor/.cache/tmp/gate-witness-out.FIy1sK matches
port:   evidence: [probe] check OK: \x00
both:   evidence: [probe] go packages ok: 1
```

Here `\x00` represents the actual NUL byte, not four printable characters.
The records are explicitly labeled adversarial fixtures, not repository gates.
All record and spool paths are under Lictor. The complete invocation outcomes
are saved in .work/lictor-6/nul-full-record-measurement.json.

The fixed recorder prints `binary-output: probe contains NUL; evidence omitted`.
It writes no tool/evidence line for that target (including no package-count
evidence), then preserves elapsed and the real target exit. Invalid UTF-8
without NUL uses `binary-output: probe contains invalid UTF-8; evidence omitted`.
NUL has precedence when both appear. Diagnostic output remains byte-identical;
the text-record requirement applies to the retained capture, including output
cut at the existing ceiling. Non-UTF-8/NUL labels or citations refuse before
opening a record; this is invalid request metadata, not a target-output refusal.

The table above preserves the original matcher measurements. The end-to-end
recorder's invalid-UTF-8 and NUL behavior is deliberately changed by Amendment 1.
The owner's premise is that current consumer targets emit no NUL. All six
conformance fixtures remain byte-identical; those fixtures alone do not prove
that premise for every real target.

## Clockfuse

Source: wiki/tools/clockfuse-gate.sh normalize and --check-stdin. The exact
normalizer function was invoked on these bytes, without changing the script.

| Input | Source / Lictor before → after | Assessment |
|---|---|---|
| `a:b_test.go:12: Foo{...} omits Now` (Now backtick-quoted) | `1|a:b_test.go:12:` / identical | LIMITATION: colon prevents sed match; source awk also keeps only the first word. |
| Type `pkg.Foo` | `1|a_test.go|pkg.Foo|Now` / identical | Right. |
| No trailing `-> falls back` | `1|a_test.go|Foo|Now` / identical | Right; neither requires the suffix. |
| Snapshot count `999999999999999999999999999999` | source exit 0 plus `integer expression expected`; Lictor exit 2, `CANNOT-EVALUATE: malformed snapshot count for a_test.go|Foo|Now` | Right under existing documented malformed-count refusal; source is unsafe here. No weakening to reproduce its green. |
| NBSP inside `a\u00a0b_test.go` | source complete triple; Lictor `1|a` → complete triple | DEFECT fixed: Go Unicode whitespace split a path the source preserves. |
| CR inside `a\rb_test.go` | source complete triple; Lictor `1|a` → complete triple | Same DEFECT, fixed. |
| VT inside `a\vb_test.go` | source complete triple; Lictor `1|a` → complete triple | Same DEFECT, fixed. |
| FF inside `a\fb_test.go` | source complete triple; Lictor `1|a` → complete triple | Same DEFECT, fixed. |

Fix: use the source awk field separators (space and tab after LF splitting),
not strings.Fields' Unicode set. A source-produced unchanged snapshot now
passes for these paths instead of manufacturing a new, truncated finding key.

## Route selection

The selector itself was authored under the charter's LICTOR-4 ruling; it has
no source-script counterpart. Achta still owns the YAML judgement.

| Input | Before → after | Assessment |
|---|---|---|
| Empty workflow directory | empty selected set → same | Right. |
| ci.yml installs rulefloor; ci.yaml installs lictor | both selected → same | Right; no basename collision. |
| Workflow file symlink to outside fixture | `map[lictor:true], error=<nil>` → `path escapes from parent` | DEFECT fixed: repository-controlled link caused an outside read. |
| Workflow directory symlink to outside fixture | `map[lictor:true], error=<nil>` → `workflow directory: ... path escapes from parent` | Same DEFECT, fixed. |

Fix: rooted directory enumeration and workflow opens. Sorted enumeration,
1 MiB read ceiling, patterns and delegated judgements stay unchanged. This
confines Lictor's selector; it is not a sandbox around the external Achta process.

## Proofs and remaining close

Red first: TestAdversarialClockfuse failed with `got "1|a" want` the full
NBSP/CR/VT/FF path; TestAdversarialRoute failed twice with
`workflow outside selected tree was read: map[lictor:true]`.
The same tests pass after the fixes. All 35 table subtests pass; the NUL matcher
case is explicitly a matcher-only measurement because recording filters first.
Binary-output tests add 24 combinations (six byte fixtures, two target names,
exits 0 and 7), each checked twice for exact deterministic output, two full-run
green/red cases, and four invalid metadata cases. All pass.

The original `check OK: \x00` fixture was red against the unmodified recorder:
it copied `evidence: [probe] check OK: \x00` and package-count evidence, rather
than the required binary-output line. The corrected implementation passes while
retaining real exits 0 and 7. The first helper invocation was invalid (missing
the testing flag separator); it is not counted as a red proof. After fixing that
fixture, the original production file was temporarily restored, the correct red
was captured, and the implementation restored and measured green.

Six source record conformances pass, each with an empty byte diff:
digest, commit, red-all-false, red-all-true, stepwise, OSS-read-only.
The last is an external diagnostic record, preserving the inherited
EMPTY-TREE limitation; it is not a consumer witness.

CE live CLI measurement: exit 2, empty stdout:
`lictor v0.4.0: refused — identuum-idp-ce declares LICTOR_VERSION v0.2.0; install the declared version: brew install ozgurcd/tap/lictor`
The brief's expectation of a live Grype evidence line is therefore stale.
No pin was changed or bypassed. A fresh scan saved under Lictor, replayed
through the unchanged Grype policy API at 2026-09-21T12:00:00Z, exits 0 and
is byte-identical to the earlier recorded CE line:

```text
check OK: grype-gate matches=0 fixable=0 allowlisted=0 unfixable=0 severe=0 — subject directory:identuum-idp-ce; config applied (3 exclude(s), db max-allowed-built-age 120h, 1 declared ignore(s)); re-check: 1 declared ignore(s), none lapsed on 2026-09-21; coverage: inventory 64 component(s) at 3 location(s), none under an ignored path (8 ignored path(s) read from git)
```

Static t.Fatal/t.Fatalf/t.Error/t.Errorf sites in new files (before → after):
cmd/lictor/adversary_test.go 0 → 6; internal/witness/adversary_test.go 0 → 3;
internal/clockfuse/adversary_test.go 0 → 3; internal/route/adversary_test.go
0 → 10; internal/witness/binary_test.go 0 → 11. Existing test files and six
rule-bound bodies are unchanged.

All three fixes meet item 5's release trigger: consumers can produce these
inputs and their printed result or decision changes. The owner authorizes one
v0.4.1 release, formula-only tap commit and brew upgrade; published and installed
checksums belong to the final slice report. No new command, rule row or schema
version was added. Consumer pin movement remains next-slice work.

The binary-output policy is the first explicitly ruled recorder divergence.
The already-shipped overflow-count exit-2 refusal is the second declared source
divergence, unchanged in this slice. Calling this the first divergence across
all ports would overlook that existing behavior.

## Queue line handed to the owner; no workspace wiki edit

- gate-witness.sh records `Binary file <mktemp path> matches` for NUL-bearing target output, leaking a random temporary path and losing the evidence line; Lictor v0.4.1 deliberately diverges under the binary-output ruling with one deterministic text marker and the real target exit, so the surviving source recorder needs its own disposition. — wiki/tools/gate-witness.sh:200; full init/step/finalize comparison in LICTOR-6, 2026-09-21

Open/adjacent limitations above are recorded here; external filing is declined
because the workspace wiki is read-only. Reader migration and consumer pin
movement remain separately authorized work. No new reader judgment is claimed.
