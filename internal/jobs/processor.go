package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

type Ingestor interface {
	Process(ctx context.Context, documentID string) error
}

type Processor struct {
	ing Ingestor
}

func NewProcessor(ing Ingestor) *Processor {
	return &Processor{ing: ing}
}

func (p *Processor) HandleIngest(ctx context.Context, t *asynq.Task) error {
	var payload IngestPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	return p.ing.Process(cctx, payload.DocumentID)
}
