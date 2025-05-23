package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort string = "8080"

type Config struct {
	RabbitMq *amqp.Connection
}

func main() {
	// Try to connect to RabbitMQ
	rabbitConnection, err := connect()

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer rabbitConnection.Close()

	app := Config{
		RabbitMq: rabbitConnection,
	}

	log.Printf("Starting broker service on port %s\n", webPort)

	// Define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// Start the server
	err = srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
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
