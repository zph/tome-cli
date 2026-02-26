BINARY=bin/tome-cli

.DEFAULT_GOAL := help

.PHONY: help build clean deps deps-chlogs test test-go test-e2e run tag changelog docs release

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

all: build test

build: $(BINARY) ## Build the binary
GO_FILES = $(shell find . -name '*.go' | grep -v /vendor/)
EMBED_FILES = $(shell find . -name '*.tmpl' | grep -v /vendor/)

$(BINARY): $(GO_FILES) $(EMBED_FILES)
	go build -o bin/tome-cli

clean: ## Remove build artifacts
	go clean
	rm -f $(BINARY)

deps-chlogs:
	go install github.com/goreleaser/chglog/cmd/chglog@latest

deps: deps-chlogs
	go mod tidy

test: build test-go test-e2e ## Run all tests (build + unit + e2e)

test/bin/wrapper.sh: build
	@ $(BINARY) --executable wrapper.sh --root examples alias --output test/bin/wrapper.sh

test-e2e: test/bin/wrapper.sh build ## Run Deno E2E tests
	@ deno test --allow-env --allow-read --allow-run test/*.ts

test-go: ## Run Go unit tests
	go test ./...

run: build ## Run the binary (pass ARGS=... for arguments)
	$(BINARY) $(ARGS)

tag:
	git tag v$(shell cat VERSION)
	git push origin main
	git push origin v$(shell cat VERSION)
	# This works with versions while add --version isn't
	chglog init

changelog: $(GO_FILES)
	chglog init
	chglog format --template repo > CHANGELOG.md
	go run main.go docs && git add docs

docs: $(GO_FILES)
	go run main.go docs

release: clean build test
	goreleaser release --clean --verbose
