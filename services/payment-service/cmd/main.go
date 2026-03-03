package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"ride-sharing/services/payment-service/internal/infrastructure/stripe"
	"ride-sharing/services/payment-service/internal/service"
	"ride-sharing/services/payment-service/pkg/types"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"syscall"
)

var (
	GrpcAddr = env.GetString("GRPC_ADDR", ":9004")
	mq       = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
	appURL   = env.GetString("APP_URL", "http://localhost:3000")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	log.Println(svc)

	log.Println("starting rabbitmq connection")
	mq, err := messaging.NewRabbitMQ(mq)
	if err != nil {
		log.Fatal(err)
	}
	defer mq.Close()

	<-ctx.Done()
	log.Println("shutting down payment service...")
}
