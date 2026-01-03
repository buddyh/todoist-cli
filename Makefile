.PHONY: build clean install test

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/todoist ./cmd/todoist

install:
	go install $(LDFLAGS) ./cmd/todoist

clean:
	rm -rf bin/

test:
	go test -v ./...

# Quick run for development
run:
	go run ./cmd/todoist $(ARGS)

# Fetch dependencies
deps:
	go mod download
	go mod tidy

# Build for all platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/todoist-darwin-amd64 ./cmd/todoist
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/todoist-darwin-arm64 ./cmd/todoist
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/todoist-linux-amd64 ./cmd/todoist
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/todoist-linux-arm64 ./cmd/todoist
