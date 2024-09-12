package chat

import (
	"context"
	"github.com/raffops/chat_commons/pkg/errs"
	"golang.org/x/sync/errgroup"
)

type Repository interface {
	GetChannel(queueId string) (chan map[string]interface{}, errs.ChatError)
}

type MessageProducer interface {
	Produce(ctx context.Context, config interface{}, messagesToProduce <-chan Message, errGroup *errgroup.Group) errs.ChatError
}

type MessageConsumer interface {
	Consume(ctx context.Context, config interface{}, messagesToConsume chan<- Message, topics <-chan []Topic, errGroup *errgroup.Group) errs.ChatError
}
