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
	service  *Service
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ, service *Service) *tripConsumer {
	return &tripConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

func (t *tripConsumer) Listen() error {
	return t.rabbitmq.ConsumeMessages(
		context.Background(),
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

			switch msg.RoutingKey {
			case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
				return t.handleFindAndNotifyDrivers(
					ctx,
					payload,
				)
			}
			log.Printf("unknown trip event: %+v", payload)

			return nil
		},
	)
}

func (t *tripConsumer) handleFindAndNotifyDrivers(ctx context.Context, payload messaging.TripEventData) error {
	suitableIDs := t.service.FindAvailableDrivers(payload.Trip.SelectedFare.PackageSlug)

	if len(suitableIDs) == 0 {
		// notify the driver that no drivers are available
		if err := t.rabbitmq.PublishMessage(
			ctx,
			contracts.TripEventNoDriversFound,
			contracts.AmqpMessage{
				OwnerID: payload.Trip.UserID,
			},
		); err != nil {
			log.Printf("failed to publish message to exchange: %v", err)
			return err
		}

		return nil
	}

	log.Printf("found suitable drivers %v", len(suitableIDs))

	suitableDriverID := suitableIDs[0]

	marshalledEvent, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Notify the driver about a potential trip
	if err := t.rabbitmq.PublishMessage(
		ctx,
		contracts.DriverCmdTripRequest,
		contracts.AmqpMessage{
			OwnerID: suitableDriverID,
			Data:    marshalledEvent,
		},
	); err != nil {
		log.Printf("failed to publish message to exchange: %v", err)
		return err
	}

	return nil
}
