package rabbitmq

import (
	"context"
	"log"
	"time"
)

// StartEventConsumer запускает обработчик событий todo.created в отдельной горутине.
// Если RabbitMQ ещё не готов, пытается перезапускать подключение с небольшим интервалом.
func StartEventConsumer(ctx context.Context) {
	for {
		err := ConsumeEvents(ctx, func(ev Event) {
			if ev.Type == "todo.created" {
				log.Printf("[event-consumer] Получено событие todo.created: %+v", ev.Data)
			}
		})
		if err != nil {
			log.Printf("[event-consumer] Ошибка запуска consumer: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}
		log.Println("[event-consumer] Consumer запущен")
		return
	}
}
