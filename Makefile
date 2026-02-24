.PHONY: test lint build run docker-up docker-down migrate-up migrate-down

test:
	go test ./... -race -count=1

lint:
	gofmt -w .
	go vet ./...

build:
	go build ./...

run:
	go run ./cmd/api

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

migrate-up:
	goose -dir internal/storage/migrations postgres "$$RAGOPS_DATABASE_URL" up

migrate-down:
	goose -dir internal/storage/migrations postgres "$$RAGOPS_DATABASE_URL" down