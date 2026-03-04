package main

import (
	"log"
	"net"
	"os"

	authdb "auth-service/internal/db"
	authserver "auth-service/internal/server"
	authv1 "ecommerce-backend-microservice/proto/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	db, err := authdb.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}

	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, authserver.NewAuthServer(db))
	reflection.Register(grpcServer) // enables grpcurl and other tools to list/call services

	log.Printf("auth-service gRPC listening on %s\n", addr)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
