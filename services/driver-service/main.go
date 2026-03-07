package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var (
	grpcAddr    = env.GetString("GRPC_ADDR", ":9092")
	rabbitmqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
	environment = env.GetString("ENVIRONMENT", "development")
	endpoint    = env.GetString("OTEL_EXPORTER_ENDPOINT", "jaeger:4318")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracerCfg := tracing.Config{
		ServiceName:      "driver-service",
		Environment:      environment,
		ExporterEndpoint: endpoint,
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

	svc := NewService()

	mq, err := messaging.NewRabbitMQ(rabbitmqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer mq.Close()

	consumer := NewTripConsumer(mq, svc)
	if err := consumer.Listen(); err != nil {
		log.Fatalf("failed to listen to the message: %v\n", err)
	}

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	NewGrpcHandler(grpcServer, svc)

	log.Printf("starting gRPC server on port %s\n", lis.Addr())
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
