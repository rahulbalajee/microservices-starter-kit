package messaging

import (
	pb "ride-sharing/shared/proto/trip"
)

const (
	FindAvailableDriversQueue = "find_available_driver"
	DriverCmdTripRequestQueue = "driver_cmd_trip_request"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}
