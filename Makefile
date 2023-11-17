SHELL := /bin/bash

.PHONY: lint-go
lint-go:
	cd game && golangci-lint run -v --config ../.golangci.yml

.PHONY: lint-proto
lint-proto:
	cd proto && buf lint

.PHONY: lint
lint: lint-go lint-proto

.PHONY: goimports
goimports:
	cd game && gofancyimports fix --local github.com/c4t-but-s4d/igra/game -w $(shell find . -type f -name '*.go' -not -path "./pkg/proto/*")

.PHONY: test
test:
	cd game && go test -race -timeout 1m ./...

.PHONY: validate
validate: lint test

.PHONY: proto
proto:
	cd proto && buf generate
