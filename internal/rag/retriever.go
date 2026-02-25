package rag

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type Source struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"doc_id"`
	Score      float64 `json:"score"`
	Content    string  `json:"content"`
}

type Retriever struct {
	pool *pgxpool.Pool
}

func NewRetriever(pool *pgxpool.Pool) *Retriever { return &Retriever{pool: pool} }

func (r *Retriever) TopK(ctx context.Context, q pgvector.Vector, k int) ([]Source, error) {
	if k <= 0 {
		k = 5
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id::text, document_id::text, content,
		       1 - (embedding <=> $1) AS score
		FROM chunks
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $2
	`, q, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Source
	for rows.Next() {
		var s Source
		if err := rows.Scan(&s.ChunkID, &s.DocumentID, &s.Content, &s.Score); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
