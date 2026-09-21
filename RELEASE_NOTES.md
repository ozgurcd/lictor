# v0.2.0 — 2026-09-21

Adds `lictor green --repo ABSOLUTE_PATH [--json]`: the Identuum gofmt, build,
vet and test floor, with the inherited 120-second test timeout and first-red
behavior. Missing Go or go.mod refuses separately from a failing tree. Human
evidence names the repository base name and Go version; lictor.green.v1 publishes
the same result and bounded diagnostic excerpt. Eight source fixture cases and
GREEN-FLOOR-1 bind the port. No hook or bypass is ported; Grype policy is unchanged.
Archives remain Darwin/Linux on arm64/amd64, publicly downloadable with published
checksums. Consumer switches are separate work.

# v0.1.0 — 2026-09-19

Initial Identuum gate executor: the OSS Grype policy and tests, explicit repository
selection, saved-report replay with as_of, versioned JSON and bounded tool calls.
The source vulnerability and subject verdicts, including dated suppressions and
image NOT APPLICABLE predicates, are retained. No consumer migration is included.
Release archives target Darwin/Linux on arm64/amd64; no Windows artifacts.
