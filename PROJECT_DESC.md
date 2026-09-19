# lictor — project description

**Status:** the judge's file (PM-owned, like `identuum/TECH_STACK.md`). Agents read it; they do not edit it. Written 2026-09-19; amended 2026-09-19 four times (§3 determinism: the clock is an input; an external tool's clock is its own; §3a technology stack added, six sentences made exact after LICTOR-0's two stops; §4 row 1 marked PORTED at bc51fed and its line count re-measured). Everything a slice brief does not restate is here; a brief that says "PROJECT_DESC.md applies" means this whole file.

## 1. What lictor is, in one paragraph

The Identuum workspace has two kinds of tooling. The **generic tools** — achta, rulefloor, gograph, legattus, scrinium — judge, never execute, and carry nothing Identuum-specific (owner rule, 2026-09-17: "these tools are generic and for any project"). Beside them sits a set of **Identuum-specific gate programs and scripts** that today are shared by copy (`gate-witness.sh` byte-identical in four repositories under digest mirror checks, the `witness` recipe byte-identical in two under `WITNESS_MD5`, `rulefloor-install-gate.sh` mirrored in three) or by sibling path (`identuum-idp-oss` runs `../wiki/tools/repo-green-gate.sh` and `clockfuse-gate.sh`; identuum-idp-ce and identuum-ui run `go run -C ../identuum-idp-oss ./tools/grype-gate`). Three failures of the week of 2026-09-15 traced to that sharing model: a retired wiki file killed the witness recipe in two repositories (wiki 46d6bf3 → OSS 8148e32, ui 610c60a); Legattus's borrowed commit trees have no siblings, so the sibling-path gates went red for a reason the commit did not contain (THE-COMMIT-JUDGED-IN-A-BORROWED-TREE); one judge change moved three evidence lines (OSS 1cbe9f6). **lictor** is the one versioned, brew-installed binary that replaces that sharing. Its charter is the exact complement of achta's: **it executes, and it knows Identuum.**

A lictor was the Roman magistrate's attendant who carried the fasces and carried out the magistrate's orders, with no judgement of his own. Here the magistrates are achta, rulefloor, gograph and legattus (and the PM); lictor runs the gates and applies the house policy.

## 2. Boundary (binding)

- lictor **never re-judges** what a generic tool already judges. It calls achta, rulefloor, gograph, legattus and scrinium and consumes their versioned outputs. If a judgement is missing from a generic tool and is generic in nature, the fix is a queue line for that tool, not code in lictor.
- Nothing generic enters lictor; nothing Identuum-specific enters the generic tools. The test: "would a second, unrelated Go/TypeScript workspace want this?" Yes → generic tool. No → lictor.
- lictor **executes**: it may run `make` targets, `go`, `grype`, `docker`, test binaries and the generic tools, through `exec.CommandContext` with a fixed argv, no shell, a bounded environment allowlist, output limits and timeouts (legattus PROJECT_SPEC §9). It never mutates Git (a witness COMMIT stays in the consumer's Makefile recipe; lictor writes the record and stops), never touches the network on its own, never reads secrets or local credential files.
- lictor's own knowledge lives in **its own repository wiki** (`lictor/wiki/`), co-versioned with the code, exactly as legattus and achta keep theirs. It does not write the workspace wiki; the workspace wiki records lictor only through its repository page and the ledger rows the next wiki slice owes any commit (workspace `agent-rules.md` §F-bis).

## 3. Requirements — legattus's, applied

Read `legattus/PROJECT_SPEC.md` §1 (Boundary) and §9 (Security and operational limits) first; they apply verbatim except where this file is narrower.

- **Deterministic.** No LLM, no natural-language inference, no randomness, no hidden inputs: the same inputs give the same output. The wall clock is an INPUT, not hidden state (ruling 2026-09-19, after LICTOR-0 stopped on it; narrowed the same day after its second stop): `--as-of <RFC3339>` governs the dated judgements LICTOR ITSELF makes — a suppression's re-check date that has lapsed, any expiry lictor evaluates — read ONCE per run, printed on the evidence line and in the JSON document (`as_of`), so any such verdict can be reproduced exactly; tests use the override, never the real clock. An EXTERNAL tool's own clock is that tool's (grype validates its database age inside its own process against `time.Now()` and exposes no evaluation-time input — curator.go, measured at v0.119.0): lictor neither drives it nor imitates it; it applies the tool's declared configuration unchanged (grype-gate's CompareConfig asserts `db.validate-age` and `db.max-allowed-built-age` effective == declared, and that stays), and the reproducible unit is the saved report — a judgement over a saved scan (`-scan <json>`) with a given `--as-of` is a pure function, and that is what tests and replays use. Without `--as-of`, a ported program's evidence line stays byte-identical to its source's (the source already prints the date: `re-check: … none lapsed on 2026-09-19`). Freezing or removing a dated judgement to look deterministic is forbidden — that changes the verdict; naming the date is what makes it deterministic.
- **Request-scoped repository selection.** Every repository-dependent command takes `--repo <absolute path>`. There is no mutable process-wide "current repository". When `--repo` is omitted the documented compatibility default is the process current directory, and every output (human and JSON) names the repository it judged so no run is ambiguous. Sibling repositories a command needs are further explicit flags (`--sibling NAME=<abs path>`), never discovered by walking `..`.
- **Exit contract:** 0 pass, 1 fail, 2 cannot-evaluate. A tool that could not measure never exits 0. Make flattens exit codes; the evidence line and the JSON document carry the distinction.
- **One JSON document per command** under `--json`, schema-named `lictor.<command>.v1` and published under `schemas/`; the human line and the JSON say the same thing. `lictor version` and `lictor capabilities --json` (`lictor.capabilities.v1`) are repository-independent.
- **Evidence lines are contracts.** Where lictor replaces a program whose line other repositories record (`GATE-RUN.txt`, `CI-WITNESS.txt`) or whose text RULE-FLOOR rows quote, the line stays byte-identical, including the historic program token (for example `grype-gate`), until a deliberate, ledgered change moves every consumer at once.
- **Rules travel with code.** Each ported program brings its RULE-FLOOR rows (rule text from the source repository's `RULE-FLOOR.md`) bound to the ported tests in `lictor/RULE-FLOOR.md`; `rulefloor check` is part of `make verify`. Consumers then bind to a lictor **version pin**, not to a file digest.
- **Shape = legattus:** `cmd/lictor`, `internal/<command>` one package per command, `schemas/`, `Makefile` with `format-check build test race vet staticcheck govulncheck tidy-check verify`, `.github/workflows/verify.yml` and `release.yml`, `PROJECT_SPEC.md`, `AGENTS.md`, `README.md`, `RELEASE_NOTES.md`, `RULE-FLOOR.md`, `wiki/` (`WIKI.md`, `index.md`, `agent-rules.md`, `repos/lictor.md` with `co_versioned: true`, `log.md`). Every commit updates the co-versioned wiki in the same commit — achta's `AGENTS.md` states that rule; legattus's `AGENTS.md` adds `make wiki-check` for wiki changes and "never use a parent or sibling wiki as a fallback". Both apply.
- **Release shape = legattus's:** a tag `vX.Y.Z` whose version equals `lictor version --json`'s; `release.yml` builds reproducible archives for darwin/arm64, darwin/amd64, linux/arm64, linux/amd64 (`CGO_ENABLED=0`, `-trimpath -buildvcs=true -ldflags="-s -w"`, tar sorted and mtime-fixed to the commit time, `checksums.txt`), publishes a GitHub release with `RELEASE_NOTES.md`; then a one-file formula commit in `ozgurcd/homebrew-tap` (`Formula/lictor.rb`, the rulefloor.rb shape, private download via `HOMEBREW_GITHUB_API_TOKEN` as legattus's README §Install describes), `brew upgrade`, checksum verified. The owner authorizes each release explicitly; no slice tags, releases or edits the tap on its own.
- **Repository:** `github.com/ozgurcd/lictor`, private, branch `main`, local checkout `/Users/odemir/Development/identuum/lictor`. Author of every commit: `Ozgur Demir <ozgurcd@gmail.com>`; plain-prose commit bodies, no generated attribution trailers.

## 3a. Technology stack and repository constraints (after legattus `project_plan.md` §23)

These are constraints unless the owner changes them explicitly.

**Language.** Go, the version `go.mod` declares — 1.27.1 at creation (the workspace's `TECH_STACK.md` §1 pin; legattus and achta declare the same) — and never below it. Simple, idiomatic Go; the standard library before any third-party dependency; a third-party module only with a concrete justification written in the commit that adds it; the dependency surface small and auditable (`go mod tidy` clean, `tidy-check` in verify).

**Repository.** Local `/Users/odemir/Development/identuum/lictor`; remote `https://github.com/ozgurcd/lictor` (private); module `github.com/ozgurcd/lictor`; branch `main`. The parent `identuum/` directory is not a Git repository and is never assumed to be one; every Git operation is scoped to `lictor/` itself.

**Identifiers.** No UUIDs unless a design needs them; if ever, UUIDv7 only. Prefer identities derived from canonical content digests (SHA-256) where reproducible identity is the actual requirement — records, reports, evidence.

**Database.** None. lictor's state is files: the consumer's tracked records and its own saved reports. If durable queryable local state is ever required: SQLite without CGO, native SQL, no ORM, explicit schema and migrations, canonical identities kept outside row IDs — and the case for it written first.

**Libraries.** `encoding/json`, `crypto/sha256`, `os`, `io`, `io/fs`, `path/filepath`, `context`, `time`; `os/exec` (always `exec.CommandContext`, fixed argv) for the tools lictor runs; `flag` or the standard-library-shaped CLI layer legattus uses. No HTTP client in the tool (lictor makes no network calls); `net/http` only if a future command's contract says otherwise, and that would be a §2 change.

**Toolchain the repository asserts on itself.** `gofmt` (format-check), `go vet`, `staticcheck` at the workspace pin (v0.8.1 today), `govulncheck` at the workspace pin (v1.7.0 today), `go test -race`, `rulefloor check` over `RULE-FLOOR.md`, `achta` for `make wiki-check` over the co-versioned wiki. Pins are read from the workspace `TECH_STACK.md` §1 and the consumers' workflow envs; when they move, lictor's `verify.yml` moves with them in the same week.

**Tools lictor calls at run time** (never re-implements): `grype` (at the consumer's declared `GRYPE_VERSION`), `go`, `docker` (image subjects only), `achta`, `rulefloor`, `gograph`, `legattus` — each through `exec.CommandContext`, each version printed on the evidence line where the consumer's toolchain-parity already prints it, each absence an exit-2 refusal by name.

**Developer tools.** Gograph before Unix text tools for every Go symbol question (callers, callees, implementations, impact, tests, architecture); a Gograph session, plan and review for structural changes; `grep`/`sed`/`awk` only outside Gograph's capability domain. `make verify` is the repository's own gate and runs before every commit.

**Layering** (separable in code even though it ships as one binary):

```text
CLI (cmd/lictor: flags, --repo, --json, exit codes)
            |
            v
command packages (internal/<command>: one policy each)
            |
            v
executors / tool adapters (internal/exec…: fixed-argv process runs, report parsing)
```

Command packages import no other command package; adapters import no command package; the CLI imports only what it dispatches to.

## 4. Migration table

Source of truth for the inventory: `wiki/contracts/retirement-ledger.md` (rows 5–31); "Lines" are the ledger's counts at its last census and drift with every edit (grype-gate: 1659 since OSS 1cbe9f6, re-measured in the ledger by wiki e7e0a87 on 2026-09-19; lictor's port of it, at that SHA, is 1,659 lines) — the ledger, not this table, is authoritative, and a port measures its source at the SHA it names. "Why not achta" is the ledger's own replacement column. Order is by pain, not size: first what is shared by path or copy today.

| Order | Program | Source | Lines | Rules bound | Why not achta | lictor command (proposed) |
|---|---|---|---|---|---|---|
| 1 | grype-gate | identuum-idp-oss `tools/grype-gate` | 1659 | GRYPE-FIXABLE-FAILS-1, GRYPE-SUBJECT-1 | vulnerability policy, outside achta's boundary by its spec; three consumers today | `lictor grype --repo` — PORTED, lictor bc51fed (LICTOR-0, 2026-09-19), unreleased; consumers not yet switched |
| 2 | repo-green-gate | wiki `tools/repo-green-gate.sh` | 366 | none | executor (build/vet/gofmt/test now); achta declares no shell execution | `lictor green --repo` |
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

## 5. How a consumer switches (one repository per slice)

1. lictor released at a version carrying the command; `brew upgrade` on the workspace machine; each consumer's CI installs that exact version the way its workflows install rulefloor today (one declared `LICTOR_VERSION` and `LICTOR_SHA256` in the workflow env, one derived download route, the version printed and asserted after install — the discipline `rulefloor-install-gate` enforces for rulefloor). A private release needs a token in CI; whether lictor's releases are private is the owner's decision (§7) and is settled before the first consumer switches.
2. The consumer's Makefile target calls `lictor <command> --repo "$(CURDIR)"` where it called the script, the sibling program or the copy; the evidence line is unchanged, so `GATE-RUN.txt`, `CI-WITNESS.txt` and every RULE-FLOOR row that quotes it keep meaning the same thing.
3. The consumer's toolchain-parity pins the lictor version (`LICTOR_VERSION` in the workflow env, asserted locally and on the runner), replacing the file digest pin where one existed.
4. The mirrored copy or the sibling-path call is deleted in the same slice; the workspace wiki's retirement ledger row gets its RETIRED date and the proof, in the next wiki slice.
5. The last consumer's switch retires the source program.

Legattus commit mode on a consumer is the acceptance test of every switch: a borrowed tree with no siblings must judge green or red for reasons the commit contains — never "No such file or directory" one repository over.

## 6. What lictor is not

Not a runner for arbitrary commands (no `lictor exec`); not a wiki writer; not a replacement for `make` (it is called by `make`, one target per command); not a place for product code, migrations, or anything that ships in an Identuum image; not a second copy of any generic judgement.

## 7. Reports and authority for lictor slices

- Report shape: one plain-text block, first line `Final Report — <slice-id>`, last line `end of report for <slice-id>`; files created or changed; the commands run with their outputs quoted (`make verify`, `rulefloor check`, `lictor capabilities --json`); test files touched with assertion counts before/after; commits and HEAD (ahead/behind origin); an explicit no-touch confirmation per read-only sibling; errors found in the brief, named as the brief's; OPEN AND ADJACENT.
- Authority: writes only inside `lictor/`; siblings are read-only and their porcelain is quoted before and after any read-only run; no push, tag, release or tap edit unless the slice brief says so in words; no amend, rebase, force.
- Every stop is a report: a STOP that names the line that stopped it is a finished slice.
