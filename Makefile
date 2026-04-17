.PHONY: build install test test-e2e

build:
	go build -o reposcan .

install:
	go install .

test:
	go test ./...

test-e2e:
	go test -tags e2e ./...
