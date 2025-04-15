package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

var (
	brokers = []string{"localhost:19092"}
	topic   = "game_score_logs"
)

func main() {
	consumer, err := newConsumer(brokers)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			panic(err)
		}
	}()

	partition, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := partition.Close(); err != nil {
			panic(err)
		}
	}()

	shutdown := make(chan struct{})
	go gracefulShutdown(shutdown)

ConsumerLoop:
	for {
		select {
		case msg := <-partition.Messages():
			log.Printf("consumed message: %s at %v offset\n", string(msg.Value), msg.Offset)
		case <-shutdown:
			log.Println("shutting down consumer...")
			break ConsumerLoop
		}
	}
}

func newConsumer(brokers []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.ClientID = uuid.New().String()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func gracefulShutdown(shutdown chan struct{}) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	<-ctx.Done()
	log.Println("received shutdown signal")
	close(shutdown)
}
