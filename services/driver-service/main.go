package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var (
	grpcAddr = env.GetString("GRPC_ADDR", ":9092")
)

func main() {
	svc := NewService()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	// rabbitmq connection
	rabbitmqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
	mq, err := messaging.NewRabbitMQ(rabbitmqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer mq.Close()
	log.Println("starting rabbitMQ connection")

	// starting the gRPC server
	grpcServer := grpcserver.NewServer()
	NewGrpcHandler(grpcServer, svc)

	consumer := NewTripConsumer(mq)
	go func() {
		if err := consumer.Listen(); err != nil {
			log.Fatalf("failed to listen to the message: %v\n", err)
		}
	}()

	log.Printf("starting gRPC server driver service on port %s\n", lis.Addr())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve: %v\n", err)
			cancel()
		}
	}()

	// wait for shutdown signal
	<-ctx.Done()
	log.Println("shutting down the server")
	grpcServer.GracefulStop()
}
