package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoUrl = "mongodb://mongo:27017"
	gRpcPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// Connect to mongo DB
	mongoClient, err := connectToMongoDB()

	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	// Create a context in order to disconnet
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	// Close connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	// Register RPC server
	err = rpc.Register(new(RpcServer))

	go app.rpcListen()
	go app.gRpcListen()

	// Start web server
	log.Println("Starting service on port", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()

	if err != nil {
		log.Panic()
	}
}

func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on port ", rpcPort)

	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))

	if err != nil {
		return err
	}

	defer listen.Close()

	for {
		rpcConnection, err := listen.Accept()

		if err != nil {
			continue
		}

		go rpc.ServeConn(rpcConnection)
	}
}

func connectToMongoDB() (*mongo.Client, error) {
	// Create connection options
	clientOptions := options.Client().ApplyURI(mongoUrl)

	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// Connect
	connection, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Println("Error connecting:", err)

		return nil, err
	}

	log.Println("Connected to mongo")

	return connection, nil
}
