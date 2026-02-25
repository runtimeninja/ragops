package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/runtimeninja/ragops/internal/rag"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// CreateTextDocument creates a document row + stores raw content in document_blobs.
// It dedupes by content_sha256 (unique constraint).
func (s *Service) CreateTextDocument(ctx context.Context, title string, text string) (docID string, deduped bool, err error) {
	hash := sha256.Sum256([]byte(text))
	sha := hex.EncodeToString(hash[:])

	id := uuid.New()

	_, err = s.pool.Exec(ctx, `
		INSERT INTO documents (id, source_type, title, status, content_sha256)
		VALUES ($1,'text',$2,'pending',$3)
	`, id, nullable(title), sha)

	if err != nil {
		// likely unique violation by sha; fetch existing
		var existing string
		e := s.pool.QueryRow(ctx, `SELECT id::text FROM documents WHERE content_sha256=$1`, sha).Scan(&existing)
		if e == nil && existing != "" {
			return existing, true, nil
		}
		return "", false, err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO document_blobs (document_id, content)
		VALUES ($1,$2)
		ON CONFLICT (document_id) DO UPDATE SET content=EXCLUDED.content
	`, id, text)
	if err != nil {
		return "", false, err
	}

	return id.String(), false, nil
}

// Process runs ingestion for a document:
// pending -> processing -> (chunk + embed + store) -> ready
func (s *Service) Process(ctx context.Context, documentID string, emb rag.Embedder) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		UPDATE documents
		SET status='processing', updated_at=now(), error_message=NULL
		WHERE id=$1
	`, documentID)
	if err != nil {
		return err
	}

	// Load raw content
	var content string
	err = tx.QueryRow(ctx, `SELECT content FROM document_blobs WHERE document_id=$1`, documentID).Scan(&content)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, _ = tx.Exec(ctx, `
				UPDATE documents
				SET status='failed', error_message='missing document blob', updated_at=now()
				WHERE id=$1
			`, documentID)
			return err
		}
		return err
	}

	if strings.TrimSpace(content) == "" {
		_, _ = tx.Exec(ctx, `
			UPDATE documents
			SET status='failed', error_message='empty content', updated_at=now()
			WHERE id=$1
		`, documentID)
		return errors.New("empty content")
	}

	// Chunk (MVP: rune-length based)
	chunks := ChunkText(content, 800, 100)
	if len(chunks) == 0 {
		_, _ = tx.Exec(ctx, `
			UPDATE documents
			SET status='failed', error_message='no chunks produced', updated_at=now()
			WHERE id=$1
		`, documentID)
		return errors.New("no chunks produced")
	}

	// Insert chunks with embeddings
	for i, c := range chunks {
		vec, tok, err := emb.Embed(ctx, c)
		if err != nil {
			_, _ = tx.Exec(ctx, `
				UPDATE documents
				SET status='failed', error_message='embedding failed', updated_at=now()
				WHERE id=$1
			`, documentID)
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO chunks (id, document_id, chunk_index, content, token_count, embedding)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (document_id, chunk_index) DO NOTHING
		`, uuid.New(), documentID, i, c, tok, vec)
		if err != nil {
			_, _ = tx.Exec(ctx, `
				UPDATE documents
				SET status='failed', error_message='chunk insert failed', updated_at=now()
				WHERE id=$1
			`, documentID)
			return err
		}
	}

	// Mark ready
	_, err = tx.Exec(ctx, `
		UPDATE documents
		SET status='ready', updated_at=now(), error_message=NULL
		WHERE id=$1
	`, documentID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func nullable(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}
