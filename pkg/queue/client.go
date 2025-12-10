package queue
import (
	"fmt"

	"github.com/handsoff/handsoff/pkg/config"
	"github.com/hibiken/asynq"
)

// Client wraps asynq.Client
type Client struct {
	*asynq.Client
}

// NewClient creates a new queue client
func NewClient(cfg config.RedisConfig) *Client {
	opt, err := asynq.ParseRedisURI(cfg.URL)
	if err != nil {
		panic(fmt.Sprintf("invalid Redis URL: %v", err))
	}
	
	client := asynq.NewClient(opt)
	return &Client{Client: client}
}
