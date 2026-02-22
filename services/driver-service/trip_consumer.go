package main

import (
	"context"
	"log"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type tripConsumer struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ) *tripConsumer {
	return &tripConsumer{
		rabbitmq: rabbitmq,
	}
}

func (t *tripConsumer) Listen() error {
	return t.rabbitmq.ConsumeMessages(
		messaging.FindAvailableDriversQueue,
		func(ctx context.Context, msg amqp091.Delivery) error {
			log.Printf("driver received message: %v\n", msg)
			return nil
		},
	)
}
