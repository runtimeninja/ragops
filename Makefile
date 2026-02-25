.PHONY: help up down reset logs ps \
        lint test build \
        migrate migrate-up migrate-down \
        health ready metrics \
        doc status chat \
        db-tables db-chunks db-emb \
        redis-pending redis-active redis-retry

help:
	@echo "Targets:"
	@echo "  make up            - docker compose up --build -d"
	@echo "  make down          - docker compose down"
	@echo "  make reset         - docker compose down -v (DANGER: deletes DB volume)"
	@echo "  make ps            - docker compose ps"
	@echo "  make logs          - tail api + worker logs"
	@echo ""
	@echo "  make migrate       - run migrations up (requires RAGOPS_DATABASE_URL or uses localhost default)"
	@echo "  make health        - curl /healthz"
	@echo "  make ready         - curl /readyz"
	@echo "  make metrics       - curl /metrics head"
	@echo ""
	@echo "  make doc           - create sample document (prints JSON including document_id)"
	@echo "  make status ID=...  - check document status"
	@echo "  make chat Q='...'   - ask /v1/chat (uses Q, default: What is RAG?)"
	@echo ""
	@echo "  make db-tables     - list tables"
	@echo "  make db-chunks     - count chunks"
	@echo "  make db-emb        - count chunks with embedding"
	@echo ""
	@echo "  make redis-pending - LLEN asynq pending"
	@echo "  make redis-active  - LLEN asynq active"
	@echo "  make redis-retry   - LLEN asynq retry"

# ---------- Docker ----------
up:
	docker compose up --build -d

down:
	docker compose down

reset:
	docker compose down -v

ps:
	docker compose ps

logs:
	@echo "--- api (last 100) ---"
	@docker logs -n 100 ragops-api || true
	@echo ""
	@echo "--- worker (last 100) ---"
	@docker logs -n 100 ragops-worker || true

# ---------- Go ----------
lint:
	gofmt -w .
	go vet ./...

test:
	go test ./... -count=1

build:
	go build ./...

# ---------- Migrations ----------
# If RAGOPS_DATABASE_URL isn't set, default to localhost DSN.
DB_URL ?= postgres://ragops:ragops@localhost:5432/ragops?sslmode=disable

migrate: migrate-up

migrate-up:
	RAGOPS_DATABASE_URL="$(DB_URL)" \
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir internal/storage/migrations postgres "$$RAGOPS_DATABASE_URL" up

migrate-down:
	RAGOPS_DATABASE_URL="$(DB_URL)" \
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir internal/storage/migrations postgres "$$RAGOPS_DATABASE_URL" down

# ---------- HTTP checks ----------
health:
	curl -i http://localhost:8080/healthz

ready:
	curl -i http://localhost:8080/readyz

metrics:
	curl -s http://localhost:8080/metrics | head

# ---------- API calls ----------
doc:
	curl -s -X POST http://localhost:8080/v1/documents \
	  -H "Content-Type: application/json" \
	  -d '{"title":"t1","text":"RAG means retrieval augmented generation. pgvector stores embeddings for similarity search."}' | cat

status:
	@if [ -z "$(ID)" ]; then echo "Usage: make status ID=<document_id>"; exit 1; fi
	curl -s http://localhost:8080/v1/documents/$(ID)/status | cat

chat:
	@if [ -z "$(Q)" ]; then Q="What is RAG?"; fi; \
	curl -s -X POST http://localhost:8080/v1/chat \
	  -H "Content-Type: application/json" \
	  -d "{\"question\":\"$$Q\",\"top_k\":5}" | cat

# ---------- DB checks ----------
db-tables:
	docker exec -it ragops-postgres psql -U ragops -d ragops -c "\dt"

db-chunks:
	docker exec -it ragops-postgres psql -U ragops -d ragops -c "SELECT count(*) FROM chunks;"

db-emb:
	docker exec -it ragops-postgres psql -U ragops -d ragops -c "SELECT count(*) FROM chunks WHERE embedding IS NOT NULL;"

# ---------- Redis queue checks ----------
redis-pending:
	docker exec -it ragops-redis redis-cli LLEN asynq:{default}:pending

redis-active:
	docker exec -it ragops-redis redis-cli LLEN asynq:{default}:active

redis-retry:
	docker exec -it ragops-redis redis-cli LLEN asynq:{default}:retry