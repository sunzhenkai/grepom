BINARY := grepom
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build clean test install lint fmt vet

build:
	go build $(LDFLAGS) -o $(BINARY) .

clean:
	rm -f $(BINARY)

test:
	go test ./...

vet:
	go vet ./...

lint: vet fmt
	@echo "lint passed"

fmt:
	gofmt -l -s .
	@test -z "$$(gofmt -l -s .)"

PREFIX ?= $(HOME)/.local

install: build
	mkdir -p $(PREFIX)/bin
	cp $(BINARY) $(PREFIX)/bin/
