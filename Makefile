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
	[[ -d test/bats ]] || git submodule add https://github.com/bats-core/bats-core.git test/bats
	[[ -d test/test_helper/bats-support ]] || git submodule add https://github.com/bats-core/bats-support.git test/test_helper/bats-support
	[[ -d test/test_helper/bats-assert ]] || git submodule add https://github.com/bats-core/bats-assert.git test/test_helper/bats-assert

test: test-go test-e2e

test/wrapper.sh: build
	@ $(BINARY) --executable wrapper.sh --root examples alias --output test/wrapper.sh

test-e2e: test/wrapper.sh build
	@ ./test/bats/bin/bats test/*.bats

test-go:
	go test ./...

run: build
	$(BINARY) $(ARGS)
