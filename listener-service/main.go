package main

import (
	"fmt"
	"listener-service/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// Try to connect to RabbitMQ
	rabbitConnection, err := connect()

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer rabbitConnection.Close()

	// Start listening to messages
	log.Println("Listening for and consuming RabbitMQ messages...")

	// Create consumer
	consumer, err := event.NewConsumer(rabbitConnection)

	if err != nil {
		panic(err)
	}

	// Watch the queue and consume events
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})

	if err != nil {
		log.Println(err)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	// Don't continue until rabbit is ready
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")

		if err != nil {
			fmt.Println("RabbitMQ not yet ready...")

			counts++
		} else {
			log.Println("Connected to RabbitMQ!")

			connection = c

			break
		}

		if counts > 5 {
			fmt.Println("Backing off...")
			time.Sleep(backOff)

			continue
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second

		log.Println("Backing off...")
		time.Sleep(backOff)

		continue
	}

	return connection, nil
}
