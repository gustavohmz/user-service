package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"user-service/internal/domain"
)

type eventPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queueName string
}

func NewEventPublisher(amqpURL, queueName string) (domain.EventPublisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("error al conectar con RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("error al abrir canal: %w", err)
	}

	_, err = ch.QueueDeclare(
		queueName, // nombre
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("error al declarar cola: %w", err)
	}

	return &eventPublisher{
		conn:      conn,
		channel:   ch,
		queueName: queueName,
	}, nil
}

func (p *eventPublisher) PublishUserCreated(ctx context.Context, userID, email string, createdAt int64) error {
	event := map[string]interface{}{
		"user_id":    userID,
		"email":      email,
		"created_at": time.Unix(createdAt, 0).Format(time.RFC3339),
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error al serializar evento: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		"",           // exchange
		p.queueName,  // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("error al publicar mensaje: %w", err)
	}

	return nil
}

func (p *eventPublisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

