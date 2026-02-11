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

	tripPreview, err := app.tripService.Load().Client.PreviewTrip(
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
