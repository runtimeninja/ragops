package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"github.com/runtimeninja/ragops/internal/rag"
)

type Processor struct {
	ing interface {
		Process(ctx context.Context, documentID string, emb rag.Embedder) error
	}
	emb rag.Embedder
}

func NewProcessor(ing interface {
	Process(ctx context.Context, documentID string, emb rag.Embedder) error
}, emb rag.Embedder) *Processor {
	return &Processor{ing: ing, emb: emb}
}

func (p *Processor) HandleIngest(ctx context.Context, t *asynq.Task) error {
	var payload IngestPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	return p.ing.Process(cctx, payload.DocumentID, p.emb)
}
