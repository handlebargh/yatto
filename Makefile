## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## lint: run golangci-lint linter package
.PHONY: lint
lint:
	@echo 'Linting .go files...'
	golangci-lint run

## fmt: run golangci-lint formatter package
.PHONY: fmt
fmt:
	@echo 'Formatting .go files...'
	golangci-lint fmt

## update: update Go dependencies
.PHONY: update
update:
	@echo 'Updating module dependencies...'
	go get -u
	@echo 'Tidying module dependencies...'
	go mod tidy -v

## tidy: tidy and verify module dependencies
.PHONY: tidy
tidy:
	@echo 'Tidying module dependencies...'
	go mod tidy -diff
	@echo 'Verifying module dependencies...'
	go mod verify

## test: run application tests
.PHONY: test
test:
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
