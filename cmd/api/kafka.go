package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/raffops/chat_commons/pkg/errs"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

type Topic struct {
	Name   string
	Offset int32
}

type Message struct {
	Id          string    `json:"id" validate:"required"`
	Topic       string    `json:"topic" validate:"required"`
	Offset      int64     `json:"offset" validate:"required"`
	DeliveredAt time.Time `json:"delivered_at,omitempty" validate:"required"`
	CreatedAt   time.Time `json:"created_at" validate:"required"`
	From        int64     `json:"from" validate:"required"`
	To          int64     `json:"to" validate:"required"`
	Content     string    `json:"content" validate:"required"`
}

type MessageProducer interface {
	Produce(ctx context.Context, messagesToProduce <-chan Message, errGroup *errgroup.Group) errs.ChatError
}

type MessageConsumer interface {
	Consume(ctx context.Context, messagesToConsume chan<- Message, topics <-chan []Topic, errGroup *errgroup.Group) errs.ChatError
}

type kafkaRepository struct {
	config            *kafka.ConfigMap
	ProduceRetryDelay time.Duration
}

func (k kafkaRepository) Produce(ctx context.Context, messagesToProduce <-chan Message, errGroup *errgroup.Group) errs.ChatError {
	producer, err := kafka.NewProducer(k.config)
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
							//if err != nil {
							//	logger.Logger.Error(
							//		fmt.Sprintf("Delivery failed: %s", err),
							//		zap.String("message_id", message.Id),
							//		zap.String("topic", message.Topic),
							//	)
							//}
							return err
						},
						retry.Context(ctx),
						//retry.MaxDelay(k.ProduceRetryDelay),
						retry.UntilSucceeded(),
					)
					if err != nil {
						//logger.Logger.Error(
						//	fmt.Sprintf("Delivery failed: %s", err),
						//	zap.String("message_id", message.Id),
						//	zap.String("topic", message.Topic),
						//)
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

func (k kafkaRepository) Consume(ctx context.Context, messagesToConsume chan<- Message, topics <-chan []Topic, errGroup *errgroup.Group) errs.ChatError {
	consumer, err := kafka.NewConsumer(k.config)
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

func NewKafkaPublisher(config *kafka.ConfigMap) MessageProducer {
	return &kafkaRepository{config: config}
}

func NewKafkaConsumer(config *kafka.ConfigMap) MessageConsumer {
	return &kafkaRepository{config: config}
}

func main() {
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:19092",
	}
	producer := NewKafkaPublisher(kafkaConfig)
	//consumer := NewKafkaConsumer(kafkaConfig)

	eg, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	messagesToProduce := make(chan Message, 10)

	err := producer.Produce(ctx, messagesToProduce, eg)
	failOnError(err, "producao")

	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			var id int64
			for {
				fmt.Printf("\nEnter message %d: ", id)
				reader := bufio.NewReader(os.Stdin)
				body, err := reader.ReadString('\n')
				if err != nil {
					log.Println("Error reading message")
					continue
				}
				body = strings.Trim(body, "\n")
				messagesToProduce <- Message{
					Id:        fmt.Sprintf("%d", id),
					Topic:     "test",
					Offset:    id,
					From:      10,
					To:        20,
					Content:   body,
					CreatedAt: time.Now(),
				}
				id += 1
			}
		}
	}()

	//messagesToConsume := make(chan Message)

	//topics := make(chan []Topic)
	//topics <- []Topic{
	//	{
	//		"test",
	//		0,
	//	},
	//}
	//
	//err = consumer.Consume(ctx, messagesToConsume, topics, eg)
	//failOnError(err, "consumo")

	errChan := make(chan error, 1)
	go func() {
		if err := eg.Wait(); err != nil {
			errChan <- err
		}
		return
	}()

	timer := time.NewTimer(10000 * time.Second)

	func() {
		for {
			select {
			case <-timer.C:
				log.Println("Timeout. Exiting")
				cancel()
				return
			case err := <-errChan:
				failOnError(err, "Exiting")
			case <-shutdown:
				log.Println("CTRL+C pressed. Exiting")
				cancel()
				return
				//case message := <-messagesToConsume:
				//	fmt.Printf("\nMessage recieved: %s", message.Content)
				//	timer.Reset(10 * time.Second)
			}
		}
	}()
}
