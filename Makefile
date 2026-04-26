.PHONY: all build test check fmt vet lint clean help sqlc run-compute run-serve

BINARY_NAME := chronos
CMD_DIR := ./cmd/chronos
BUILD_DIR := ./bin

all: check build

help:
	@echo "Chronos - Generic pattern detection engine"
	@echo ""
	@echo "Available targets:"
	@echo "  build         Build the chronos binary"
	@echo "  test          Run all tests"
	@echo "  check         Run fmt, vet, and test"
	@echo "  fmt           Format Go code"
	@echo "  vet           Run go vet"
	@echo "  lint          Run golangci-lint"
	@echo "  clean         Remove build artifacts"
	@echo "  sqlc          Regenerate sqlc code"
	@echo "  run-compute   Run compute command (set env vars first)"
	@echo "  run-serve     Run HTTP server"

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

test:
	go test -race -count=1 ./...

check: fmt vet test

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)
	rm -f *.db

sqlc:
	sqlc generate

run-compute: build
	$(BUILD_DIR)/$(BINARY_NAME) compute --adapter=ascend --coach-id=$${CHRONOS_COACH_ID}

run-serve: build
	$(BUILD_DIR)/$(BINARY_NAME) serve
