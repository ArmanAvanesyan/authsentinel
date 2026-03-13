SHELL := /usr/bin/env bash

.PHONY: help test test-go test-js test-js-coverage build \
	go-fmt go-vet go-tidy go-check lint

help:
	@echo "Available targets:"
	@echo "  test             Run all tests (Go + JS)"
	@echo "  test-go          Run Go tests"
	@echo "  test-js          Run JS SDK tests"
	@echo "  test-js-coverage Run JS SDK tests with coverage"
	@echo "  build            Build Go and JS packages"
	@echo "  go-fmt           Run gofmt on Go files"
	@echo "  go-vet           Run go vet on Go packages"
	@echo "  go-tidy          Run go mod tidy"
	@echo "  go-check         go-fmt + go-vet + test-go"
	@echo "  lint             Run JS/TS linters (workspace)"

test: test-go test-js

test-go:
	go test ./...

test-js:
	pnpm -r test

test-js-coverage:
	pnpm -r test:coverage

build:
	pnpm -r build

go-fmt:
	go fmt ./...

go-vet:
	go vet ./...

go-tidy:
	go mod tidy

go-check: go-fmt go-vet test-go

lint:
	pnpm -r lint

