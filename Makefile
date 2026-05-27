NAME := nowsecure-ci
BIN  := ./bin/ns
EXE  := .

default: ci

ci: lint test dependencies-analyze

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN) $(EXE)

PACKAGES := $(shell go list ./...)

test:
	go run gotest.tools/gotestsum@latest --format testname $(PACKAGES)

LINTER_ARGS := "--fix"

lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Run 'nix develop' or install it manually." && exit 1)
	golangci-lint run ${LINTER_ARGS} ./...

dependencies-analyze:
	which govulncheck || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

.PHONY: default ci build test lint dependencies-analyze
