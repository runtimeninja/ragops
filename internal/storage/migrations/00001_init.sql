-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE documents (
  id             UUID PRIMARY KEY,
  source_type    TEXT NOT NULL, -- url | text | file (later)
  source_uri     TEXT NULL,
  title          TEXT NULL,
  status         TEXT NOT NULL, -- pending|processing|ready|failed
  error_message  TEXT NULL,
  content_sha256 TEXT NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX documents_content_sha256_uniq ON documents(content_sha256);

-- Raw content store (MVP). Later: S3/minio + reference.
CREATE TABLE document_blobs (
  document_id UUID PRIMARY KEY REFERENCES documents(id) ON DELETE CASCADE,
  content     TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE chunks (
  id           UUID PRIMARY KEY,
  document_id  UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  chunk_index  INT  NOT NULL,
  content      TEXT NOT NULL,
  token_count  INT  NOT NULL DEFAULT 0,
  embedding    vector(1536),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX chunks_doc_chunkindex_uniq ON chunks(document_id, chunk_index);

-- Vector index (works once you have enough rows; ok to create early)
CREATE INDEX chunks_embedding_ivfflat
ON chunks USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

CREATE TABLE requests_usage (
  id            UUID PRIMARY KEY,
  request_id    TEXT NOT NULL,
  route         TEXT NOT NULL,
  model         TEXT NULL,
  input_tokens  INT NOT NULL DEFAULT 0,
  output_tokens INT NOT NULL DEFAULT 0,
  cost_usd      NUMERIC(12,6) NOT NULL DEFAULT 0,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX requests_usage_request_id_idx ON requests_usage(request_id);

CREATE TABLE idempotency_keys (
  key           TEXT PRIMARY KEY,
  request_hash  TEXT NOT NULL,
  response_body BYTEA NOT NULL,
  status_code   INT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX idempotency_keys_expires_idx ON idempotency_keys(expires_at);

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS requests_usage;
DROP TABLE IF EXISTS chunks;
DROP TABLE IF EXISTS document_blobs;
DROP TABLE IF EXISTS documents;
DROP EXTENSION IF EXISTS vector;