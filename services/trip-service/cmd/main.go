package main

import (
	"context"
	"fmt"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"time"
)

func main() {
	ctx := context.Background()

	inmemRepo := repository.NewInmemRepository()
	svc := service.NewService(inmemRepo)

	fare := &domain.RideFareModel{
		UserID: "42",
	}
	t, err := svc.CreateTrip(ctx, fare)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(t)

	// keep program running for now
	for {
		time.Sleep(time.Second)
	}
}
