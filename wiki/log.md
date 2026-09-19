# Lictor change log

## 2026-09-19 — LICTOR-0

Created the standalone Identuum executor and ported OSS's Grype judge at
1cbe9f6c1df8dee83f8ba93e66217cf170b454a7. The source policy remains intact;
CLI selection, bounded subprocesses, JSON contracts and explicit suppression-date
input surround it. Both migrated rules have fresh mutation observations.

The first stop identified the charter's clock contradiction. Amendment 1 made
clock inputs explicit; the second stop established that Grype's internal database
clock cannot be overridden. Amendment 2 confines as_of to Lictor's own judgement
and uses saved reports as replay inputs. No scanner threshold is changed.
See docs/lictor-0.md for measurements, adaptations, validation and limits.
No consumer, parent wiki, tag, release or Homebrew formula is changed here.

Adjacent findings are recorded in docs/lictor-0.md. Filing into the parent wiki
or tool queues is declined here because only Lictor is writable: Rulefloor
counts ignored module-cache sources, and Gograph found an ancestor session before
a local graph existed. Neither generic tool is changed by this slice. Consumer
adoption and workspace-wiki closure remain separate authorized work.
The final Gograph review also found package-qualified Run/run source selection
ambiguous; exact file reads completed the review. Filing that tool finding is
declined here under the same read-only sibling scope.
