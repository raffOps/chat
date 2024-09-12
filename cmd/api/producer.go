package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/raffops/chat/internal/app/chat"
	"github.com/raffops/chat/pkg"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	writeTimeout, err := time.ParseDuration(chat.MessageWriteTimeout)
	pkg.FailOnError(err, "Invalid timeout")
	kafkaConfig := map[string]string{
		"bootstrap.servers": "localhost:19092",
	}

	eg, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	messagesToProduce := make(chan chat.Message, 10)
	producer := chat.NewKafkaProducer()
	err = producer.Produce(ctx, kafkaConfig, messagesToProduce, eg)
	pkg.FailOnError(err, "producao")

	timer := time.NewTimer(writeTimeout)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
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
				messagesToProduce <- chat.Message{
					Id:        fmt.Sprintf("%d", id),
					Topic:     "test",
					Offset:    id,
					From:      10,
					To:        20,
					Content:   body,
					CreatedAt: time.Now(),
				}
				id += 1
				timer.Reset(writeTimeout)
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

	func() {
		for {
			select {
			case <-timer.C:
				log.Println("Timeout. Exiting")
				cancel()
				return
			case err := <-errChan:
				pkg.FailOnError(err, "Exiting")
			case <-shutdown:
				log.Println("CTRL+C pressed. Exiting")
				cancel()
				return
			}
		}
	}()
}
