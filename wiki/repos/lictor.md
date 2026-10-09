---
title: Lictor
category: repo
status: authoritative
co_versioned: true
verified: 2026-10-09
---

# Lictor

Module: github.com/ozgurcd/lictor. Release v0.4.5 adds consumer-declared gate
environment names to witness run, name/presence evidence, named refusals and
streaming value redaction. Runs without declarations retain their record bytes.
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
| 2026-09-30 | co-versioned | agent-lictor-sbom: explicit SPDX/CycloneDX subject, basename and exact-byte SHA-256 evidence, directory predicates not applicable, unchanged vulnerability policy and fail-closed unbound replay; GRYPE-SBOM-1 and real scanner measurements in docs/sbom.md; Unreleased only. |
| 2026-09-30 | co-versioned | agent-lictor-patch-release: v0.4.3 version and schema constants, SBOM README usage and dated release notes; owner-authorized main/tag publication and Homebrew formula/install, with final remote and installation evidence in the release report. |
| 2026-10-02 | co-versioned | LICTOR-8: v0.4.4 prints clean --all blocked-dependent banners and NOT-RUN reasons in plan order; mutation-proved rule and frozen-clock source/baseline comparisons in docs/lictor-8.md; authorized release and installation. |
| 2026-10-05 | co-versioned | LICTOR-9: v0.4.5 adds repeatable witness run environment-name declarations, presence-only records, controlled-name refusals and pre-output redaction; offline and mutation proofs plus undeclared v0.4.4 byte comparisons in docs/lictor-9.md; authorized release verification reported separately. |
| 2026-10-05 | co-versioned | ACHTA-0.5.13 + LICTOR PIN: align the local Achta predicate, CI version and published Linux amd64 checksum, and live validation documentation with v0.5.13; no Lictor release. |
| 2026-10-06 | co-versioned | ACHTA-CLAIM: pin local and CI validation to installed Achta v0.5.14, with its published Linux amd64 checksum and current validation docs; no Lictor tag or claim adoption. |
| 2026-10-06 | co-versioned | ACHTA-0.5.15: advance the local predicate, CI version and published Linux amd64 checksum, and live validation docs to Achta v0.5.15; no Lictor release or code change. |
| 2026-10-06 | co-versioned | LICTOR-10: accept compatible stable local Achta and Rulefloor patch versions with agreement pass; strict JSON/range selftests join verify; exact CI pins and consumer route comparisons stay unchanged. |
| 2026-10-07 | co-versioned | LICTOR-11: v0.4.6 leaves declared values shorter than eight bytes unchanged with explicit presence evidence; preserves long-value streaming redaction and adds checked Homebrew formula publication to the release workflow. |
| 2026-10-09 | co-versioned | ACHTA-GATE-RUN security prerequisite: advance Go to 1.27.2 for GO-2026-6604; local and CI use go.mod; keep every validation predicate. |
| 2026-10-09 | co-versioned | ACHTA-GATE-RUN: pin CI to published Achta v0.5.16 and its Linux amd64 checksum; retain the local v0.5.15 through below-v0.6.0 compatibility range; no Lictor release. |
| 2026-10-09 | co-versioned | ACHTA-GATE-RUN postcheck: declare the repository page category so freshness evaluates its existing co-versioned contract instead of skipping it. |
| 2026-10-09 | co-versioned | GO-1.27.2-TOOLS: build CI staticcheck from the pinned brew-identical source and patches so Go 1.27.2 export data is supported; no version or gate relaxation. |
| 2026-10-09 | co-versioned | ACHTA-JOURNAL-3: pin CI to published Achta v0.5.17 and its Linux amd64 checksum; retain local compatibility bounds and all product behavior. |
