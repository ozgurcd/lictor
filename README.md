# Lictor

Identuum-specific gate execution. Current release: v0.2.0.
The binding charter is [PROJECT_DESC.md](PROJECT_DESC.md); implemented contracts,
boundaries and the migration plan are in [PROJECT_SPEC.md](PROJECT_SPEC.md).

## Install

```sh
brew install ozgurcd/tap/lictor
lictor version --json
```

The repository and release downloads are public; no GitHub token is required.
The [releases page](https://github.com/ozgurcd/lictor/releases) publishes archives
for Darwin/Linux on arm64/amd64 and a `checksums.txt` file for each release.
For CI, pin `LICTOR_VERSION` and the matching platform archive's published
`LICTOR_SHA256`, download from the public release URL, verify that checksum and
assert `lictor version --json` before running a gate.

## Build and use

```sh
make build
./bin/lictor version
./bin/lictor capabilities --json
./bin/lictor green --repo /absolute/path/to/repository
./bin/lictor grype --repo /absolute/path/to/repository
./bin/lictor grype --repo /absolute/path/to/repository -scan report.json -inventory inventory.cdx.json --as-of 2026-09-19T00:00:00Z --json
```

The source judge's stdout evidence line is preserved. Human stderr separately
names the selected repository and as_of; JSON carries the same evidence as
`evaluation.reason`. Exit codes are 0 pass, 1 fail, 2 cannot-evaluate.
Without --repo, the selected repository is cwd. Saved directory reports must
name that selected tree. --as-of governs Lictor's suppression dates, never the
external scanner's clock. Replays require unchanged report, inventory, allowlist,
declaration and ignored-path inputs. Run `lictor grype --help` for all flags.

Live scans require Grype at the consumer's declared GRYPE_VERSION (source baseline
0.119.0) and Git. Lictor leaves toolchain-parity ownership with the consumer.
Executors have deadlines and output bounds; missing tools never pass.

`green` runs gofmt, build, vet and tests in that order, stopping at the first red.
The test invocation keeps `-count=1 -timeout=120s`. Human output names the subject
by base name and the Go version. Exit 0 is GREEN, 1 NOT-GREEN, 2 CANNOT-EVALUATE
(including a missing Go toolchain or go.mod). `--json` emits lictor.green.v1 with
the same cause, subject, version and exit code plus the failing output excerpt.
The excerpt is also on stderr: at most twelve lines and 16 KiB. Go test excerpts
keep only FAIL headings and test-file lines. `lictor green --help` lists flags;
omitting --repo selects cwd. No hook, bypass or selftest command is included.
See PROJECT_SPEC.md for subprocess bounds and the Go environment allowlist.

## Validation

```sh
make verify
make wiki-check
```

Go 1.27.1, staticcheck v0.8.1, rulefloor v0.9.1, Achta v0.5.10 and jq must be on PATH.
Make installs pinned govulncheck v1.7.0 under .cache/tools. Dependency download and
vulnerability-DB access belong to developer validation; unit tests are offline.
The runtime has no third-party dependencies. JSON Schema validation is test-only.
The published schemas are in schemas/; port measurements are in docs/lictor-0.md.
Build from a normal Git checkout: Make stores downloaded module sources under
.git/lictor-cache/go-mod and other build artifacts under .cache/.

## Releases

An owner-authorized annotated version tag triggers .github/workflows/release.yml;
its dispatch input can publish an existing immutable tag. Version, archives,
checksums, release notes and Formula/lictor.rb must agree. Consumer migrations
remain separate work, one repository per slice.
