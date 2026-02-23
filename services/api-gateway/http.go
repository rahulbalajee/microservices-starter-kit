package main

import (
	"encoding/json"
	"log"
	"net/http"

	"ride-sharing/shared/contracts"
)

func (app *application) handleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// TODO: Add more validation
	if reqBody.UserID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	tripService := app.tripService.Load()
	if tripService == nil {
		http.Error(w, "trip service unavailable", http.StatusServiceUnavailable)
		return
	}

	tripPreview, err := tripService.Client.PreviewTrip(
		r.Context(),
		reqBody.toProto(),
	)
	if err != nil {
		log.Printf("failed to preview trip: %v\n", err)
		http.Error(w, "failed to preview trip", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: tripPreview}
	writeJSON(w, http.StatusCreated, response)
}

func (app *application) handleTripStart(w http.ResponseWriter, r *http.Request) {
	var reqBody startTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// TODO: Add more validation
	if reqBody.UserID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	tripService := app.tripService.Load()
	if tripService == nil {
		http.Error(w, "trip service unavailable", http.StatusServiceUnavailable)
		return
	}

	tripStart, err := tripService.Client.CreateTrip(
		r.Context(),
		reqBody.toProto(),
	)
	if err != nil {
		log.Printf("failed to start trip: %v\n", err)
		http.Error(w, "failed to start trip", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: tripStart}
	writeJSON(w, http.StatusOK, response)
}
