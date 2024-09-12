package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/raffops/chat_commons/pkg/errs"
	"golang.org/x/sync/errgroup"
	"log"
	"time"
)

type kafkaRepository struct{}

func (k kafkaRepository) Produce(ctx context.Context, config interface{}, messagesToProduce <-chan Message, errGroup *errgroup.Group) errs.ChatError {
	kafkaConfig, errParse := parseConfig(config)
	if errParse != nil {
		return errParse
	}
	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		return errs.NewError(
			errs.ErrInternal,
			err,
		)
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
							err := produce(ctx, message, producer, deliveryChan)
							return err
						},
						retry.Context(ctx),
						retry.UntilSucceeded(),
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

func produce(ctx context.Context, message Message, producer *kafka.Producer, deliveryChan chan kafka.Event) error {
	message.DeliveredAt = time.Now()
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}
	errProduce := producer.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &message.Topic},
			Key:            []byte(message.Id),
			Value:          jsonMessage,
		},
		deliveryChan,
	)
	if errProduce != nil {
		return errProduce
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-deliveryChan:
		if m, ok := e.(*kafka.Message); ok {
			return m.TopicPartition.Error
		}
		return fmt.Errorf("unexpected event type received: %T", e)
	}
}

func (k kafkaRepository) Consume(ctx context.Context, config interface{}, messagesToConsume chan<- Message, topics <-chan []Topic, errGroup *errgroup.Group) errs.ChatError {
	kafkaConfig, errParse := parseConfig(config)
	if errParse != nil {
		return errParse
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
						log.Println(errConsume)
					}
					var parsedMessage Message
					err = json.Unmarshal(message.Value, &parsedMessage)
					messagesToConsume <- parsedMessage
					if err != nil {
						return err
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

func NewKafkaProducer() MessageProducer {
	return &kafkaRepository{}
}

func NewKafkaConsumer() MessageConsumer {
	return &kafkaRepository{}
}
