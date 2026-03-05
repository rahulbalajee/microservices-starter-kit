package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"

	pdb "ride-sharing/shared/proto/driver"
)

type driverConsumer struct {
	RabbitMQ *messaging.RabbitMQ
	Service  domain.TripService
}

func NewDriverConsumer(rabbitmq *messaging.RabbitMQ, svc domain.TripService) *driverConsumer {
	return &driverConsumer{
		RabbitMQ: rabbitmq,
		Service:  svc,
	}
}

func (d *driverConsumer) Listen() error {
	return d.RabbitMQ.ConsumeMessages(
		context.Background(),
		messaging.DriverTripResponseQueue,
		func(ctx context.Context, msg amqp091.Delivery) error {
			var driverEvent contracts.AmqpMessage
			if err := json.Unmarshal(msg.Body, &driverEvent); err != nil {
				log.Printf("failed to unmarshal message: %v", err)
				return err
			}

			var payload messaging.DriverTripResponseData
			if err := json.Unmarshal(driverEvent.Data, &payload); err != nil {
				log.Printf("failed to unmarshal message: %v", err)
				return err
			}

			log.Printf("driver response received message: %+v\n", payload)

			switch msg.RoutingKey {
			case contracts.DriverCmdTripAccept:
				if err := d.handleTripAccepted(ctx, payload.TripID, payload.Driver); err != nil {
					log.Printf("failed to handle trip accept: %v", err)
					return err
				}
			case contracts.DriverCmdTripDecline:
				if err := d.handleTripDeclined(ctx, payload.TripID, payload.RiderID); err != nil {
					log.Printf("failed to handle trip decline: %v", err)
					return err
				}
			default:
				log.Printf("unknown trip event: %+v", payload)
			}

			return nil
		},
	)
}

func (d *driverConsumer) handleTripAccepted(ctx context.Context, tripID string, driver *pdb.Driver) error {
	trip, err := d.Service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	if trip == nil {
		return fmt.Errorf("trip not found: %s", tripID)
	}

	if err := d.Service.UpdateTrip(ctx, tripID, "accepted", driver); err != nil {
		log.Printf("failed to update the trip: %v", err)
		return err
	}

	trip, err = d.Service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	marshalledTrip, err := json.Marshal(trip)
	if err != nil {
		return err
	}

	if err := d.RabbitMQ.PublishMessage(
		ctx,
		contracts.TripEventDriverAssigned,
		contracts.AmqpMessage{
			OwnerID: trip.UserID,
			Data:    marshalledTrip,
		},
	); err != nil {
		return err
	}

	// TODO: Notify payment service to initiate payment process
	marshalledPayload, err := json.Marshal(messaging.PaymentTripResponseData{
		TripID:   tripID,
		UserID:   trip.UserID,
		DriverID: driver.Id,
		Amount:   trip.RideFare.TotalPriceInCents,
		Currency: "USD",
	})

	if err := d.RabbitMQ.PublishMessage(
		ctx,
		contracts.PaymentCmdCreateSession,
		contracts.AmqpMessage{
			OwnerID: trip.UserID,
			Data:    marshalledPayload,
		},
	); err != nil {
		return err
	}

	return nil
}

func (d *driverConsumer) handleTripDeclined(ctx context.Context, tripID, riderID string) error {
	trip, err := d.Service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	if trip == nil {
		return fmt.Errorf("trip not found: %s", tripID)
	}

	newPayload := messaging.TripEventData{
		Trip: trip.ToProto(),
	}

	marshalledPayload, err := json.Marshal(newPayload)
	if err != nil {
		return err
	}

	if err := d.RabbitMQ.PublishMessage(
		ctx,
		contracts.TripEventDriverNotInterested,
		contracts.AmqpMessage{
			OwnerID: riderID,
			Data:    marshalledPayload,
		},
	); err != nil {
		return err
	}

	return nil
}
