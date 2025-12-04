package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"user-service/internal/domain"
)

type eventConsumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func NewEventConsumer(amqpURL, queueName string) (domain.EventConsumer, error) {
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

	return &eventConsumer{
		conn:      conn,
		channel:   ch,
		queueName: queueName,
	}, nil
}

func (c *eventConsumer) ConsumeUserCreated(ctx context.Context, handler func(userID, email string, createdAt int64) error) error {
	msgs, err := c.channel.Consume(
		c.queueName, // queue
		"",          // consumer
		false,       // auto-ack (manual ack)
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("error al registrar consumidor: %w", err)
	}

	log.Printf("Consumiendo mensajes de la cola: %s", c.queueName)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
		return fmt.Errorf("canal de mensajes cerrado")
		}

		var event map[string]interface{}
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("Error al parsear mensaje: %v", err)
				msg.Nack(false, false)
				continue
			}

		userID, ok := event["user_id"].(string)
			if !ok {
				log.Printf("Error: user_id no es string")
				msg.Nack(false, false)
				continue
			}

			email, ok := event["email"].(string)
			if !ok {
				log.Printf("Error: email no es string")
				msg.Nack(false, false)
				continue
			}

			createdAtStr, ok := event["created_at"].(string)
			if !ok {
				log.Printf("Error: created_at no es string")
				msg.Nack(false, false)
				continue
			}

		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				log.Printf("Error al parsear fecha: %v", err)
				msg.Nack(false, false)
				continue
			}

			if err := handler(userID, email, createdAt.Unix()); err != nil {
				log.Printf("Error en handler: %v", err)
				msg.Nack(false, true)
				continue
			}

		msg.Ack(false)
			log.Printf("Evento procesado: user_id=%s, email=%s", userID, email)
		}
	}
}

func (c *eventConsumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

