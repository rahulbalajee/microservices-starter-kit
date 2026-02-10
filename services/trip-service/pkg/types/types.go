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

func (o *OsrmAPIResponse) toProto() *pb.Route {
	return &pb.Route{}
}
