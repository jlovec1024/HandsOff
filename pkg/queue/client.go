package queue

import (
	"github.com/handsoff/handsoff/pkg/config"
	"github.com/hibiken/asynq"
)

// Client wraps asynq.Client
type Client struct {
	*asynq.Client
}

// NewClient creates a new queue client
func NewClient(cfg config.RedisConfig) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.URL,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &Client{Client: client}
}
