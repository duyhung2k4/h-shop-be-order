package config

import "github.com/rabbitmq/amqp091-go"

func connectRabbitMQ() error {
	conn, err := amqp091.Dial(urlRabbitMq)
	if err != nil {
		return err
	}

	var errCh error
	rabbitChannel, errCh = conn.Channel()
	if errCh != nil {
		return errCh
	}

	return nil
}
