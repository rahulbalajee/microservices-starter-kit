package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"ride-sharing/services/payment-service/internal/infrastructure/events"
	"ride-sharing/services/payment-service/internal/infrastructure/stripe"
	"ride-sharing/services/payment-service/internal/service"
	"ride-sharing/services/payment-service/pkg/types"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"
	"syscall"
)

var (
	GrpcAddr    = env.GetString("GRPC_ADDR", ":9004")
	mq          = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
	appURL      = env.GetString("APP_URL", "http://localhost:3000")
	environment = env.GetString("ENVIRONMENT", "development")
	endpoint    = env.GetString("OTEL_EXPORTER_ENDPOINT", "jaeger:4318")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracerCfg := tracing.Config{
		ServiceName:      "payment-service",
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

	stripeCfg := &types.PaymentConfig{
		StripeSecretKey: env.GetString("STRIPE_SECRET_KEY", ""),
		SuccessURL:      env.GetString("STRIPE_SUCCESS_URL", appURL+"?payment=success"),
		CancelURL:       env.GetString("STRIPE_CANCEL_URL", appURL+"?payment=cancel"),
	}

	if stripeCfg.StripeSecretKey == "" {
		log.Fatalf("STRIPE_SECRET_KEY is not set")
	}

	paymentProcessor := stripe.NewStripeClient(stripeCfg)

	svc := service.NewPaymentService(paymentProcessor)

	log.Println("starting rabbitmq connection")
	mq, err := messaging.NewRabbitMQ(mq)
	if err != nil {
		log.Fatal(err)
	}
	defer mq.Close()

	tripConsumer := events.NewTripConsumer(mq, svc)
	go tripConsumer.Listen()

	<-ctx.Done()
	log.Println("shutting down payment service...")
}
