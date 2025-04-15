package main

import (
	"log"

	"github.com/IBM/sarama"
)

type handler struct {
	ready chan struct{}
}

func (h *handler) Setup(_ sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
consumeLoop:
	for {
		select {
		case msg := <-claim.Messages():
			session.MarkMessage(msg, "")
			log.Printf("consumed message: %s at %v offset\n", string(msg.Value), msg.Offset)
		case <-session.Context().Done():
			log.Println("shutting down consumer group...")
			break consumeLoop
		}
	}
	log.Printf("session context done: %v\n", session.Context().Err())
	return session.Context().Err()
}
