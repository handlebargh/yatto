# Print this help message
help:
    @echo 'Usage:'
    @just --list

# Run golangci-lint linter package
lint:
    @echo 'Linting .go files...'
    golangci-lint run

# Run golangci-lint formatter package
fmt:
    @echo 'Formatting .go files...'
    golangci-lint fmt

# Update Go dependencies
update:
    @echo 'Updating module dependencies...'
    go get -u ./...
    @echo 'Tidying module dependencies...'
    go mod tidy -v

# Tidy and verify module dependencies
tidy:
    @echo 'Tidying module dependencies...'
    go mod tidy -diff
    @echo 'Verifying module dependencies...'
    go mod verify

# Run application tests
test:
    @echo 'Running tests...'
    go test -v -race ./...

# Generate test coverage report as HTML
test-cover:
    #!/usr/bin/env bash
    echo 'Generating coverage report...'
    go test -coverpkg=./internal/...,./cmd/... -covermode=count -coverprofile coverage.out ./...
    go tool cover -html coverage.out -o coverage.html
    if command -v open >/dev/null 2>&1; then
        open coverage.html
    elif command -v xdg-open >/dev/null 2>&1; then
        xdg-open coverage.html
    else
        echo "No opener found; please open manually."
    fi
    echo "Coverage report available at file://$(pwd)/coverage.html"

# Build the application
build:
    @echo 'Building yatto...'
    @CGO_ENABLED=0 \
    GO111MODULE=on \
    GOFLAGS="-mod=readonly -trimpath" \
    go build -v \
        -ldflags="-s -w -extldflags=-zrelro -extldflags=-znow" \
        -o=bin/yatto
    @echo 'Placing binary in ./bin directory'
