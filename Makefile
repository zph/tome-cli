BINARY=bin/tome-cli

all: build test

build: $(BINARY)
GO_FILES = $(shell find . -name '*.go' | grep -v /vendor/)
EMBED_FILES = $(shell find . -name '*.tmpl' | grep -v /vendor/)

$(BINARY): $(GO_FILES) $(EMBED_FILES)
	go build -o bin/tome-cli

clean:
	go clean
	rm -f $(BINARY)

deps-chlogs:
	go install github.com/goreleaser/chglog/cmd/chglog@latest

deps: deps-chlogs
	go mod tidy
	git submodule init
	git submodule update

test: build test-go test-e2e

test/bin/wrapper.sh: build
	@ $(BINARY) --executable wrapper.sh --root examples alias --output test/bin/wrapper.sh

test-e2e: test/bin/wrapper.sh build
	@ deno test --allow-env --allow-read --allow-run test/*.ts

test-go:
	go test ./...

run: build
	$(BINARY) $(ARGS)

tag:
	git tag v$(shell cat VERSION)
	git push origin main
	git push origin v$(shell cat VERSION)
	# This works with versions while add --version isn't
	chglog init

changelog:
	chglog init
	chglog format --template repo > CHANGELOG.md
	go run main.go docs && git add docs

docs: $(GO_FILES)
	go run main.go docs

release: clean build test
	goreleaser release --clean --verbose
