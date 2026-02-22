package domain

import (
	"context"
	"ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"

	tripTypes "ride-sharing/services/trip-service/pkg/types"

	pb "ride-sharing/shared/proto/trip"
)

type TripModel struct {
	ID       primitive.ObjectID
	UserID   string
	Status   string
	RideFare *RideFareModel
	Driver   *pb.TripDriver
}

func (t *TripModel) ToProto() *pb.Trip {
	return &pb.Trip{
		Id:           t.ID.Hex(),
		UserID:       t.UserID,
		SelectedFare: t.RideFare.ToProto(),
		Status:       t.Status,
		Driver:       t.Driver,
		Route:        t.RideFare.Route.ToProto(),
	}
}

// Store the trip in the in-mem repo (or a database later on)
type TripRepository interface {
	CreateTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
	SaveRideFare(ctx context.Context, f *RideFareModel) error
	GetRideFareByID(ctx context.Context, id string) (*RideFareModel, error)
}

// Use-case interface: what the app can do (user intent). Handlers call this; implementation lives in service layer.
type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmAPIResponse, error)
	EstimatePackagesPriceWithRoute(route *tripTypes.OsrmAPIResponse) []*RideFareModel
	GenerateTripFares(ctx context.Context, fares []*RideFareModel, userID string, route *tripTypes.OsrmAPIResponse) ([]*RideFareModel, error)
	GetAndValidateFare(ctx context.Context, fareID, userID string) (*RideFareModel, error)
}
