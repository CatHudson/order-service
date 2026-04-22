package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cathudson/order-service/internal/proto"
	"github.com/cathudson/order-service/internal/protoxjson"
	"github.com/redis/go-redis/v9"
)

const (
	prefix         = "order-result:"
	orderResultTTL = 10 * time.Minute
)

var ErrOrderResultNotFound = errors.New("not found")

type OrderResultStore interface {
	Save(ctx context.Context, orderID string, result *proto.OrderResultEvent) error
	Get(ctx context.Context, orderID string, timeout time.Duration) (*proto.OrderResultEvent, error)
}

type orderResultStore struct {
	redis *redis.Client
}

func NewOrderResultStore(redis *redis.Client) OrderResultStore {
	return &orderResultStore{redis: redis}
}

func (s *orderResultStore) Save(ctx context.Context, orderID string, result *proto.OrderResultEvent) error {
	m, err := protoxjson.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal order result: %w", err)
	}

	pipe := s.redis.Pipeline()
	pipe.RPush(ctx, prefixedKey(orderID), m)
	pipe.Expire(ctx, prefixedKey(orderID), orderResultTTL)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("save order result: %w", err)
	}

	return nil
}

func (s *orderResultStore) Get(ctx context.Context, orderID string, timeout time.Duration) (*proto.OrderResultEvent, error) {
	entity, err := s.redis.BLMove(ctx, prefixedKey(orderID), prefixedKey(orderID), "RIGHT", "LEFT", timeout).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrOrderResultNotFound
		}
		return nil, fmt.Errorf("get order result: %w", err)
	}

	protoResult := &proto.OrderResultEvent{}
	err = protoxjson.Unmarshal([]byte(entity), protoResult)
	if err != nil {
		return nil, fmt.Errorf("unmarshal order result: %w", err)
	}

	return protoResult, nil
}

func prefixedKey(key string) string {
	return prefix + key
}
