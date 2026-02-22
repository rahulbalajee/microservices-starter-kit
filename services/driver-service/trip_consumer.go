package main

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/contracts"
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
			var tripEvent contracts.AmqpMessage
			if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
				log.Printf("failed to unmarshall message: %v", err)
				return err
			}

			var payload messaging.TripEventData
			if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
				log.Printf("failed to unmarshall message: %v", err)
				return err
			}

			log.Printf("driver received message: %+v\n", payload)
			return nil
		},
	)
}
