VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: build install test test-e2e

build:
	go build -ldflags "-X main.version=$(VERSION)" -o reposcan .

install:
	go install -ldflags "-X main.version=$(VERSION)" .

test:
	go test ./...

test-e2e:
	go test -tags e2e ./...
