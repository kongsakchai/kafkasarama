package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

var (
	brokers = []string{"localhost:19092"}
	topic   = "game_score_logs"
	groupID = "log_group"
)

func main() {
	consumer, err := newConsumerGroup(brokers, groupID)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			panic(err)
		}
	}()
	handler := &handler{ready: make(chan struct{})}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx, gracefully := context.WithCancel(context.Background())
	go func() {
		defer wg.Done()

		for {
			if err := consumer.Consume(ctx, []string{topic}, handler); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					log.Printf("error consuming: %v", err)
					return
				}
			}

			if ctx.Err() != nil {
				return
			}

			log.Println("rebalancing...")
			handler.ready = make(chan struct{}) // reinitialize the ready channel for close() in setup function again
		}
	}()

	<-handler.ready

	go gracefulShutdown(gracefully)
	wg.Wait()
	log.Println("consumer group closed")
}

func newConsumerGroup(brokers []string, groupID string) (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()
	config.ClientID = uuid.New().String()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func gracefulShutdown(shutdown context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	<-ctx.Done()
	log.Println("received shutdown signal")
	shutdown()
}
