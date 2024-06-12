package utils

import (
	"app/config"
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type queueProductUtils struct {
	channelQueueProduct *amqp091.Channel
}

type QueueProductUtils interface {
	PushMessInQueue(data interface{}, queueName string) error
}

func (s *queueProductUtils) PushMessInQueue(data interface{}, queueName string) error {
	dataBytes, errConvert := json.Marshal(data)
	if errConvert != nil {
		return errConvert
	}

	errPush := s.channelQueueProduct.PublishWithContext(context.Background(),
		"",
		queueName,
		false, // mandatory
		false, // immediate,
		amqp091.Publishing{
			ContentType:  "text/plain",
			Body:         dataBytes,
			DeliveryMode: amqp091.Persistent,
		},
	)

	if errPush != nil {
		return errPush
	}
	return nil
}

func NewQueueProductService() QueueProductUtils {
	return &queueProductUtils{
		channelQueueProduct: config.GetRabbitChannel(),
	}
}
