package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

//go:generate moq -rm -out gen/asynq_client_mock.go -pkg asynqgen . AsynqClient

type AsynqClient interface {
	Enqueue(ctx context.Context, typename string, task any, opts ...asynq.Option) error
	Close() error
}

type asynqClient struct {
	client *asynq.Client
}

func NewAsynqClient(redis asynq.RedisConnOpt) AsynqClient {
	return &asynqClient{client: asynq.NewClient(redis)}
}

func (c *asynqClient) Enqueue(ctx context.Context, typename string, task any, opts ...asynq.Option) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task in asynq-client: %w", err)
	}

	t := asynq.NewTask(typename, payload)
	_, err = c.client.EnqueueContext(ctx, t, opts...)
	if err != nil {
		return fmt.Errorf("enqueue task in asynq-client: %w", err)
	}

	return nil
}

func (c *asynqClient) Close() error {
	return c.client.Close()
}
