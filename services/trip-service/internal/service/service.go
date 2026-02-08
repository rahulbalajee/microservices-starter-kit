package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/types"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repo   domain.TripRepository
	client *http.Client
}

func NewService(repo domain.TripRepository) *Service {
	return &Service{
		repo:   repo,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *Service) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	t := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "pending",
		RideFare: fare,
	}

	return s.repo.CreateTrip(ctx, t)
}

func (s *Service) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*types.OsrmAPIResponse, error) {
	url := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson", pickup.Longitude, pickup.Latitude, destination.Longitude, destination.Latitude)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error framing request: %v\n", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling osrm api: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("osrm api returned %d status code\n", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading osrm body: %v\n", err)
	}

	var routeResp types.OsrmAPIResponse
	if err = json.Unmarshal(body, &routeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v\n", err)
	}

	return &routeResp, nil
}
