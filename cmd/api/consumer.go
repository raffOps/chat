package main

import (
	"context"
	"fmt"
	"github.com/raffops/chat/internal/app/chat"
	"github.com/raffops/chat/pkg"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	kafkaConfig := map[string]string{
		"bootstrap.servers": "localhost:19092",
		"group.id":          "1",
	}

	eg, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	consumer := chat.NewKafkaConsumer()

	messagesToConsume := make(chan chat.Message, 100)

	topics := make(chan []chat.Topic, 1)
	topics <- []chat.Topic{
		{
			"test",
			0,
		},
	}

	err := consumer.Consume(ctx, kafkaConfig, messagesToConsume, topics, eg)
	pkg.FailOnError(err, "consumo")

	errChan := make(chan error, 1)
	go func() {
		if err := eg.Wait(); err != nil {
			errChan <- err
		}
		return
	}()

	func() {
		for {
			select {
			case err := <-errChan:
				cancel()
				pkg.FailOnError(err, "Exiting")
			case <-shutdown:
				log.Println("CTRL+C pressed. Exiting")
				cancel()
				return
			case message := <-messagesToConsume:
				fmt.Printf("\nMessage recieved: %s", message.Content)
			}
		}
	}()
}
