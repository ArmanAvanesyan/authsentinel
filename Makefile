SHELL := /usr/bin/env bash

.PHONY: help test test-go test-js test-js-coverage build \
	go-fmt go-vet go-tidy go-check lint schema \
	validate-config print-schema render-config-example \
	proto-lint proto-breaking proto-generate e2e-docker

# Config/schema tooling: CONFIG_PATH=path BINARY=agent|proxy for validate-config;
# BINARY=agent|proxy and optional SCHEMA=path for print-schema;
# BINARY=agent|proxy FORMAT=json|yaml for render-config-example.
BINARY ?= agent
CONFIG_PATH ?=
SCHEMA ?=
FORMAT ?= json

help:
	@echo "Available targets:"
	@echo "  test                  Run all tests (Go + JS)"
	@echo "  test-go               Run Go tests"
	@echo "  test-js                Run JS SDK tests"
	@echo "  test-js-coverage       Run JS SDK tests with coverage"
	@echo "  build                  Build Go and JS packages"
	@echo "  go-fmt                 Run gofmt on Go files"
	@echo "  go-vet                 Run go vet on Go packages"
	@echo "  go-tidy                Run go mod tidy"
	@echo "  go-check               go-fmt + go-vet + test-go"
	@echo "  lint                   Run JS/TS linters (workspace)"
	@echo "  schema                 Generate JSON Schemas for configs (agent, proxy)"
	@echo "  validate-config        Validate config file (CONFIG_PATH=path BINARY=agent|proxy)"
	@echo "  print-schema           Print schema to stdout (BINARY=agent|proxy or SCHEMA=path)"
	@echo "  render-config-example  Render example config from defaults (BINARY=agent|proxy FORMAT=json|yaml)"
	@echo "  proto-lint              Run buf lint on protobuf APIs"
	@echo "  proto-breaking          Run buf breaking checks against main"
	@echo "  proto-generate          Generate Go and TS code from protobufs (via buf)"
	@echo "  e2e-docker              Start docker-compose (agent, proxy, redis, bff), run E2E smoke playbook, then down"

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

schema:
	go run ./cmd/schema

validate-config:
	@if [ -z "$(CONFIG_PATH)" ]; then echo "CONFIG_PATH=path BINARY=agent|proxy required"; exit 2; fi
	CONFIG_PATH="$(CONFIG_PATH)" BINARY="$(BINARY)" go run ./cmd/validateconfig

print-schema:
	@if [ -n "$(SCHEMA)" ]; then cat "$(SCHEMA)"; else cat "schemas/$(BINARY).schema.json"; fi

render-config-example:
	BINARY="$(BINARY)" FORMAT="$(FORMAT)" go run ./cmd/renderconfig

proto-lint:
	buf lint

proto-breaking:
	buf breaking --against '.git#branch=main'

proto-generate:
	buf generate

# E2E: start compose, wait for health, run test/e2e/playbook.sh, then compose down.
# Requires: docker, docker-compose, curl. Set .env from deployments/docker/.env.example.
e2e-docker:
	@cd deployments/docker && docker-compose up -d && sleep 15; \
	cd ../.. && bash test/e2e/playbook.sh; rc=$$?; \
	cd deployments/docker && docker-compose down; exit $$rc


