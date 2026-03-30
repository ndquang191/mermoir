package queue

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	Exchange     = "memoir"
	QueueName    = "image-jobs"
	DeadQueue    = "image-jobs-dead"
	DeadExchange = "memoir-dead"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type ImageJob struct {
	PhotoID string `json:"photo_id"`
	RawPath string `json:"raw_path"`
}

func NewPublisher(url string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare dead letter exchange and queue
	if err := ch.ExchangeDeclare(DeadExchange, "direct", true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("failed to declare dead exchange: %w", err)
	}
	if _, err := ch.QueueDeclare(DeadQueue, true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("failed to declare dead queue: %w", err)
	}
	if err := ch.QueueBind(DeadQueue, QueueName, DeadExchange, false, nil); err != nil {
		return nil, fmt.Errorf("failed to bind dead queue: %w", err)
	}

	// Declare main exchange and queue with dead letter routing
	if err := ch.ExchangeDeclare(Exchange, "direct", true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}
	args := amqp.Table{
		"x-dead-letter-exchange":    DeadExchange,
		"x-dead-letter-routing-key": QueueName,
	}
	if _, err := ch.QueueDeclare(QueueName, true, false, false, false, args); err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}
	if err := ch.QueueBind(QueueName, QueueName, Exchange, false, nil); err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &Publisher{conn: conn, channel: ch}, nil
}

func (p *Publisher) PublishImageJob(photoID, rawPath string) error {
	job := ImageJob{PhotoID: photoID, RawPath: rawPath}
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	return p.channel.Publish(Exchange, QueueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
