.PHONY: build test run clean fmt tidy lint install all

BINARY := gitree
BUILD_DIR := bin

all: fmt tidy test build

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/gitree

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
	go install ./cmd/gitree
