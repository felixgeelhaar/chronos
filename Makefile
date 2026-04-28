.PHONY: all build test check fmt vet lint clean help sqlc \
        release-snapshot release-check docker-build \
        run-compute run-serve precommit precommit-install \
        coverctl-check coverctl-suggest nox-scan nox-baseline-update

BINARY_NAME := chronos
CMD_DIR     := ./cmd/chronos
BUILD_DIR   := ./bin

VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)

all: check build

help:
	@echo "Chronos — Time / Pattern Perception engine"
	@echo ""
	@echo "Available targets:"
	@echo "  build              Build the chronos binary into ./bin/"
	@echo "  test               Run all tests with the race detector"
	@echo "  check              Run fmt, vet, and test"
	@echo "  fmt                Format Go code"
	@echo "  vet                Run go vet"
	@echo "  lint               Run golangci-lint (uses .golangci.yml)"
	@echo "  precommit          Run all pre-commit hooks against every tracked file"
	@echo "  precommit-install  Install pre-commit and commit-msg git hooks for this clone"
	@echo "  coverctl-check     Enforce per-domain coverage policy (.coverctl.yaml)"
	@echo "  coverctl-suggest   Re-baseline coverage thresholds from current coverage"
	@echo "  nox-scan           Security scan, gates on new findings vs .nox/baseline.json"
	@echo "  nox-baseline-update Refresh the committed nox baseline after accepting findings"
	@echo "  clean              Remove build artefacts and *.db"
	@echo "  sqlc               Regenerate sqlc code into internal/store/sqlite/sqlcgen"
	@echo "  release-snapshot   Build snapshot release artefacts via goreleaser"
	@echo "  release-check      Validate .goreleaser.yaml without publishing"
	@echo "  docker-build       Build a local Docker image with version metadata"
	@echo "  run-compute        Run a compute against the configured DB"
	@echo "  run-serve          Run the HTTP server"

build:
	@mkdir -p $(BUILD_DIR)
	go build -trimpath -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

test:
	go test -race -count=1 ./...

check: fmt vet test

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

precommit:
	pre-commit run --all-files

precommit-install:
	pre-commit install
	pre-commit install --hook-type commit-msg

coverctl-check:
	coverctl check

coverctl-suggest:
	coverctl suggest --strategy current --apply --force

nox-scan:
	nox scan . -format all -output ./.nox/out

nox-baseline-update:
	nox baseline update .

clean:
	rm -rf $(BUILD_DIR) dist
	rm -f *.db *.db-journal *.db-shm *.db-wal

sqlc:
	sqlc generate

release-snapshot:
	goreleaser release --snapshot --clean --skip=publish

release-check:
	goreleaser check

docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(BINARY_NAME):$(VERSION) \
		-t $(BINARY_NAME):latest \
		.

run-compute: build
	$(BUILD_DIR)/$(BINARY_NAME) compute --adapter=ascend --scope-id=$${CHRONOS_COACH_ID}

run-serve: build
	$(BUILD_DIR)/$(BINARY_NAME) serve
