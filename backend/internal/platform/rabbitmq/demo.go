package rabbitmq

import (
	"context"
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DemoRouting(ctx context.Context) error {
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
	fanoutName := group + "." + itoa(number) + ".fanout"
	directName := group + "." + itoa(number) + ".direct"
	topicName := group + "." + itoa(number) + ".topic"
	headersName := group + "." + itoa(number) + ".headers"
	queueName := "queue." + group + "." + itoa(number)
	routingKey := group + "." + itoa(number) + ".routing.key"
	topicRoutingKey := group + ".21.routing.key"

	msgs, err := ch.Consume(queueName, "demo-consumer", true, false, false, false, nil)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	ch.PublishWithContext(ctx, fanoutName, "", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("fanout message"),
	})
	ch.PublishWithContext(ctx, directName, routingKey, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("direct message"),
	})
	ch.PublishWithContext(ctx, topicName, topicRoutingKey, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("topic message"),
	})
	ch.PublishWithContext(ctx, headersName, "", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("headers message"),
		Headers:     amqp.Table{"group": group, "number": itoa(number)},
	})

	fmt.Println("[demo] Published messages to all exchanges. Waiting for delivery...")
	received := 0
	for {
		select {
		case msg := <-msgs:
			fmt.Printf("[demo] Received: %s\n", string(msg.Body))
			received++
			if received == 4 {
				return nil
			}
		case <-ctx.Done():
			fmt.Println("[demo] Timeout waiting for messages.")
			return nil
		}
	}
}
