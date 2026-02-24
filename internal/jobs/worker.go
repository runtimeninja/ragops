package jobs

import (
	"log/slog"

	"github.com/hibiken/asynq"
)

type Worker struct {
	log       *slog.Logger
	processor *Processor
}

func NewWorker(log *slog.Logger, processor *Processor) *Worker {
	return &Worker{log: log, processor: processor}
}

func (w *Worker) Run(redisAddr string) error {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"default": 5,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskIngestDocument, w.processor.HandleIngest)

	w.log.Info("worker started", "redis", redisAddr)
	return srv.Run(mux)
}
