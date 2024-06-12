package kafka

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/aitsvet/saga_example/internal/model"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Client struct {
	cm *kafka.ConfigMap
	p  *kafka.Producer
}

func (c *Client) Close() {
	if c.p != nil {
		c.p.Close()
	}
}

func Connect(queue ...string) (*Client, error) {
	c := &Client{cm: &kafka.ConfigMap{"bootstrap.servers": os.Getenv("KAFKA_BROKERS")}}
	a, err := kafka.NewAdminClient(c.cm)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin client: %w", err)
	}
	defer a.Close()
	ctx := context.Background()
	topicSpecs := make([]kafka.TopicSpecification, len(queue))
	for i, q := range queue {
		topicSpecs[i] = kafka.TopicSpecification{
			Topic:             q,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
	}
	results, err := a.CreateTopics(ctx, topicSpecs)
	if err != nil {
		return nil, fmt.Errorf("failed to create topics: %w", err)
	}
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			return nil, fmt.Errorf("failed to create topic: %w", result.Error)
		}
	}
	c.p, err = kafka.NewProducer(c.cm)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) Publish(m *model.Message) error {
	var txId []byte
	binary.LittleEndian.PutUint64(txId, m.TxId)
	err := c.p.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &m.Queue, Partition: kafka.PartitionAny},
			Headers:        []kafka.Header{{Key: "txId", Value: []byte(txId)}},
			Value:          m.Body,
		}, nil)
	if err != nil {
		return fmt.Errorf("failed to send tx %d to Kafka %s: %w", m.TxId, m.Queue, err)
	}
	log.Printf(" [x] Sent tx %d to Kafka %s\n", m.TxId, m.Queue)
	return nil
}

func (c *Client) ListenTo(queue string) <-chan *model.Message {
	output := make(chan *model.Message)
	go func() {
		defer close(output)
		l, err := kafka.NewConsumer(c.cm)
		if err != nil {
			fmt.Printf("failed to create consumer for %s: %v\n", queue, err)
			return
		}
		err = l.Subscribe(queue, nil)
		if err != nil {
			fmt.Printf("failed to create consumer for %s: %v\n", queue, err)
			return
		}
		for {
			ev := l.Poll(100)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				m := &model.Message{
					Queue: queue,
					Body:  e.Value,
				}
				for _, h := range e.Headers {
					if h.Key == "txId" {
						m.TxId = binary.LittleEndian.Uint64(h.Value)
					}
				}
				log.Printf(" [x] Received tx %d from Kafka %s\n", m.TxId, m.Queue)
				output <- m
			case kafka.Error:
				fmt.Printf("error while consuming message: %v\n", e)
			default:
				log.Printf("ignored event: %s\n", e)
			}
		}
	}()
	return output
}
