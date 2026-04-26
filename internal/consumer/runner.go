package consumer

import (
	"context"
	"errors"
	"time"

	"github.com/cathudson/order-service/internal/protoxjson"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	baseDelay = time.Second
	maxDelay  = 30 * time.Second
)

type MessageHandler[M proto.Message] interface {
	Handle(ctx context.Context, msg M) error
}

type Runner[M proto.Message] struct {
	reader     *kafka.Reader
	handler    MessageHandler[M]
	newMessage func() M
	logger     *zap.SugaredLogger
	dlqWriter  *kafka.Writer
}

func NewRunner[M proto.Message](reader *kafka.Reader, dlqWriter *kafka.Writer, handler MessageHandler[M], newMessage func() M, logger *zap.SugaredLogger) *Runner[M] {
	return &Runner[M]{
		reader:     reader,
		dlqWriter:  dlqWriter,
		handler:    handler,
		newMessage: newMessage,
		logger:     logger,
	}
}

func (r *Runner[M]) Run(ctx context.Context) {
	for {
		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			r.logger.Errorw("runner failed to fetch message", "error", err)
			return
		}

		entity := r.newMessage()
		err = protoxjson.Unmarshal(msg.Value, entity)
		if err != nil {
			r.logger.Errorw("runner failed to unmarshal message, skip", "error", err)
			cErr := r.reader.CommitMessages(ctx, msg)
			if cErr != nil {
				r.logger.Errorw("runner failed to commit message", "error", cErr)
			}
			dErr := r.dlqWriter.WriteMessages(ctx, kafka.Message{Key: msg.Key, Value: msg.Value})
			if dErr != nil {
				r.logger.Errorw("runner failed to write message to DLQ", "error", dErr)
			}
			continue
		}

		r.handleWithRetry(ctx, entity, msg.Key)

		if ctx.Err() != nil {
			return
		}

		err = r.reader.CommitMessages(ctx, msg)
		if err != nil {
			r.logger.Errorw("runner failed to commit message", "error", err)
		}
	}
}

func (r *Runner[M]) handleWithRetry(ctx context.Context, entity M, key []byte) {
	attempt := 0
	for {
		err := r.handler.Handle(ctx, entity)
		if err == nil {
			break
		}
		if errors.Is(err, errSkipMessage) {
			r.logger.Infow("skip message handler")
			value, mErr := protoxjson.Marshal(entity)
			if mErr != nil {
				r.logger.Errorw("runner failed to marshal message, skip", "error", mErr)
				break
			}
			dErr := r.dlqWriter.WriteMessages(ctx, kafka.Message{Key: key, Value: value})
			if dErr != nil {
				r.logger.Errorw("runner failed to write message to DLQ", "error", dErr)
			}
			break
		}
		r.logger.Errorw("runner failed to handle message, retrying", "error", err, "attempt", attempt)

		delay := baseDelay << attempt
		if delay > maxDelay {
			delay = maxDelay
		}
		attempt++

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}
