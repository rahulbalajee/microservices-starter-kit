package grpc_clients

import (
	"fmt"
	"log"
	"os"
	pb "ride-sharing/shared/proto/driver"
	"ride-sharing/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DriverServiceClient struct {
	Client pb.DriverServiceClient
	conn   *grpc.ClientConn
}

func NewDriverServiceClient() (*DriverServiceClient, error) {
	driverServiceURL := os.Getenv("DRIVER_SERVICE_GRPC_URL")
	if driverServiceURL == "" {
		driverServiceURL = "driver-service:9092"
	}

	dialOpts := append(
		tracing.DialOptionsWithTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	conn, err := grpc.NewClient(driverServiceURL, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("error opening driver service conn: %v", err)
	}

	client := pb.NewDriverServiceClient(conn)

	return &DriverServiceClient{Client: client, conn: conn}, nil
}

func (c *DriverServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Println(err)
		}
	}
}
