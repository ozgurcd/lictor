# Lictor

Identuum-specific gate execution. Initial source version: v0.1.0, unreleased.
The binding charter is [PROJECT_DESC.md](PROJECT_DESC.md); implemented contracts,
boundaries and the migration plan are in [PROJECT_SPEC.md](PROJECT_SPEC.md).

## Build and use

```sh
make build
./bin/lictor version
./bin/lictor capabilities --json
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

## Release and future Homebrew installation

No Lictor release or formula exists as part of this slice. An owner-authorized
annotated version tag triggers .github/workflows/release.yml. Version, archives,
checksums, release notes and later Formula/lictor.rb must agree. If the release
requires private GitHub access, Homebrew uses HOMEBREW_GITHUB_API_TOKEN supplied by
the operator; never commit or print its value. Consumer migrations happen only
after release, one repository per slice. No consumers were rewired for LICTOR-0.
