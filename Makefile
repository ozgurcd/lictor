.DEFAULT_GOAL := verify

export GOCACHE := $(CURDIR)/.cache/go-build
export GOMODCACHE := $(CURDIR)/.git/lictor-cache/go-mod
export TMPDIR := $(CURDIR)/.cache/tmp
export GIT_OPTIONAL_LOCKS := 0
GOVULNCHECK := $(CURDIR)/.cache/tools/govulncheck
TEST_ARGS ?=
REPO ?= $(CURDIR)

.PHONY: directories toolchain-check format-check build test race vet staticcheck govulncheck tidy-check rulefloor wiki-check verify grype-scan

directories:
	mkdir -p .cache/go-build "$(GOMODCACHE)" .cache/tmp .cache/tools bin

toolchain-check: directories $(GOVULNCHECK)
	@test "$$(go env GOVERSION)" = "go$$(awk '$$1 == "go" {print $$2}' go.mod)"
	staticcheck -version | grep -F '(0.8.1)'
	$(GOVULNCHECK) -version | grep -F 'govulncheck@v1.7.0'
	rulefloor version --json | jq -e '.version == "v0.9.1" and .version_agreement == "pass"'
	achta version --json | jq -e '.version == "v0.5.10" and .version_agreement == "pass"'

format-check:
	@test -z "$$(gofmt -l cmd internal)" || { gofmt -l cmd internal; exit 1; }

build: directories
	go build ./...
	go build -trimpath -buildvcs=true -o bin/lictor ./cmd/lictor

# Explicit live scan; the default verify chain runs offline fixtures instead.
grype-scan: build
	./bin/lictor grype --repo "$(REPO)" --unpinned

test: directories
	go test ./... -count=1 -timeout=120s $(TEST_ARGS)

# Explicit source-versus-port proof; SOURCE_SCRIPT is read-only and owner-selected.
.PHONY: clockfuse-conformance
clockfuse-conformance: directories
	go test ./internal/clockfuse -count=1 -timeout=120s -v -clockfuse-source-script "$(SOURCE_SCRIPT)"

.PHONY: witness-conformance
witness-conformance: directories
	go test ./internal/witness -count=1 -timeout=120s -v -witness-source-script "$(SOURCE_SCRIPT)" -witness-all-script "$(ALL_SCRIPT)" -witness-consumer "$(CONSUMER)"

race: directories
	go test -race ./... -count=1 -timeout=120s

vet: directories
	go vet ./...

staticcheck: directories
	staticcheck ./...

$(GOVULNCHECK):
	mkdir -p .cache/tools .cache/tmp .cache/go-build "$(GOMODCACHE)"
	GOBIN="$(CURDIR)/.cache/tools" go install golang.org/x/vuln/cmd/govulncheck@v1.7.0

govulncheck: directories $(GOVULNCHECK)
	$(GOVULNCHECK) ./...

tidy-check: directories
	go mod tidy -diff

rulefloor: directories
	rulefloor check --repo "$(CURDIR)"

wiki-check:
	@test -r wiki/WIKI.md && test -r wiki/index.md && test -r wiki/agent-rules.md && test -r wiki/repos/lictor.md && test -r wiki/log.md && test -r wiki/platform/decisions.md
	achta --wiki-dir "$(CURDIR)/wiki" wiki check --json

verify:
	$(MAKE) toolchain-check
	$(MAKE) format-check
	$(MAKE) build
	$(MAKE) test
	$(MAKE) race
	$(MAKE) vet
	$(MAKE) staticcheck
	$(MAKE) govulncheck
	$(MAKE) tidy-check
	$(MAKE) rulefloor
	$(MAKE) wiki-check
