package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
	"ride-sharing/services/trip-service/internal/infrastructure/grpc"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var (
	grpcAddr    = env.GetString("GRPC_ADDR", ":9093")
	rabbitmqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracerCfg := tracing.Config{
		ServiceName:      "trip-service",
		Environment:      env.GetString("ENVIRONMENT", "development"),
		ExporterEndpoint: env.GetString("OTEL_EXPORTER_ENDPOINT", "jaeger:4318"),
	}

	sh, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("failed to initialize tracer %v", err)
	}
	defer sh(ctx)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	inmemRepo := repository.NewInmemRepository()
	svc := service.NewService(inmemRepo)

	log.Println("starting rabbitMQ connection")
	mq, err := messaging.NewRabbitMQ(rabbitmqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer mq.Close()

	publisher := events.NewTripEventPublisher(mq)

	driverConsumer := events.NewDriverConsumer(mq, svc)
	go driverConsumer.Listen()

	paymentConsumer := events.NewPaymentConsumer(mq, svc)
	go paymentConsumer.Listen()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	grpcServer := grpcserver.NewServer()
	grpc.NewGRPCHandler(grpcServer, svc, publisher)

	log.Printf("starting gRPC server on port %s\n", lis.Addr())
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve: %v\n", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Println("shutting down the server")
	grpcServer.GracefulStop()
}
