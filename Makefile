BINARY=bin/tome-cli

all: build test

build: $(BINARY)

$(BINARY): $(shell find . -name '*.go')
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

test: test-go test-e2e

test/wrapper.sh: build
	@ $(BINARY) --executable wrapper.sh --root examples alias --output test/wrapper.sh

test-e2e: test/wrapper.sh build
	@ ./test/bats/bin/bats test/*.bats
	@ deno test --allow-env --allow-read --allow-run test/*.ts

test-go:
	go test ./...

run: build
	$(BINARY) $(ARGS)

tag:
	git tag v$(shell cat VERSION)
	git push origin v$(shell cat VERSION)
	chglog add --version v$(shell cat VERSION)

changelog:
	chglog add --version v$(shell cat VERSION)
