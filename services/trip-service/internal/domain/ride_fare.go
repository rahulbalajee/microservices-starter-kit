package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	pb "ride-sharing/shared/proto/trip"
)

type RideFareModel struct {
	ID                primitive.ObjectID
	UserID            string
	PackageSlug       string //ex: van, sedan, luxury
	TotalPriceInCents float64
	Expires           time.Time
}

func (r *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:                r.ID.Hex(),
		UserID:            r.UserID,
		PackageSlug:       r.PackageSlug,
		TotalPriceInCents: r.TotalPriceInCents,
	}
}

func ToRideFaresProto(fares []*RideFareModel) []*pb.RideFare {
	var protoFares []*pb.RideFare
	for _, f := range fares {
		protoFares = append(protoFares, f.ToProto())
	}
	return protoFares
}
