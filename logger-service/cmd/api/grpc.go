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

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	//write the log
	logEntry := data.LogEntry{
		Name: input.GetName(),
		Data: input.GetData(),
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		return &logs.LogResponse{
			Result: "Error writing log",
		}, err
	}
	return &logs.LogResponse{
		Result: "Log written successfully",
	}, nil
}

func (app *Config) gRPCListen() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on port 50001: %v", err)
	}

	s := grpc.NewServer()
	logs.RegisterLogServiceServer(s, &LogServer{
		Models: app.Models,
	})
	log.Println("Starting gRPC server on port 50001")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen for gRPC : %v", err)
	}
}
