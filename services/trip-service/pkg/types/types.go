package types

import (
	pb "ride-sharing/shared/proto/trip"
)

type OsrmAPIResponse struct {
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"routes"`
}

func (o *OsrmAPIResponse) ToProto() *pb.Route {
	route := o.Routes[0]
	coords := route.Geometry.Coordinates
	coordinates := make([]*pb.Coordinate, len(coords))
	for i, coor := range coords {
		coordinates[i] = &pb.Coordinate{
			Latitude:  coor[0],
			Longitude: coor[1],
		}
	}

	return &pb.Route{
		Geometry: []*pb.Geometry{
			{
				Coordinates: coordinates,
			},
		},
		Distance: route.Distance,
		Duration: route.Duration,
	}
}

type PricingConfig struct {
	PricePerUnitOfDistance float64
	PricePerMinute         float64
}

func DefaultPricingConfig() *PricingConfig {
	return &PricingConfig{
		PricePerUnitOfDistance: 1.5,
		PricePerMinute:         0.25,
	}
}
