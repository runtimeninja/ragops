# ragops

Production-grade RAG backend (Go + Postgres + pgvector).

This project demonstrates how to build and operate an AI-ready backend service
with proper infrastructure practices (Docker, migrations, readiness checks, CI).

---

## Current Features (Step 4)

- Go API server (`chi`)
- Docker Compose stack:
  - Postgres (pgvector)
  - Redis
  - API container
- Database migrations using `goose`
- pgvector extension enabled
- `/healthz` endpoint
- `/readyz` endpoint (DB readiness check)
- GitHub Actions CI (fmt, vet, test, build)

---

## Tech Stack

- Go 1.24
- PostgreSQL (pgvector)
- Redis
- Goose (migrations)
- Docker + Docker Compose
- GitHub Actions CI

---

## Local Development

### 1. Start Infrastructure

```bash
make docker-up

Services:

Postgres → localhost:5432

Redis → localhost:6379

API → localhost:8080

2. Run Database Migrations

Install goose (one-time):

go install github.com/pressly/goose/v3/cmd/goose@latest

Then run:

export RAGOPS_DATABASE_URL="postgres://ragops:ragops@localhost:5432/ragops?sslmode=disable"
make migrate-up
3. Test API
curl http://localhost:8080/healthz
# ok

curl http://localhost:8080/readyz
# ready
Makefile Commands
make run           # run API locally
make test          # run tests
make lint          # fmt + vet
make build         # build project
make docker-up     # start docker stack
make docker-down   # stop docker stack
make migrate-up    # apply migrations
make migrate-down  # rollback migration