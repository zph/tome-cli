BINARY=bin/tome-cli

all: build test

build: $(BINARY)

$(BINARY): $(shell find . -name '*.go')
	go build -o bin/tome-cli

clean:
	go clean
	rm -f $(BINARY)

deps:
	go mod tidy
	git submodule add https://github.com/bats-core/bats-core.git test/bats
	git submodule add https://github.com/bats-core/bats-support.git test/test_helper/bats-support
	git submodule add https://github.com/bats-core/bats-assert.git test/test_helper/bats-assert

test:
	go test ./...

run: build
	$(BINARY) $(ARGS)
