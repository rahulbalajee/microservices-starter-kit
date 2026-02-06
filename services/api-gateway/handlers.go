package main

import (
	"encoding/json"
	"io"
	"log"
	"maps"
	"net/http"
	"time"

	"ride-sharing/shared/env"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {
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

	// Call trip service (use K8s service name in cluster, localhost for local dev)
	tripServiceURL := env.GetString("TRIP_SERVICE_URL", "http://localhost:8083")
	req, err := http.NewRequest(http.MethodGet, tripServiceURL+"/preview", nil)
	if err != nil {
		http.Error(w, "error framing request", http.StatusInternalServerError)
		return
	}

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		http.Error(w, "error making request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Forward trip service response to client (status, headers, body)
	maps.Copy(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)

	//writeJSON(w, resp.StatusCode, nil)
}
