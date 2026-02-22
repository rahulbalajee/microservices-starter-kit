package messaging

import (
	"context"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"

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

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message string) error {
	log.Printf("publishing message with routing key: %s", routingKey)
	return r.Channel.PublishWithContext(
		ctx,
		TripExchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         []byte(message),
			DeliveryMode: amqp.Persistent,
		},
	)
}

type MessageHandler func(context.Context, amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	err := r.Channel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %v", err)
	}

	msgs, err := r.Channel.Consume(
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

	ctx := context.Background()

	go func() {
		for msg := range msgs {
			log.Printf("received a message: %s\n", msg.Body)

			if err := handler(ctx, msg); err != nil {
				log.Printf("failed to handle the message: %v\n", err)

				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("error: failed to nack message: %v", nackErr)
				}

				continue
			}

			if ackErr := msg.Ack(false); ackErr != nil {
				log.Printf("error: failed to ack message: %v. Message body: %s\n", ackErr, msg.Body)
			}
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
