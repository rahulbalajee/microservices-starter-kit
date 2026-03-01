package messaging

import (
	pdb "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
)

const (
	FindAvailableDriversQueue = "find_available_driver"
	DriverCmdTripRequestQueue = "driver_cmd_trip_request"
	DriverTripResponseQueue   = "driver_trip_response"
	NotifyDriverNotFound      = "notify_driver_not_found"
	NotifyDriverAssignQueue   = "notify_driver_assign"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}

type DriverTripResponseData struct {
	Driver  *pdb.Driver `json:"driver"`
	TripID  string      `json:"tripID"`
	RiderID string      `json:"riderID"`
}
