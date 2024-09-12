package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/raffops/chat/internal/app/chat"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
)

func publish(ctx context.Context, messageChan <-chan chat.Message, queueName string, ch *amqp.Channel) error {
	for {
		message := <-messageChan
		body, err := json.Marshal(message)
		if err != nil {
			return err
		}
		err = ch.PublishWithContext(ctx,
			queueName, // exchange
			"",        // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType: "text/json",
				Body:        body,
			})
		if err != nil {
			return err
		}
	}
}

func consume(ctx context.Context, messages chan<- chat.Message, eg *errgroup.Group, queueName string, ch *amqp.Channel) error {
	q, err := ch.QueueDeclare(
		"logs", // name
		false,  // durable
		false,  // delete when unused
		true,   // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return err
	}
	err = ch.QueueBind(
		q.Name,    // queue name
		"",        // routing key
		queueName, // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}
	msgs, err := ch.ConsumeWithContext(
		ctx,
		q.Name,       // queue
		"consumer-0", // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)

	if err != nil {
		return err
	}

	eg.Go(
		func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case msg := <-msgs:
					var message chat.Message
					err := json.Unmarshal(msg.Body, &message)
					if err != nil {
						return err
					}
					messages <- message
					err = msg.Ack(true)
					if err != nil {
						return err
					}
				}
			}
		},
	)

	return nil
}

func main() {
	connPublish, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer connPublish.Close()

	chPublish, err := connPublish.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer chPublish.Close()

	connConsume, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer connConsume.Close()

	chConsume, err := connConsume.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer chConsume.Close()

	err = chPublish.ExchangeDeclare(
		"logs",              // name
		amqp.ExchangeFanout, // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)

	eg, ctx := errgroup.WithContext(context.Background())

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	messagesPublished := make(chan chat.Message)
	eg.Go(
		func() error { return publish(ctx, messagesPublished, "logs", chPublish) },
	)

	messagesConsumed := make(chan chat.Message)
	eg.Go(
		func() error { return consume(ctx, messagesConsumed, eg, "logs", chConsume) },
	)
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			for {
				fmt.Print("\nEnter message: ")
				reader := bufio.NewReader(os.Stdin)
				body, err := reader.ReadString('\n')
				if err != nil {
					log.Println("Error reading message")
					continue
				}
				messagesPublished <- chat.Message{Content: body}
			}
		}
	}()

	errChan := make(chan error, 1)
	go func() {
		if err := eg.Wait(); err != nil {
			errChan <- err
		}
		return
	}()

	timer := time.NewTimer(10 * time.Second)

	func() {
		for {
			select {
			case <-timer.C:
				log.Println("Timeout. Exiting")
				cancel()
				return
			case err := <-errChan:
				log.Fatal(err)
			case <-shutdown:
				log.Println("CTRL+C pressed. Exiting")
				cancel()
				return
			case message := <-messagesConsumed:
				fmt.Printf("\nMessage recieved: %s", message.Content)
				timer.Reset(10 * time.Second)
			}
		}
	}()

}
