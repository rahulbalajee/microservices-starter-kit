package domain

import (
	"context"
	"ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"

	tripTypes "ride-sharing/services/trip-service/pkg/types"
)

type TripModel struct {
	ID       primitive.ObjectID
	UserID   string
	Status   string
	RideFare *RideFareModel
}

// Store the trip in the in-mem repo (or a database later on)
type TripRepository interface {
	CreateTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
}

// Use-case interface: what the app can do (user intent). Handlers call this; implementation lives in service layer.
type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmAPIResponse, error)
	StartTrip(ctx context.Context, ridefareid, userid string) (string, error)
}
