# Print this help message
help:
    @just --list

# Run golangci-lint linter package
lint:
    golangci-lint run

# Run golangci-lint formatter package
fmt:
    golangci-lint fmt

# Tidy and verify module dependencies
tidy:
    go mod tidy -v
    go mod verify

# Run application tests
test:
    go test -v -race -count=1 ./...

# Generate test coverage report as HTML
test-cover:
    #!/usr/bin/env bash
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

# Build yatto
build:
    CGO_ENABLED=0 \
    GOFLAGS="-trimpath -mod=readonly" \
    go build -v \
        -ldflags="-s -w -buildid= -extldflags=-static-pie,-zrelro,-znow" \
        -o=bin/yatto
    @echo 'Placed binary in ./bin directory'

# Build and run yatto
run:
    go run .

# Remove build and coverage artifacts
clean:
    rm -rf bin/ coverage.out coverage.html
