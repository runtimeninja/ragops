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

.PHONY: docker-up docker-down

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v