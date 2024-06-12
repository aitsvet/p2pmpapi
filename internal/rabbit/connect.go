package rabbit

import (
	"fmt"
	"log"
	"os"

	"github.com/aitsvet/saga_example/internal/model"
	"github.com/streadway/amqp"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func Connect(queues ...string) (c *Client, err error) {
	c = &Client{}

	c.conn, err = amqp.Dial(os.Getenv("RABBITMQ_HOST"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a RabbitMQ channel: %w", err)
	}

	for _, q := range queues {
		_, err := c.channel.QueueDeclare(
			q,     // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)

		if err != nil {
			return nil, fmt.Errorf("failed to declare RabbitMQ queue: %w", err)
		}
	}

	return c, nil
}

func (c *Client) Publish(m *model.Message) error {
	err := c.channel.Publish(
		"",      // exchange
		m.Queue, // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			Headers:     amqp.Table{"txId": m.TxId},
			ContentType: "application/json",
			Body:        m.Body,
		})
	if err != nil {
		return fmt.Errorf("failed to send tx %d to RabbitMQ %s: %w", m.TxId, m.Queue, err)
	}
	log.Printf(" [x] Sent tx %d to RabbitMQ %s\n", m.TxId, m.Queue)
	return nil
}

func (c *Client) ListenTo(queue string) <-chan *model.Message {
	output := make(chan *model.Message)
	go func() {
		for {
			msgs, err := c.channel.Consume(
				queue, // queue
				"",    // consumer
				true,  // auto-ack
				false, // exclusive
				false, // no-local
				false, // no-wait
				nil,   // args
			)
			if err != nil {
				fmt.Printf("failed to register consumer for RabbitMQ %s: %v\n", queue, err)
				break
			}
			for d := range msgs {
				m := &model.Message{
					Queue: queue,
					TxId:  d.Headers["txId"].(uint64),
					Body:  d.Body,
				}
				log.Printf(" [x] Received tx %d from RabbitMQ %s\n", m.TxId, m.Queue)
				output <- m
			}
		}
		close(output)
	}()
	return output
}

func (c *Client) Close() error {
	err := c.channel.Close()
	if err != nil {
		return fmt.Errorf("failed to close RabbitMQ channel: %w", err)
	}
	err = c.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close RabbitMQ connection: %w", err)
	}
	return nil
}
