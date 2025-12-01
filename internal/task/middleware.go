package task

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

// LoggingMiddleware logs task execution
func LoggingMiddleware(log Logger) asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			start := time.Now()
			
		log.Info("Task started",
			"type", t.Type(),
			"task_id", t.ResultWriter().TaskID())

			err := next.ProcessTask(ctx, t)
			
			duration := time.Since(start)
			
			if err != nil {
				log.Error("Task failed",
					"type", t.Type(),
					"task_id", t.ResultWriter().TaskID(),
					"duration", duration,
					"error", err)
			} else {
				log.Info("Task completed",
					"type", t.Type(),
					"task_id", t.ResultWriter().TaskID(),
					"duration", duration)
			}

			return err
		})
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(log Logger) asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Error("Task panicked",
						"type", t.Type(),
						"task_id", t.ResultWriter().TaskID(),
						"panic", r)
					err = asynq.SkipRetry // Don't retry panicked tasks
				}
			}()
			
			return next.ProcessTask(ctx, t)
		})
	}
}
