package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/raffops/chat_commons/pkg/errs"
	"github.com/raffops/chat_commons/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type kafkaRepository struct {
	UuidGenerator    UuidGenerator
	MaxDeliveryDelay time.Duration
}

func (k kafkaRepository) Produce(ctx context.Context, config interface{}, messagesToProduce <-chan Message, errGroup *errgroup.Group) errs.ChatError {
	kafkaConfig, errParse := parseConfig(config)
	if errParse != nil {
		return errParse
	}
	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		return errs.NewInternalError(err)
	}
	errGroup.Go(
		func() error {
			deliveryChan := make(chan kafka.Event, 1) // sync production
			for {
				select {
				case <-ctx.Done():
					producer.Close()
					return nil
				case message := <-messagesToProduce:
					err := retry.Do(
						func() error {
							err := produce(ctx, message, producer, deliveryChan, k.UuidGenerator)
							return err
						},
						retry.Context(ctx),
						retry.MaxDelay(k.MaxDeliveryDelay),
						retry.DelayType(retry.BackOffDelay),
						retry.UntilSucceeded(),
						retry.OnRetry(
							func(attempt uint, err error) {
								logger.Info(fmt.Sprintf("Error delivering message. Retry %d", attempt),
									zap.String("chat_id", message.ChatId.String()),
									zap.Int64("from", message.From),
									zap.Time("created_at", message.CreatedAt),
									zap.Error(err),
								)
							},
						),
					)
					if err != nil {
						return err
					}
				}
			}
		},
	)
	return nil
}

func produce(ctx context.Context, message Message, producer *kafka.Producer, deliveryChan chan kafka.Event, uuidGenerator UuidGenerator) errs.ChatError {
	message.DeliveredAt = time.Now()
	message.Id = uuidGenerator.GenerateUuid(message.ChatId, []byte(strconv.FormatInt(message.DeliveredAt.UnixNano(), 10)))
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return errs.NewInternalError(err)
	}
	chatId := message.ChatId.String()
	errProduce := producer.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &chatId},
			Key:            []byte(message.Id.String()),
			Value:          jsonMessage,
		},
		deliveryChan,
	)
	if errProduce != nil {
		return errs.NewInternalError(errProduce)
	}

	select {
	case <-ctx.Done():
		return nil
	case e := <-deliveryChan:
		m, ok := e.(*kafka.Message)
		if !ok {
			return errs.NewInternalError(fmt.Errorf("unexpected event type received: %T", e))
		}
		if m.TopicPartition.Error != nil {
			return errs.NewInternalError(m.TopicPartition.Error)
		}
		return nil
	}
}

func (k kafkaRepository) Consume(ctx context.Context, config interface{}, messagesToConsume chan<- Message, topics <-chan []Topic, errGroup *errgroup.Group) errs.ChatError {
	kafkaConfig, errParse := parseConfig(config)
	if errParse != nil {
		return errs.NewInternalError(errParse)
	}

	consumer, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		return errs.NewError(
			errs.ErrInternal,
			err,
		)
	}

	errGroup.Go(
		func() error {
			for {
				select {
				case topicsToConsume := <-topics:
					var kafkaTopics []kafka.TopicPartition
					for _, topic := range topicsToConsume {
						kafkaTopics = append(kafkaTopics, kafka.TopicPartition{
							Topic:     &topic.Name,
							Partition: topic.Offset,
						})
					}

					err = consumer.Assign(kafkaTopics)
					if err != nil {
						return err
					}

				case <-ctx.Done():
					err := consumer.Close()
					return err
				default:
					message, errConsume := consumer.ReadMessage(-1)
					if errConsume != nil {
						return errs.NewInternalError(err)
					}
					var parsedMessage Message
					err = json.Unmarshal(message.Value, &parsedMessage)
					messagesToConsume <- parsedMessage
					if err != nil {
						return errs.NewInternalError(err)
					}
				}
			}
		},
	)

	return nil
}

func parseConfig(config interface{}) (*kafka.ConfigMap, errs.ChatError) {
	configMap := config.(map[string]string)
	kafkaConfig := kafka.ConfigMap{}
	for k, v := range configMap {
		kv := fmt.Sprintf("%s=%s", k, v)
		err := kafkaConfig.Set(kv)
		if err != nil {
			return nil, errs.NewError(
				errs.ErrInternal,
				err,
			)
		}
	}
	return &kafkaConfig, nil
}

func NewKafkaProducer(maxDeliveryDelay time.Duration, uuidGenerator UuidGenerator) MessageProducer {
	return &kafkaRepository{MaxDeliveryDelay: maxDeliveryDelay, UuidGenerator: uuidGenerator}
}

func NewKafkaConsumer() MessageConsumer {
	return &kafkaRepository{}
}
