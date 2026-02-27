package messaging

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/contracts"
)

type QueueConsumer struct {
	mq        *RabbitMQ
	connMgr   *ConnectionManager
	queueName string
}

func NewQueueConsumer(mq *RabbitMQ, connMgr *ConnectionManager, queueName string) *QueueConsumer {
	return &QueueConsumer{
		mq:        mq,
		connMgr:   connMgr,
		queueName: queueName,
	}
}

func (qc *QueueConsumer) Start(ctx context.Context) error {
	msgs, err := qc.mq.Channel.ConsumeWithContext(
		ctx,
		qc.queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				var msgBody contracts.AmqpMessage
				if err := json.Unmarshal(msg.Body, &msgBody); err != nil {
					log.Println("failed to unmarshal message:", err)
					continue
				}

				userID := msgBody.OwnerID

				var payload any
				if msgBody.Data != nil {
					if err := json.Unmarshal(msgBody.Data, &payload); err != nil {
						log.Println("failed to unmarshall payload:", err)
						continue
					}
				}

				clientMsg := contracts.WSMessage{
					Type: msg.RoutingKey,
					Data: payload,
				}

				if err := qc.connMgr.SendMessage(userID, clientMsg); err != nil {
					log.Printf("failed to send message to user %s: %v", userID, err)
				}
			}
		}
	}()

	return nil
}
