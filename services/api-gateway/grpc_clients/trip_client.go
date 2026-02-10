package grpc_clients

import (
	"fmt"
	"log"
	"os"
	pb "ride-sharing/shared/proto/trip"

	"google.golang.org/grpc"
)

type TripServiceClient struct {
	Client pb.TripServiceClient
	conn   *grpc.ClientConn
}

func NewTripServiceClient() (*TripServiceClient, error) {
	tripServiceURL := os.Getenv("TRIP_SERVICE_GRPC_URL")
	if tripServiceURL == "" {
		tripServiceURL = "trip-service:9093"
	}

	conn, err := grpc.NewClient(tripServiceURL)
	if err != nil {
		return nil, fmt.Errorf("error opening conn: %v\n", err)
	}

	client := pb.NewTripServiceClient(conn)

	return &TripServiceClient{
		Client: client,
		conn:   conn,
	}, nil
}

func (c *TripServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Println(err)
			return
		}
	}
}
