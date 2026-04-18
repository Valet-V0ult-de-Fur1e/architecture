package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	group  = "БИВТ-23-СП-3"
	number = 21
)

func SetupExchangesAndQueue(ctx context.Context) error {
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

	fanoutName := group + "." + itoa(number) + ".fanout"
	directName := group + "." + itoa(number) + ".direct"
	topicName := group + "." + itoa(number) + ".topic"
	headersName := group + "." + itoa(number) + ".headers"
	queueName := "queue." + group + "." + itoa(number)
	routingKey := group + "." + itoa(number) + ".routing.key"
	topicRoutingKey := group + ".*.routing.key"

	if err := ch.ExchangeDeclare(fanoutName, "fanout", true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.ExchangeDeclare(directName, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.ExchangeDeclare(topicName, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.ExchangeDeclare(headersName, "headers", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(queueName, "", fanoutName, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(queueName, routingKey, directName, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(queueName, topicRoutingKey, topicName, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(queueName, "", headersName, false, amqp.Table{"x-match": "all", "group": group, "number": itoa(number)}); err != nil {
		return err
	}

	log.Println("RabbitMQ exchanges, queue, and bindings set up successfully")
	return nil
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
