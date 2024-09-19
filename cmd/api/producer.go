package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/raffops/chat/internal/app/chat"
	"github.com/raffops/chat/pkg"
	"golang.org/x/sync/errgroup"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	writeBufferSize := 0

	writeTimeout, err := time.ParseDuration(chat.MessageWriteTimeout)
	pkg.FailOnError(err, "Invalid write timeout")

	maxDeliveryDelay, err := time.ParseDuration(chat.MaxDeliveryDelay)
	pkg.FailOnError(err, "Invalid max delivery delay")

	kafkaConfig := map[string]string{
		"bootstrap.servers": "localhost:19092",
	}

	eg, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	messagesToProduce := make(chan chat.Message, writeBufferSize)
	uuidGenerator := pkg.NewUuidGenerator()
	producer := chat.NewKafkaProducer(maxDeliveryDelay, uuidGenerator)

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
			for {
				fmt.Printf("\nEnter message: ")
				reader := bufio.NewReader(os.Stdin)
				body, err := reader.ReadString('\n')
				if err != nil {
					log.Println("Error reading message")
					continue
				}
				body = strings.Trim(body, "\n")
				chatId := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
				messagesToProduce <- chat.Message{
					ChatId:    chatId,
					From:      10,
					To:        20,
					Content:   body,
					CreatedAt: time.Now(),
				}
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
