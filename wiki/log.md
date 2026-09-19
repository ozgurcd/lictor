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

## 2026-09-19 — LICTOR-1

Prepared v0.1.0 for the first owner-authorized release. This commit dates the
release notes and records the annotated v0.1.0 tag to be placed on this commit;
the port, tests and release workflow remain byte-identical to bc51fed.
The owner authorized a fast-forward main push, one annotated tag push, the
release workflow, one Homebrew tap entry commit and local installation.
Publication, archive checksums, installation and the read-only CE proof are
measured after tagging and reported in LICTOR-1's final report.

The authoritative tap's Legattus entry is a Formula with GitHub asset API URLs
and an Authorization header supplied by HOMEBREW_GITHUB_API_TOKEN. Lictor follows
that existing private-download shape. No consumer is switched in this slice.

OPEN AND ADJACENT: the first consumer switch needs LICTOR_VERSION and a published
archive checksum pin, a single authenticated CI download route with explicit
private-release access, an installed-version assertion and replacement of its
sibling-path call. Filing into the workspace wiki is declined because it is
read-only. README.md still describes the LICTOR-0 unreleased state; its broader
installation prose is left for an authorized documentation change under this
slice's restriction to release notes and required wiki records. The owner's
PROJECT_DESC.md amendment was present at startup and is left held unchanged.
