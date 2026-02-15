package main

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "ride-sharing/shared/proto/driver"
)

type grpcHandler struct {
	pb.UnimplementedDriverServiceServer
	Service *Service
}

func NewGrpcHandler(s *grpc.Server, service *Service) {
	handler := &grpcHandler{
		Service: service,
	}
	pb.RegisterDriverServiceServer(s, handler)
}

func (g *grpcHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	driverId := req.GetDriverID()
	packageSlug := req.GetPackageSlug()

	driver, err := g.Service.RegisterDriver(driverId, packageSlug)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register driver: %v", err)
	}

	return &pb.RegisterDriverResponse{Driver: driver}, nil
}

func (g *grpcHandler) UnregisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	driverId := req.GetDriverID()

	g.Service.UnregisterDriver(driverId)

	return &pb.RegisterDriverResponse{
		Driver: &pb.Driver{
			Id: driverId,
		},
	}, nil
}
