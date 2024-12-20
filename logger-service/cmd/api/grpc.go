package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"log-service/logs"
	"net"

	"google.golang.org/grpc"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (logServer *LogServer) WriteLog(ctx context.Context, request *logs.LogRequest) (*logs.LogResponse, error) {
	input := request.GetLogEntry()

	// Write the log
	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := logServer.Models.LogEntry.Insert(logEntry)

	if err != nil {
		response := &logs.LogResponse{
			Result: "Failed",
		}

		return response, err
	}

	// Return response
	response := &logs.LogResponse{
		Result: "Logged!",
	}

	return response, nil
}

func (app *Config) gRpcListen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))

	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	server := grpc.NewServer()

	logs.RegisterLogServiceServer(server, &LogServer{Models: app.Models})
	log.Printf("gRPC server started on port %s", rpcPort)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}
}
