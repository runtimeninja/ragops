package jobs

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

type Client struct {
	c *asynq.Client
}

func NewClient(redisAddr string) *Client {
	return &Client{c: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})}
}

func (cl *Client) Close() error { return cl.c.Close() }

type IngestPayload struct {
	DocumentID string `json:"document_id"`
}

func (cl *Client) EnqueueIngest(documentID string) (*asynq.TaskInfo, error) {
	b, _ := json.Marshal(IngestPayload{DocumentID: documentID})
	t := asynq.NewTask(TaskIngestDocument, b)
	return cl.c.Enqueue(t, asynq.Queue("default"), asynq.MaxRetry(10))
}
