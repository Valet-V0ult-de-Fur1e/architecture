package rabbitmq

import (
	"context"
	"encoding/json"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	Time time.Time   `json:"time"`
}

func PublishEvent(ctx context.Context, eventType string, data interface{}) error {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	group := "БИВТ-23-СП-3"
	number := 21
	directName := group + "." + itoa(number) + ".direct"
	routingKey := group + "." + itoa(number) + ".routing.key"
	// Ensure the direct exchange exists (idempotent)
	if err := ch.ExchangeDeclare(directName, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	event := Event{
		Type: eventType,
		Data: data,
		Time: time.Now(),
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return ch.PublishWithContext(ctx, directName, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
	})
}

func ConsumeEvents(ctx context.Context, handler func(Event)) error {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	group := "БИВТ-23-СП-3"
	number := 21
	directName := group + "." + itoa(number) + ".direct"
	routingKey := group + "." + itoa(number) + ".routing.key"
	queueName := "queue." + group + "." + itoa(number)

	// Ensure exchange, queue and binding exist (idempotent). This makes the consumer robust
	// if setup at bootstrap failed due to RabbitMQ not being ready yet.
	if err := ch.ExchangeDeclare(directName, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(queueName, routingKey, directName, false, nil); err != nil {
		return err
	}

	msgs, err := ch.Consume(queueName, "app-consumer", true, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for msg := range msgs {
			var ev Event
			if err := json.Unmarshal(msg.Body, &ev); err == nil {
				handler(ev)
			}
		}
	}()
	return nil
}
