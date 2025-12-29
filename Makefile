BINARY := deputy
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "\
	-X github.com/salmonumbrella/deputy-cli/internal/cmd.Version=$(VERSION) \
	-X github.com/salmonumbrella/deputy-cli/internal/cmd.CommitSHA=$(COMMIT) \
	-X github.com/salmonumbrella/deputy-cli/internal/cmd.BuildDate=$(DATE)"

.PHONY: build test lint fmt install clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/deputy

test:
	go test -race -cover ./...

lint:
	golangci-lint run

fmt:
	goimports -w .
	gofumpt -w .

install:
	go install $(LDFLAGS) ./cmd/deputy

clean:
	rm -rf bin/

.PHONY: deps
deps:
	go mod download
	go mod tidy
