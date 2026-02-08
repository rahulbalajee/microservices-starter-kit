package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"ride-sharing/shared/contracts"
	"ride-sharing/shared/env"
)

func (app *application) handleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validation
	if reqBody.UserID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}
	reader := bytes.NewReader(jsonBody)

	// Call trip service (use K8s service name in cluster, localhost for local dev)
	tripServiceURL := env.GetString("TRIP_SERVICE_URL", "http://localhost:8083")
	req, err := http.NewRequest(http.MethodPost, tripServiceURL+"/preview", reader)
	if err != nil {
		http.Error(w, "error framing request", http.StatusInternalServerError)
		return
	}

	resp, err := app.client.Do(req)
	if err != nil {
		log.Println(err)
		http.Error(w, "error making request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var respBody any
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	response := contracts.APIResponse{Data: respBody}
	writeJSON(w, http.StatusCreated, response)
}
