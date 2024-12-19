package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func declareExchange(channel *amqp.Channel) error {
	return channel.ExchangeDeclare(
		"logs_topic", // Name
		"topic",      // Type
		true,         // Durable?
		false,        // Auto-deleted?
		false,        // Internal?
		false,        // No-wait?
		nil,          // Arguments?
	)
}

func declareRandomQueue(channel *amqp.Channel) (amqp.Queue, error) {
	return channel.QueueDeclare(
		"",    // Name?
		false, // Durable?
		false, // Delete when unused?
		true,  // Exclusive?
		false, // No-wait?
		nil,   // Arguments?
	)
}
