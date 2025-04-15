package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

var (
	brokers = []string{"localhost:19092"}
	topic   = "game_score_logs"
)

func main() {
	producer, err := newProducer(brokers)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			panic(err)
		}
	}()

	for i := range 10 {
		msg := &sarama.ProducerMessage{
			Topic:     topic,
			Value:     sarama.StringEncoder(createData(i, rand.Intn(100))),
			Timestamp: time.Now(),
		}

		partition, offset, err := producer.SendMessage(msg)
		if err != nil {
			log.Println("error sending message:", err)
			continue
		}
		log.Printf("message sent to partition %d at offset %d\n", partition, offset)
	}
}

func createData(round, score int) string {
	data := struct {
		Round int `json:"round"`
		Score int `json:"score"`
	}{
		Round: round,
		Score: score,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("error marshalling data:", err)
		return ""
	}

	return string(jsonData)
}

func newProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.ClientID = uuid.New().String()
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return producer, nil
}
