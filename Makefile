.PHONY: build test lint fmt run tidy

BIN := bin/archivist

build:
	go build -o $(BIN) ./cmd/archivist

test:
	go test -race ./...

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

tidy:
	go mod tidy

run:
	go run ./cmd/archivist $(ARGS)
