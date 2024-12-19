package main

import (
	"context"
	"log"
	"log-service/data"
	"time"
)

type RpcServer struct{}

type RpcPayload struct {
	Name string
	Data string
}

func (rpc *RpcServer) LogInfo(rpcPayload RpcPayload, resp *string) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      rpcPayload.Name,
		Data:      rpcPayload.Data,
		CreatedAt: time.Now(),
	})

	if err != nil {
		log.Println("Error writing to mongo", err)

		return err
	}

	*resp = "Processed payload via RPC: " + rpcPayload.Name

	return nil
}
