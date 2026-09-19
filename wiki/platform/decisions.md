# Lictor decisions

The owner-controlled decisions are in [PROJECT_DESC.md](../../PROJECT_DESC.md).
The implemented contracts are in [PROJECT_SPEC.md](../../PROJECT_SPEC.md).

- 2026-09-19, LICTOR-0: Lictor owns Identuum execution policy; generic tools
  retain generic judgement. The first migration is the OSS Grype judge.
- 2026-09-19, amendments 1 and 2: the explicit as_of input governs Lictor's
  dated judgements only. External tool clocks and applied configuration remain
  unchanged. Replays judge saved reports and unchanged repository inputs.
- LICTOR-0 authorizes initial commits and one main push, with no release,
  consumer migration or Homebrew change. Each release needs separate authority.
