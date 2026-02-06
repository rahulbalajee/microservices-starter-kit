package main

import (
	"context"
	"log"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
)

func (app *application) createTrip(w http.ResponseWriter, r *http.Request) {
	fare := &domain.RideFareModel{
		UserID: "32",
	}

	_, err := app.svc.CreateTrip(context.Background(), fare)
	if err != nil {
		log.Println(err)
	}

	w.Write([]byte("Trip created!\n"))
}
