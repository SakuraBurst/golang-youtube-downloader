# Makefile for golang-youtube-downloader

.PHONY: build test lint clean all

# Binary name
BINARY_NAME=ytdl

# Build the binary
build:
	go build -o $(BINARY_NAME) ./cmd/ytdl

# Run all tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe

# Run all quality gates
all: lint test build
