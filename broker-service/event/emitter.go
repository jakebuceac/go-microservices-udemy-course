package event

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	connection *amqp.Connection
}

func (emitter *Emitter) setup() error {
	channel, err := emitter.connection.Channel()

	if err != nil {
		return err
	}

	defer channel.Close()

	return declareExchange(channel)
}

func (emitter *Emitter) Push(event string, severity string) error {
	channel, err := emitter.connection.Channel()

	if err != nil {
		return err
	}

	defer channel.Close()

	log.Println("Pushing to channel")

	err = channel.Publish(
		"logs_topic",
		severity,
		false,
		false,
		amqp.Publishing{
			ContentType: "text",
			Body:        []byte(event),
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func NewEventEmitter(connection *amqp.Connection) (Emitter, error) {
	emitter := Emitter{
		connection: connection,
	}

	err := emitter.setup()

	if err != nil {
		return Emitter{}, err
	}

	return emitter, nil
}
