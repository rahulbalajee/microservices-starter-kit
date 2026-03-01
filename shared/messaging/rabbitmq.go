package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange = "trip"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		Channel: ch,
	}

	if err = rmq.setupExchangesAndQueues(); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("failed to setup exchanges and queues: %w", err)
	}

	return rmq, nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message contracts.AmqpMessage) error {
	log.Printf("publishing message with routing key: %s", routingKey)

	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return r.Channel.PublishWithContext(
		ctx,
		TripExchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         jsonMsg,
			DeliveryMode: amqp.Persistent,
		},
	)
}

type MessageHandler func(context.Context, amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(ctx context.Context, queueName string, handler MessageHandler) error {
	err := r.Channel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %v", err)
	}

	msgs, err := r.Channel.ConsumeWithContext(
		ctx,
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	go func() {
		for msg := range msgs {
			log.Printf("received a message: %s\n", msg.Body)

			msgCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

			if err := handler(msgCtx, msg); err != nil {
				log.Printf("failed to handle the message: %v\n", err)

				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("error: failed to nack message: %v", nackErr)
				}

				cancel()
				continue
			}

			if ackErr := msg.Ack(false); ackErr != nil {
				log.Printf("error: failed to ack message: %v. Message body: %s\n", ackErr, msg.Body)
			}

			cancel()
		}
	}()

	return nil
}

func (r *RabbitMQ) setupExchangesAndQueues() error {
	err := r.Channel.ExchangeDeclare(
		TripExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %s: %w", TripExchange, err)
	}

	if err = r.declareAndBindQueue(
		FindAvailableDriversQueue,
		[]string{contracts.TripEventCreated, contracts.TripEventDriverNotInterested},
		TripExchange,
	); err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	if err = r.declareAndBindQueue(
		DriverCmdTripRequestQueue,
		[]string{contracts.DriverCmdTripRequest},
		TripExchange,
	); err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	if err = r.declareAndBindQueue(
		DriverTripResponseQueue,
		[]string{contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline},
		TripExchange,
	); err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	if err = r.declareAndBindQueue(
		NotifyDriverNotFound,
		[]string{contracts.TripEventNoDriversFound},
		TripExchange,
	); err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	if err = r.declareAndBindQueue(
		NotifyDriverAssignQueue,
		[]string{contracts.TripEventDriverAssigned},
		TripExchange,
	); err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName string, messageTypes []string, exchange string) error {
	q, err := r.Channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("failed to declare queue: %w", err)
	}

	for _, msg := range messageTypes {
		if err = r.Channel.QueueBind(
			q.Name,
			msg,
			exchange,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to bind to queue: %s: %w", queueName, err)
		}
	}

	return nil
}

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}

	if r.Channel != nil {
		r.Channel.Close()
	}
}
