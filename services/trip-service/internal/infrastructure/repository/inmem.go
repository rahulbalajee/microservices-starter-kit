package repository

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
	"sync"

	pdb "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
)

type inmemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
	mu        sync.RWMutex
}

func NewInmemRepository() *inmemRepository {
	return &inmemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inmemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.trips[trip.ID.Hex()] = trip

	return trip, nil
}

func (r *inmemRepository) SaveRideFare(ctx context.Context, fare *domain.RideFareModel) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rideFares[fare.ID.Hex()] = fare

	return nil
}

func (r *inmemRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFareModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rideFare, exists := r.rideFares[id]
	if !exists {
		return nil, fmt.Errorf("ride fare does not exist with ID: %s\n", id)
	}

	return rideFare, nil
}

func (r *inmemRepository) GetTripByID(ctx context.Context, id string) (*domain.TripModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	trip, ok := r.trips[id]
	if !ok {
		return nil, nil
	}

	return trip, nil
}

func (r *inmemRepository) UpdateTrip(ctx context.Context, tripID string, status string, driver *pdb.Driver) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	trip, ok := r.trips[tripID]
	if !ok {
		return fmt.Errorf("trip not found with id: %s", tripID)
	}

	trip.Status = status

	if driver != nil {
		trip.Driver = &pb.TripDriver{
			Id:             driver.Id,
			Name:           driver.Name,
			CarPlate:       driver.CarPlate,
			ProfilePicture: driver.ProfilePicture,
		}
	}

	return nil
}
