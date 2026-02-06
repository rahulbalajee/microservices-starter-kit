package main

import (
	"log"
	"net/http"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"
)

var (
	httpAddr = env.GetString("HTTP_ADDR", ":8083")
)

type application struct {
	svc *service.Service
}

func main() {
	inmemRepo := repository.NewInmemRepository()
	svc := service.NewService(inmemRepo)

	app := &application{
		svc: svc,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /preview", app.createTrip)

	srv := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server error: %v\n", err)
	}
}
