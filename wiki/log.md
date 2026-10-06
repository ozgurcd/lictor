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

## 2026-09-21 — LICTOR-3 implementation

Added green as the second gate command, porting green_check and eight non-hook
fixtures from wiki 662585b's 366-line repo-green-gate.sh. The floor still runs
gofmt, build, vet and tests in that order, stopping at the first red and keeping
-count=1 -timeout=120s. Source exit 3 maps to Lictor exit 2. A separate Go adapter
bounds execution without altering the Grype package or adapter. Evidence names
the base-name subject and Go version; JSON publishes lictor.green.v1.

The CLI contract test first failed against the old dispatcher (unknown command:
green). A fresh mutation that ignored every step failure made five fixture
cases incorrectly GREEN, and all five failed. The restored port agrees with
the script on all eight cases; an uncompilable test fails at vet on Go 1.27.1.
GREEN-FLOOR-1 binds these tests. Consumer proofs passed on OSS 63ee215 and CE
128dc78; every sibling's before/after porcelain and head were unchanged.

The first source-comparison invocation passed its package-local flag to every
package through make test, which rejected it. The corrected diagnostic selected
internal/green only. Normal make test uses no custom flag. Initial Rulefloor
arming refused the lowercase annotation; the required // RULE: annotation was
added before arming. These were setup errors, not hidden green runs.

PROJECT_SPEC.md now states public release URLs and checksums, without private
access or a token. v0.2.0 updates version/capabilities contracts and dated release
notes. Publication and installed-byte proof occur after this commit under the
owner's explicit tag, push, tap and upgrade authority, and are in the final report.
See docs/lictor-3.md for test predicates and measured evidence.

OPEN AND ADJACENT: OSS and CE still need their own authorized switches to green,
version assertions/pins and the runner's published platform checksum. Once the
runner installs that release, repo-green no longer needs the wiki sibling and
can run in CI. Neither consumer's plan changes here. The source's hook and wiki
selftest remain; the last consumer's retirement is separate work. External queue
filing is declined because all sibling repositories, including the wiki, are
read-only in this slice. Gograph source lookup required exact current symbols;
two nonexistent symbol queries refused and are counted in its audit.

## 2026-09-21 — LICTOR-4 charter and stop

Committed the owner's PROJECT_DESC.md byte-unchanged first, SHA-256
251ea74c26501a7e4f3450b6d90a30badc837190f9e7943203205e9d1e73db8c.
No implementation, version bump, rule change, release, push, tag or tap change
belongs to this stopped slice. Runtime and installed version remain v0.2.0.

STOP: the Row-4 ruling requires UI's declared-but-not-installed Lictor route to
be required-key-only. Installed Achta v0.5.10 has no such mode. Its source at
de3395b, internal/cli/declared_route.go, exposes one required-key option tied to
a required-route-pattern. internal/declaredroute/check.go sets Required=true
only when that route pattern has a nonzero count; key scope and document-wide
uniqueness are checked only inside `if documentResult.RequiredKey.Required`.

Measured on scratch YAML under Lictor with the exact proposed Lictor route and
three bans, per-file-any, required key LICTOR_VERSION and scope /env:
one declared key with no install, a key duplicated at workflow and job scope
with no install, and no key with no install ALL exit 0. All report
`"status":"pass"` and `"required":false,"scope":"/env","key":"LICTOR_VERSION","count":0`.
An optional route therefore does not produce the requested required-key-only
judgement. Fabricating a matching route or implementing YAML key judgement in
Lictor would evade the boundary rather than supply the missing Achta capability.

The exact stopping charter sentence is: "If a judgement is missing from a
generic tool and is generic in nature, the fix is a queue line for that tool,
not code in lictor." External queue filing is declined because the workspace
wiki and Achta are read-only. The missing generic capability needs a separate
Achta slice; this log preserves the finding for the owner.

Source measurements: wiki is now e2e7b44, not the source reference 662585b;
both requested scripts and the Achta pattern set are unchanged between them.
The wiki's repo-green selftest was retired meanwhile, so the charter's Row-2
statement that the wiki still runs it is stale. The charter is owner-controlled
and was not corrected here. OSS e642f92, CE 56a6675 and UI 4371909 all declare
LICTOR_VERSION v0.2.0. UI declares no LICTOR_SHA256 and installs no Lictor binary.
The brief also needs to specify which key(s) its UI required-key-only case
requires, since its installation set names both version and checksum.

The first scratch Achta invocation lacked its required wiki workspace scaffold
and refused; selecting Lictor's existing workspace corrected the setup before
the semantic measurements above. Achta source was read with git show without
building or writing a graph in that read-only repository. Gograph discovery
and source reads in Lictor changed no Go code; plan/review are not claimed.

OPEN AND ADJACENT: after Achta supplies the missing judgement, the resumed
Lictor slice still owes all implementation and v0.3.0 release proofs. Its future
consumer slice must move all version pins/checksums together, replace the
LICTOR_ASSERT macro with the absent-tool check plus Lictor's own self-check,
switch clock-fuse-gate, add route targets, install pinned Achta on OSS/UI runners,
and retire the vendored installation-gate copies and their digest pins. These
are recorded here rather than filed into read-only sibling queues.

## 2026-09-21 — LICTOR-4 Amendment 1 charter

Commit the owner-controlled PROJECT_DESC.md byte-unchanged on top of 274f82f.
SHA-256: cd04f8478605bef875a528883018bda2d5e0ddd882fde34e4baf6099dc240768.
The owner resolves the earlier STOP: Lictor enforces its own version declaration
at runtime, refusing absence unless --unpinned is explicit. Achta judges only
installation route sets selected by non-comment install lines; no required-key-only
judgement is added here. The retired repo-green selftest statement is corrected
by the owner. This charter commit contains no implementation or release change.
The implementation still owes pin, environment, line, clockfuse and route proofs,
the full gate, v0.3.0 publication and Homebrew installation.

## 2026-09-21 — LICTOR-4 Amendment 1 implementation and v0.3.0

Charter commit d9db9b1 preserved SHA-256
cd04f8478605bef875a528883018bda2d5e0ddd882fde34e4baf6099dc240768.
Every repository command now checks the first exact LICTOR_VERSION line in
ci.yml. Absence and mismatch refuse before execution; --unpinned permits only
absence. The four required cases and the mismatch-with-unpinned case were red
on old code and pass now. Existing fixture declarations were made explicit.
Grype's policy, executor environment and CE evidence line remain byte-identical.

Go execution drops GOFLAGS and sets GOWORK=off and GOENV=off. The adversarial
fixture initially passed because inherited -tags=nonexistent hid a broken Go
file; now it fails while a clean fixture remains green. Human green evidence
ends in go1.27.1; JSON retains the full version and platform.

Clockfuse preserves all five source fixture outcomes and normalized live data.
OSS currently has 14 classes, not the snapshot's 15: the removed class is
internal/service/mfa_stepup_password_rule_test.go / MFAEnrollmentService / now
(count 2 in the snapshot). CE has 28, exactly matching its snapshot. Both raw
analyzers exit 1; the wrapper deliberately tolerates that exit, as its source
does. Check writes nothing; an explicit scratch snapshot proves all four header
lines and identical finding lines. The analyzer remains consumer-owned.

Route delegates to Achta v0.5.10 without YAML judgement. OSS selects Rulefloor
and both Lictor required-key calls; UI selects Rulefloor and explicitly skips
Lictor. Contrary to the brief, CE installs neither tool and skips both. Two real
Achta fixtures reject go install and a literal release URL with exit 1. Nested
fixtures exposed Achta's all-ancestor workspace discovery: it refuses ambiguity
even though declared-route needs only YAML. Optional --achta-workspace forwards
an explicit absolute selection to Achta; there is no Lictor workspace discovery.
An isolated CI runner without Achta's required wiki layout still refuses: this
is OPEN AND ADJACENT for Achta/consumer adoption, not a duplicated judgement.

CLOCKFUSE-SNAPSHOT-1 hash d63cb3513614 binds new-class/count-rise rejection;
neutralizing the comparison made both tests fail. INSTALL-ROUTE-1 hash
e591453742f1 binds complete selected sets and native result preservation;
forcing every set to skip failed three selection cases and native evidence.
Both mutations were restored byte-exactly. FLOOR and RED-PROOFS are now 5; the
three original rows/hashes and their ported tests are unchanged.

Full make verify passed before release, including race, staticcheck, govulncheck,
five rule proofs and the local wiki. Final-head validation and publication
checksums are reported after the commit; they are not self-referential fields
in this co-versioned record. Release authority is the owner's LICTOR-4 Amendment 1:
main fast-forward, one annotated v0.3.0, the existing workflow, one public formula
commit and local brew upgrade. Four platform archives remain unchanged in scope.

Expected reds: initial pin/environment/line tests, initial absent commands,
the two deliberate rule mutations and two prohibited route fixtures. Setup
errors corrected: an old positional Result literal after adding platform;
a schema-generation Python bracket typo; a route test that inadvertently
created the workflow it expected absent; nested Achta discovery ambiguity.
Gograph symbol/path queries that it could not resolve are counted honestly in
the audit (plan/review true; grade C). Test and proof details are in docs/lictor-4.md.

All siblings remain read-only. Consumer follow-up moves pins/checksums together,
replaces the copied assertion macro with command-v plus Lictor's check, switches
clock-fuse-gate, adopts route, supplies pinned Achta and its workspace prerequisite
on runners, and retires copied scripts/digest pins after the last switch.
No consumer macro, snapshot, gate record, policy or external queue is edited.
Installed v0.3.0 refusing the consumers' v0.2.0 declarations is intentional.

## 2026-09-21 — LICTOR-4 Amendment 1 selected-repository context

Final contract review after the green 9055f5a gate found that native clockfuse
pass text and skipped route lines do not themselves name the selected tree.
The strengthened CLI snapshot/check proof first failed with empty stderr.
The dispatcher now names the selected repository on stderr after pin acceptance,
and prints an explicitly supplied Achta workspace there too. Native stdout is
unchanged and pin refusals still emit exactly their one required line.
This closes the charter's request-selection visibility requirement before the
single v0.3.0 tag; no consumer, rule, version or release scope changes.

## 2026-09-21 — LICTOR-5 charter and record-reader boundary

The owner's PROJECT_DESC.md is committed byte-unchanged first, SHA-256
`a42c5039d6407bfac60e16c7bfd1a4eda284b1b636ed780322a1a504ca5a3074`.
Its four new rulings cover execute-only scope, argv plans, the 64 MiB recorder
ceiling and the gate-run.v1 record contract. No Go implementation is changed.

STOP: required item 6 asks for the exact OSS WITNESS-ONE-RUN-PER-RECORD-1
sentence, measured at OSS e642f92f78e32b95242e83a73038d2fa34292fbe,
RULE-FLOOR.md:253:

> A gate-run record carries ONE run's evidence: check REFUSES any record with more than one gate header or more than one result verdict, naming how many of each it found, so a record two runs wrote into cannot be believed even though the interleaving that produces it is still possible.

The source implements that judgement in gate-witness.sh check_mode:545-550,
not in run/init/step/finalize. Its comments explicitly call it a reader
backstop, not a writer fix. A scratch record inside Lictor with two headers
and two verdicts was measured against the unchanged workspace script:
exit 1, `GATE-WITNESS INTERLEAVED`, naming `2 'gate:' header(s)` and
`2 'result:' verdict(s)`. No consumer record was written.

The brief forbids porting check, --selftest or --sync-check; charter section 2
also leaves generic record judgement to Achta. Writer-lock tests would not
prove the required sentence, and calling the sibling reader from the port's
unit tests would retain the sibling dependency this migration removes.
The exact reader rule cannot honestly be armed against an execute-only port.
An owner ruling must supply the intended execute-side rule or explicitly
change the record-reader scope. The existing floor remains five, with no
changed assertions. No v0.4.0 tag, release, tap update or upgrade is made.

The consumer switch still owes verify-all.sh plan/dependency handling,
GATE_RECORD_DRIVER integration, and the version/checksum pins. The four
script copies remain until check's record judgement moves to Achta.
The argv, record-conformance, refusal and output-ceiling implementation
proofs are not claimed; implementation stops at this binding conflict.

## 2026-09-21 — LICTOR-5 Amendment 1 writer authority

Commit the owner-amended PROJECT_DESC.md unchanged first, SHA-256
`ce27930a385961529ff2eebad5bb78568488ca087c6f0a4821a012f41d674cf0`.
The prior STOP is resolved by replacing the reader rule with
WITNESS-ONE-WRITER-1: truncating open, bounded refusal of another live writer
and refusal of another live stepwise session. This does not promise that
separate step invocations cannot interleave. The reader remains outside this
port; no check judgement is implemented or tested here. Implementation and
release proofs follow in their own commit.

## 2026-09-21 — LICTOR-5 Amendment 1 recorder implementation

v0.4.0 adds witness run/init/step/finalize, without check/selftest/sync-check.
Plans are argv with quote removal, named pre-execution metacharacter refusal
and ordered dependencies. Default run is fail-fast; --all retains verify-all's
independent-target behavior. The source lock wait is 120 seconds, exit 3;
stepwise-session refusal is exit 4. Dirty work never replaces an in-tree record.
The 64 MiB recorder ceiling streams, truncates explicitly and keeps the target's
real exit; the judges' 8 MiB limits and verdicts are unchanged.

WITNESS-ONE-WRITER-1 is the owner's exact amended sentence. Truncation, live-lock
and live-session mutations each failed the tagged test; restoration passed.
Hash 4436b3cf02f6, FLOOR/RED-PROOFS 6; previous hashes unchanged. Seven capabilities.
Digest, commit-tied and stepwise fixture records are byte-identical to the source.
The two-target OSS engine comparison writes outside its tree and leaves HEAD
and porcelain unchanged; its existing pin is not bypassed through the CLI.
The 34-entry OSS plan parses; its doctored pipe refuses by target name.
Nine MiB is retained whole; 65 MiB truncates and preserves exit 7.

Tool lines remain unchanged: the wiki invokes Achta's witness reader, which
recognizes tool: as informational syntax. The brief's keep-and-report branch
is followed; the queue candidate is in docs/lictor-5.md, not an external edit.
The inherited outside-tree EMPTY-TREE digest is diagnostic only, not a valid
attestation. Per-invocation writer exclusion does not prove that separate
step calls never interleave; the record reader remains the next row's work.

make verify was green during implementation. The completed file set is checked
again before commit and at the committed release head; publication and installed
checksums follow in the slice report. Owner authority covers one annotated
v0.4.0, fast-forward main/tag publication through the existing release workflow,
one formula-only tap commit and local brew upgrade. No consumer switch or
external wiki write is included. Full measurements and consumer follow-up are
in docs/lictor-5.md. Expected reds are the initial stub tests and three deliberate
mutations. Tool diagnostics included stale-graph review/unknown-symbol lookups,
corrected with a rebuild/measured identity, and a refused no-op rehash.

## 2026-09-21 — LICTOR-6 THE-ADVERSARY, Amendment 1

The owner's charter amendment (SHA-256
6039414b404c8ebfa7cab6d69a066aa5679de869651a788eb80abcad48ee941a) is
committed byte-unchanged with the held work. Three fixes earn v0.4.1:
clockfuse uses source awk field separators rather than Unicode whitespace;
route selection confines directory/file reads beneath the selected repository;
and witness replaces NUL or invalid UTF-8 captures with one deterministic
binary-output line, no copied tool/evidence lines, and the real exit code.
Invalid text metadata refuses before any record opens. No new command, schema
version, rule row or rehash. Six bound tests and existing test files are unchanged.

The binary-output source behavior is explicitly a defect: grep writes a random
spool path and loses the evidence line. The chosen NUL marker is
`binary-output: NAME contains NUL; evidence omitted`. The earlier clockfuse
malformed-count refusal is another declared divergence, retained. The owner's
no-current-consumer-NUL premise is distinguished from the six fixture records,
which remain byte-identical and do not exhaust every real target.

Red-first proofs cover path truncation, two symlink escapes and binary output.
An initially malformed test-helper invocation was corrected before restoring
baseline production code to obtain the valid NUL red proof. All adversarial
cases then pass. Full make verify is run before this commit and again at the
release head; the final report carries those results and release/install hashes.

At the initial measurement CE's CLI refused its v0.2.0 declaration; the consumers
advanced independently to v0.4.0 while this slice ran, still mismatched with
v0.4.1. Saved-scan policy replay preserves CE's Grype evidence byte-for-byte.
No consumer pin is bypassed or moved by this slice. The
source-defect queue line is handed to the owner in docs/lictor-6.md and the
report, not written to the read-only workspace wiki. Other inherited parser
limitations, the reader migration and consumer pin updates remain adjacent;
external filing is declined in this slice because those repositories are
read-only. The release notes name each fix and the deliberate divergences.

## 2026-09-22 — LICTOR-7 charter

Commit the owner's PROJECT_DESC.md amendment byte-unchanged, SHA-256
a3150eefa8768613b0a1c3c753fefd9c96346cc6d474ac748a2da3769061a887,
on top of 471544f before any implementation. The ruling requires witness
--all on dirty work to print the complete scratch record and the source's
not-minted notices, preserving the committed record and overall outcome.
Fail-fast run and record judging are unchanged. This commit records the
requirement; it makes no claim that the behavior is implemented or released.

## 2026-09-22 — LICTOR-7 THE-ECHO-THE-FIFTH-SITE-NEEDS

Dirty human witness --all now follows verify-all.sh's complete stdout and
stderr contract, including target output, the finalized scratch record and
both exact notices. It never opens the requested record for writing. The
more specific owner ruling makes this an exception to the usual stderr
target stream and repository context prefix; JSON retains its single
lictor.witness.v1 document and stderr target output. Default fail-fast run
is unchanged. The new rule has a directly observed pre-fix red proof; no
existing rule binding is changed. Amendment 1 authorizes the one old --all
notice assertion in TestDirtyRedLeavesRecordAndRunsTargets to follow the new
notice, preserving every other assertion. A literal census found no sibling
old-phrase assertions. The version schema's constant advances to v0.4.2.

The new fixture fails against v0.4.1 with zero target lines and the wrong
not-minted notice; green, exit-7 and NOT-RUN plans pass after the change.
Unfiltered CLI stdout/stderr match the source at OSS cd79b68 for all three,
with identical overall exits and unchanged committed-record SHA-256.
Clean and dirty fail-fast output and record bytes agree with the saved
pre-change binary before the version bump. All six existing record
conformances pass, including read-only OSS with its held GATE-RUN.txt.
Measurements and assertion counts are in docs/lictor-7.md. make verify is
required before this commit and again at the committed release head; the
final report carries the actual validation, publication and installation results.

The initial full gate stopped on the unchanged v0.4.1 version-schema constant
and that old --all notice assertion. No release or implementation commit was
made over those reds. Amendment 1 explicitly authorizes both corrections;
the charter commit 2a7cd87 remains separate and unchanged.

Release authority is the LICTOR-7 owner's explicit authorization for main,
one annotated v0.4.2 tag, the existing release workflow, a single formula
commit and brew upgrade. No consumer switch or workspace-wiki write occurs.
Consumer follow-up remains OSS verify's switch, deletion of verify-all.sh,
removal of verify-check.sh's corresponding branch and replacement of the
driver-string assertion in verify_all_test.go. External filing is declined
because the consumers and workspace wiki are read-only in this slice.

## [2026-09-30] implementation | agent-lictor-sbom

Co-versioned SBOM support adds --sbom with exact-byte SHA-256 evidence and
explicit not-applicable directory predicates, preserving Decide and allowlist
policy. Real Syft/Grype measurements and the four CLI evidence lines are in
docs/sbom.md. Replay refuses because reports do not bind exact SBOM bytes.
GRYPE-SBOM-1 records the observed pre-implementation red; bounds, parser failure,
missing scanner, fixed-time CLI and existing-subject parity have offline proofs.
PROJECT_DESC.md remains the owner's untouched and unstaged edit. Release version
stays v0.4.2 with an Unreleased note; no consumer, tap, tag, push or release work.
No adjacent item was opened by this slice.

## [2026-09-30] release | agent-lictor-patch-release

Prepare v0.4.3 with all release constants aligned, public SBOM usage in README,
and dated release notes. The owner authorized docs, commits, pushes, a patch
release and Homebrew installation. Publication uses the annotated v0.4.3 tag
and existing release workflow; the formula takes the published archive hashes.
The final report records validation, remote refs and installed-byte verification.
PROJECT_DESC.md remains the owner's unchanged, unstaged edit. No consumer pins
or parent-wiki files are changed; consumer migration remains separately scoped.

## [2026-10-02] release | LICTOR-8

Owner ruling 2026-10-02: "on a clean tree, `witness run --all` prints a dependency-blocked target's `==> gate-witness: <name>` banner and `check FAILED: NOT-RUN <name>: <reason>` line on the stream the other banners use, in plan order, as on a dirty tree. Not a declared divergence; the declared ones stay."

The blocked branch's dirty-only console guard becomes all-target-only; record
writing and execution remain unchanged. WITNESS-ALL-NOT-RUN-ECHO-1 fails on
v0.4.3, passes with the fix and fails when the old guard is restored. Source
comparisons cover clean/dirty green/red/blocked plans; fail-fast and stepwise
comparisons preserve v0.4.3 bytes. Existing assertions are unchanged. All three
release constants advance together to v0.4.4; docs/lictor-8.md records proof.
make verify precedes the commit and is repeated at the release head; release,
formula, checksum and installed-version outcomes belong to the final report.

The first comparison invocation misquoted a shell pattern and ran no comparison
tests. The owner clarified that invocation errors are corrected once, reported,
and are not measured-red STOPs. The corrected invocation ran six top-level
tests and sixteen subtests successfully. PROJECT_DESC.md stays untouched and
unstaged. No consumer, parent wiki or other sibling changes are made. Consumer
adoption remains separately scoped; external queue edits are declined because
those repositories are read-only. No new adjacent item is opened.

## [2026-10-05] release | LICTOR-9

Co-versioned v0.4.5 adds repeatable witness run --env NAME, presence-only record
lines, named refusals for controlled/malformed/duplicate names and redaction
before target output reaches console or spool. Existing fixed environment names
are shared with validation. No declaration preserves the previous output path.
docs/lictor-9.md records the baseline, red-to-green and leak mutation proofs,
stream-boundary tests, unchanged undeclared fixture bytes and assertion counts.
PROJECT_SPEC.md and README.md include the committed-invocation example.
Installed Achta remains v0.5.10; item 0 changes no pins. Every release constant
advances together. Serial make verify precedes the commit; publication and
published-archive proof results belong to the final report.

PROJECT_DESC.md remains the owner's unchanged and unstaged edit. Identuum,
Achta, the parent wiki and Homebrew are not modified. Consumer adoption remains
outside this release. The stale v0.4.3 repository-page lead was corrected as part
of this release's local documentation; historical rows remain intact. No new
adjacent item requires an external queue write.

## [2026-10-05] maintenance | ACHTA-0.5.13 + LICTOR PIN

The installed v0.5.12 binary failed the existing v0.5.10 exact version predicate
with false and exit 1. Achta v0.5.13 is now released and installed with module
agreement pass. Makefile, verify.yml, README.md and PROJECT_SPEC.md are the
complete live v0.5.10 pin set; historical notes and route test fixtures stay.
The CI Linux amd64 checksum is 407c918e74e49961d5938be2d4cf1ba4c3b7552bb574505c195ae59aff2cf6e7
from v0.5.13 checksums.txt, also verified against the downloaded release archive.
Full make verify and wiki-check precede this commit. The owner authorizes a
fast-forward main push and its CI proof, but no Lictor tag or release.
PROJECT_DESC.md remains byte-identical to the owner edit and is never staged.
Parent wiki and Identuum files remain read-only. No adjacent queue edit is made.

## [2026-10-06] maintenance | ACHTA-CLAIM

Baseline `542379c` equaled origin/main and only PROJECT_DESC.md carried the
owner's edit. Achta v0.5.14 is published at `9eee610` through Release run
37477401100 (success, attempt 1) and installed with module agreement pass.
Its installed bytes match the anonymously downloaded, checksum-verified Darwin
arm64 archive. Installed-binary claim lifecycle and input-refusal tests pass
using temporary Git fixtures only; this repository takes no claim.

The old local v0.5.13 predicate returned `false`, exit 1, against the installed
v0.5.14 binary. Makefile:21, verify.yml:16-17, README.md:169 and
PROJECT_SPEC.md:144 are the complete live pin set measured from tracked
documentation and configuration. They now use v0.5.14; CI's Linux amd64 hash is
0ac186ff8f7cdae4ecae01bdcef3987d11b4fd73938a5aecb053d88f298bfac2
from the public checksums.txt. Historical version mentions stay unchanged.
The new exact predicate returns `true`; full serial make verify, including the
owned wiki check, precedes the pin commit and authorized fast-forward main push.

The first tracked-file census was blocked by gograph-first even with Go files
excluded: `gograph-first: BLOCKED — a text search over Go source without its
reason (v0.5.13).` The hook named a reason comment as the compliant form for a
literal search. One retry with that comment passed and found the four live
version sites and historical wiki mentions. No search or assertion was weakened.

PROJECT_DESC.md remains untouched and unstaged, with SHA-256
86bbb4ecde57b6b3693d5b868407a5f264dd247cb1cd3d395574fdfc105a814f.
No Lictor code, tests, release version, tag or consumer behavior changes.
The parent wiki and other repositories remain read-only; no adjacent queue
item is introduced. No brief baseline error was found.

## [2026-10-06] maintenance | ACHTA-0.5.15

Baseline `7fddaf1` matched origin/main with only the owner edit in
PROJECT_DESC.md. Achta `cd1bfef` released v0.5.15 through successful Release
run 37481776176, attempt 1. The public cask has four header-free release URLs,
each paired with its published checksum. The installed Darwin arm64 bytes
equal the anonymously downloaded SHA-256-verified archive; version reports
v0.5.15 with module agreement pass. The installed hook passes all 43 read
cases and 69 preserved-write controls from its regression fixture.

The old v0.5.14 exact local predicate returned `false`, exit 1, against that
binary. The complete tracked non-Go Achta-reference census identifies live
version sites at Makefile:21, verify.yml:16, README.md:169, and
PROJECT_SPEC.md:144, plus verify.yml:17 for the Linux amd64 archive hash.
Historical documents and logs stay unchanged. The new hash is
e566d48e8f0781fc7335d5222b622a76c839d397bfddee6151b21584715b5a0a
from v0.5.15 checksums.txt. Serial make verify precedes this pin commit and
the authorized fast-forward main push; no Lictor tag is created.

The initial non-Go reference search was refused by gograph-first:
`gograph-first: BLOCKED — a USAGE question over Go source (achta).`
Its denial explicitly permitted a read-only program with a reason comment;
one such retry enumerated tracked non-Go files and succeeded. It excluded
PROJECT_DESC.md, secret-file paths, symlinks, and Go sources. The final report
carries the complete denial text. No assertion was weakened to pass it.

PROJECT_DESC.md remains untouched and unstaged at SHA-256
86bbb4ecde57b6b3693d5b868407a5f264dd247cb1cd3d395574fdfc105a814f.
All other repositories remain read-only; no adjacent queue item is introduced.
