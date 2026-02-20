package events

import (
	"context"
	"ride-sharing/shared/messaging"
)

type TripEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripEventPublisher(rabbitmq *messaging.RabbitMQ) *TripEventPublisher {
	return &TripEventPublisher{
		rabbitmq: rabbitmq,
	}
}

func (t *TripEventPublisher) PublishTripCreated(ctx context.Context) error {
	return t.rabbitmq.PublishMessage(ctx, "hello", "Hello, World!")
}
