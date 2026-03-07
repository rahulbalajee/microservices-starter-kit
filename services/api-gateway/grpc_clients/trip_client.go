package grpc_clients

import (
	"fmt"
	"log"
	"os"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	dialOpts := append(
		tracing.DialOptionsWithTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	conn, err := grpc.NewClient(tripServiceURL, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("error opening trip service conn: %v", err)
	}

	client := pb.NewTripServiceClient(conn)

	return &TripServiceClient{Client: client, conn: conn}, nil
}

func (c *TripServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Println(err)
		}
	}
}
