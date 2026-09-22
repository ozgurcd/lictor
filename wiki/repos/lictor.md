---
title: Lictor
status: authoritative
co_versioned: true
verified: 2026-09-22
---

# Lictor

Module: github.com/ozgurcd/lictor. Release v0.3.0 adds runtime pin enforcement,
clockfuse and delegated installation routes; the Grype policy is unchanged.
Lictor executes Identuum gates; generic tools retain their judgements. The first
command is grype, ported from OSS 1cbe9f6c1df8dee83f8ba93e66217cf170b454a7.
CLI, policy and fixed tool execution occupy separate packages. make verify is
the canonical gate; no consumer is modified or automatically discovered.

## Commit ledger

| Date | Commit | Change |
|---|---|---|
| 2026-09-19 | co-versioned | LICTOR-0: scaffold and Grype port, explicit evaluation time, bounded executors, schemas, migrated rule proofs and consumer evidence comparison. |
| 2026-09-19 | co-versioned | LICTOR-1: dated release notes and owner-authorized annotated v0.1.0 tag on this release commit; port and tests unchanged. |
| 2026-09-19 | co-versioned | LICTOR-1 amendment 1: dispatch the existing immutable release tag, restore its runner-local annotation and assert its target equals HEAD; commit the unchanged owner charter amendment. |
| 2026-09-21 | co-versioned | LICTOR-2: document public Homebrew and CI release downloads without a token; preserve v0.1.0 code, schemas and release artifacts. |
| 2026-09-21 | co-versioned | LICTOR-3 charter: commit the owner's unchanged Row-2 ruling before implementing the green command. |
| 2026-09-21 | co-versioned | LICTOR-3 implementation: v0.2.0 adds green, bounded Go execution, lictor.green.v1 and GREEN-FLOOR-1; eight source fixtures agree and OSS/CE pass without porcelain changes. |
| 2026-09-21 | co-versioned | LICTOR-4 charter and stop: preserve the owner amendment unchanged; Achta cannot enforce required-key-only without a route match, so the UI route proof needs a generic-tool change before this port can finish. |
| 2026-09-21 | co-versioned | LICTOR-4 Amendment 1 charter: preserve the owner amendment unchanged; runtime pin enforcement is fail-closed and route sets apply only where a tool is installed. |
| 2026-09-21 | co-versioned | LICTOR-4 Amendment 1 implementation: v0.3.0 adds fail-closed pins, explicit non-consumer unpinned use, isolated Go inputs, platform-independent green evidence, clockfuse and Achta route delegation; five armed mutation-proved rules, six commands, consumer read-only proofs and authorized release. |
| 2026-09-21 | co-versioned | LICTOR-4 Amendment 1 final context review: clockfuse and route name the selected repository on stderr; native stdout and one-line pin refusals stay unchanged. |
| 2026-09-21 | co-versioned | LICTOR-5 charter and stop: preserve the owner amendment unchanged; the required WITNESS-ONE-RUN-PER-RECORD-1 sentence judges records through check, which this execute-only slice forbids porting. No recorder code or release. |
| 2026-09-21 | co-versioned | LICTOR-5 Amendment 1 charter: preserve the owner amendment unchanged; WITNESS-ONE-WRITER-1 replaces the reader rule and binds truncation, bounded writer exclusion and stepwise-session refusal. |
| 2026-09-21 | co-versioned | LICTOR-5 Amendment 1 implementation: v0.4.0 adds argv witness execution, source-identical records, dirty-work no-mint, writer locks/sessions and 64 MiB output handling; six rules, seven commands, consumer read-only conformance; the reader remains outside the port. |
| 2026-09-21 | co-versioned | LICTOR-6 Amendment 1: preserve owner charter 6039414b byte-unchanged; v0.4.1 preserves source clockfuse path bytes, confines route reads, and records binary output deterministically with real exits; six record conformances and six armed rules retained; declared source divergences and queue handoff in docs/lictor-6.md. |
| 2026-09-22 | co-versioned | LICTOR-7 charter: preserve the owner's unchanged dirty-tree all-target ruling before implementation; fail-fast run and record judging remain outside this change. |
| 2026-09-22 | co-versioned | LICTOR-7 implementation: v0.4.2 echoes dirty --all scratch records with source-identical streams; green, red and NOT-RUN proofs preserve the committed record; WITNESS-DIRTY-ECHO-1 raises the armed floor and observed red proofs to seven. |
