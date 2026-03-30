package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	Exchange     = "memoir"
	QueueName    = "image-jobs"
	DeadQueue    = "image-jobs-dead"
	DeadExchange = "memoir-dead"
)

type ImageJob struct {
	PhotoID string `json:"photo_id"`
	RawPath string `json:"raw_path"`
}

func Consume(ctx context.Context, conn *amqp.Connection, jobs chan<- ImageJob) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	// Declare dead letter exchange and queue
	if err := ch.ExchangeDeclare(DeadExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare dead exchange: %w", err)
	}
	if _, err := ch.QueueDeclare(DeadQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare dead queue: %w", err)
	}
	if err := ch.QueueBind(DeadQueue, QueueName, DeadExchange, false, nil); err != nil {
		return fmt.Errorf("failed to bind dead queue: %w", err)
	}

	// Declare main exchange and queue
	if err := ch.ExchangeDeclare(Exchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}
	args := amqp.Table{
		"x-dead-letter-exchange":    DeadExchange,
		"x-dead-letter-routing-key": QueueName,
	}
	if _, err := ch.QueueDeclare(QueueName, true, false, false, false, args); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	if err := ch.QueueBind(QueueName, QueueName, Exchange, false, nil); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	if err := ch.Qos(4, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := ch.Consume(QueueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Println("Consumer started, waiting for image jobs...")

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed")
			}
			var job ImageJob
			if err := json.Unmarshal(msg.Body, &job); err != nil {
				log.Printf("failed to unmarshal job: %v", err)
				msg.Nack(false, false) // discard malformed message
				continue
			}
			jobs <- job
			msg.Ack(false)
		}
	}
}
