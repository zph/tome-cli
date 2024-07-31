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

test:
	go test ./...

run: build
	$(BINARY) $(ARGS)
