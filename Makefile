.PHONY: build test run clean fmt tidy lint install all

BINARY := gitree
BUILD_DIR := bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w \
	-X github.com/nogo/gitree/internal/version.Version=$(VERSION) \
	-X github.com/nogo/gitree/internal/version.GitCommit=$(COMMIT)

all: fmt tidy test build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./cmd/gitree

test:
	go test ./...

test-v:
	go test -v ./...

run: build
	./$(BUILD_DIR)/$(BINARY)

clean:
	rm -rf $(BUILD_DIR)
	go clean

fmt:
	go fmt ./...

tidy:
	go mod tidy

lint:
	golangci-lint run

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/gitree
