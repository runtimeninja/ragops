.PHONY: test lint build run

test:
	go test ./... -race -count=1

lint:
	gofmt -w .
	go vet ./...

build:
	go build ./...

run:
	go run ./cmd/api