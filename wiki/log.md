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

## 2026-09-19 — LICTOR-1 amendment 1

The first Release run failed because checkout fetched the tag's target commit
directly into its runner-local tag ref. GitHub's annotated v0.1.0 object
9ebd6ae2eaf3e157a913fe29606506931eca9504 remained intact at release commit
a1bee4dfb50a014b6cdf6aac4141731fc9b8a6a8.

Added workflow_dispatch with an explicit existing-tag input and tagged-source
checkout. The runner restores that remote tag object before requiring annotation,
matching tag target and HEAD, matching source version and passing tests. Archive
and publication commands are unchanged. The owner authorized dispatching this
workflow for v0.1.0 without moving or recreating the tag.

The owner's PROJECT_DESC.md amendment is committed unchanged at SHA-256
6875314ec4219041765f4267237764cfed84fd3f3a9cbf1d5ab92e48c369abd7. Publication,
published checksums, the single authenticated Homebrew formula entry and install
proof are reported after the workflow. No consumer or workspace-wiki edit is
authorized. The consumer pin and authenticated CI install route remain the next
switch's work; external queue filing is declined under this slice's scope.

## 2026-09-21 — LICTOR-2

Measured the public v0.1.0 checksums URL with curl in an empty environment and
curl configuration disabled: HTTP 200. The unauthenticated GitHub API also
reports private=false. README installation instructions now use the public
Homebrew route and link the release archives and published checksums. The tap
Formula switches from authenticated asset API URLs to the corresponding public
release URLs; all four checksums are re-read from the public checksums.txt and
are equal to the previous values. The existing Formula test is retained.

Rulefloor and Scrinium declare MIT in their Formulae. Lictor retains
license :cannot_represent because its NOTICE records AllRightsReserved; making
the repository public does not change that declaration. RELEASE_NOTES.md has no
private-distribution statement and remains unchanged. The owner charter is held
unchanged under this slice's restricted commit scope. No code, schema, tag or
release is changed. The no-token reinstall used a fresh cache, passed its published archive checksum,
and passed the unchanged Formula test and style check. Installed bytes match the
archive. CE porcelain stayed empty before/after both installed-binary proofs.
The current command reports 2026-09-21; --as-of 2026-09-19T00:00:00Z reproduces
LICTOR-1's evidence byte-for-byte. The date is the only live-line difference.
Homebrew's reinstall log printed the archive filename rather than the full URL;
its info JSON supplied the selected public download URL.

OPEN AND ADJACENT: OSS needs LICTOR_VERSION=v0.1.0, a public installation route
and LICTOR_SHA256=c051b5fa995211de4b4910395dc10b46d7b7c05be32c6af87e65345bd4eca739
for lictor_0.1.0_linux_amd64.tar.gz, taken from published checksums.txt.
PROJECT_SPEC.md still says to resolve private-release access before switching;
PROJECT_DESC.md's older release-shape bullet still names private authentication,
while its updated public-repository ruling and this brief settle the public
route. Those files are outside this edit scope. External queue filing is declined
because the workspace wiki and consumers are read-only.

## 2026-09-21 — LICTOR-3 charter

Committed the owner's PROJECT_DESC.md unchanged before the green port, SHA-256
c5cf54faa3b45e55e197fca70efceece8413e11cda27937552afc102127d09f3.
The Row-2 ruling preserves the source script's ordering, cause phrases and
three outcomes, mapping its cannot-evaluate exit 3 to Lictor's exit 2. It excludes
hook mode and bypass. The owner authorizes v0.2.0, the public Formula update and
local upgrade in this slice. No implementation changes belong to this commit.
