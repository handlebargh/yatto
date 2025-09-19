PATH := $(HOME)/go/bin:$(PATH)

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## run: run the the application
.PHONY: run
run:
	go run main.go -config=examples/config.toml

## lint: run golangci-lint
.PHONY: lint
lint:
	@echo 'Linting .go files...'
	golangci-lint run ./...

## tidy: format all .go files and tidy module dependencies
.PHONY: tidy
tidy:
	@echo 'Formatting .go files...'
	gofumpt -w .
	goimports -w .
	golines -w .
	@echo 'Tidying module dependencies...'
	go mod tidy
	@echo 'Verifying module dependencies...'
	go mod verify

## audit: run quality control checks
.PHONY: audit
audit:
	@echo 'Checking module dependencies'
	go mod tidy -diff
	go mod verify
	@echo 'Linting code...'
	golangci-lint run ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## build: build the application
.PHONY: build
build:
	@echo 'Building yatto...'
	CGO_ENABLED=0 go build -v \
		-ldflags="-s -w" \
		-o=bin/yatto
	@echo 'Placing binary in ./bin directory'
