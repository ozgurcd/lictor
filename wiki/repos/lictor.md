---
title: Lictor
status: authoritative
co_versioned: true
verified: 2026-09-19
---

# Lictor

Module: github.com/ozgurcd/lictor. Release v0.1.0 uses the port implemented in bc51fed, unchanged.
Lictor executes Identuum gates; generic tools retain their judgements. The first
command is grype, ported from OSS 1cbe9f6c1df8dee83f8ba93e66217cf170b454a7.
CLI, policy and fixed tool execution occupy separate packages. make verify is
the canonical gate; no consumer is modified or automatically discovered.

## Commit ledger

| Date | Commit | Change |
|---|---|---|
| 2026-09-19 | co-versioned | LICTOR-0: scaffold and Grype port, explicit evaluation time, bounded executors, schemas, migrated rule proofs and consumer evidence comparison. |
| 2026-09-19 | co-versioned | LICTOR-1: dated release notes and owner-authorized annotated v0.1.0 tag on this release commit; port and tests unchanged. |
