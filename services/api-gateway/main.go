package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
)

var (
	httpAddr    = env.GetString("HTTP_ADDR", ":8081")
	rabbitmqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

type application struct {
	mq            *messaging.RabbitMQ
	client        *http.Client
	tripService   *atomic.Pointer[grpc_clients.TripServiceClient]
	driverService *atomic.Pointer[grpc_clients.DriverServiceClient]
}

func main() {
	log.Println("Starting API Gateway")

	maxBackoff := 30 * time.Second

	tripServicePtr := &atomic.Pointer[grpc_clients.TripServiceClient]{}
	defer func() {
		if c := tripServicePtr.Load(); c != nil {
			c.Close()
		}
	}()

	go connectWithBackoff(
		maxBackoff,
		"trip service",
		tripServicePtr,
		grpc_clients.NewTripServiceClient,
	)

	driverServicePtr := &atomic.Pointer[grpc_clients.DriverServiceClient]{}
	defer func() {
		if c := driverServicePtr.Load(); c != nil {
			c.Close()
		}
	}()

	go connectWithBackoff(
		maxBackoff,
		"driver service",
		driverServicePtr,
		grpc_clients.NewDriverServiceClient,
	)

	mq, err := messaging.NewRabbitMQ(rabbitmqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer mq.Close()

	app := application{
		mq:            mq,
		client:        &http.Client{Timeout: 10 * time.Second},
		tripService:   tripServicePtr,
		driverService: driverServicePtr,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /trip/preview", enableCORS(app.handleTripPreview))
	mux.HandleFunc("POST /trip/start", enableCORS(app.handleTripStart))
	mux.HandleFunc("/ws/riders", app.handleRidersWebSocket)
	mux.HandleFunc("/ws/drivers", app.handleDriversWebSocket)

	srv := &http.Server{
		Addr:              httpAddr,
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       time.Minute,
		MaxHeaderBytes:    1 << 20,
	}

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("server listening on %s\n", httpAddr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Printf("error starting the server: %v\n", err)
	case sig := <-shutdown:
		log.Printf("server is shutting down due to %v signal\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown server gracefully: %v\n", err)
			srv.Close()
		}
	}
}
