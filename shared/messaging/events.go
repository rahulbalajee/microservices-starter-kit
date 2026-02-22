package messaging

import (
	pb "ride-sharing/shared/proto/trip"
)

const (
	FindAvailableDriversQueue = "find_available_driver"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}
