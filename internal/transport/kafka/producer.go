package kfktransport

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"time"
)

type Producer struct {
	producer *kafka.Producer
	topic    string
}

func New(servers []string, clientID string, topic string) *Producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": servers[0],
		"client.id":         clientID,
		"acks":              "all",
	})

	if err != nil {
		panic(err)
	}

	return &Producer{
		producer: p,
		topic:    topic,
	}
}

func (p *Producer) ProduceMessage(msg []byte) error {
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Value:     msg,
		Key:       nil,
		Timestamp: time.Now(),
	}

	respch := make(chan kafka.Event)

	err := p.producer.Produce(kafkaMsg, respch)
	if err != nil {
		return err
	}

	e := <-respch
	switch ev := e.(type) {
	case *kafka.Message:
		return nil
	case kafka.Error:
		return ev
	default:
		return fmt.Errorf("unknow message type %v", ev)
	}
}

func (p *Producer) Close() {
	p.producer.Close()
}
