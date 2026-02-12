package grpc

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type PreviewTripRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserID        string                 `protobuf:"bytes,1,opt,name=userID,proto3" json:"userID,omitempty"`
	StartLocation *Coordinate            `protobuf:"bytes,2,opt,name=startLocation,proto3" json:"startLocation,omitempty"`
	EndLocation   *Coordinate            `protobuf:"bytes,3,opt,name=endLocation,proto3" json:"endLocation,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type Coordinate struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Latitude      float64                `protobuf:"fixed64,1,opt,name=latitude,proto3" json:"latitude,omitempty"`
	Longitude     float64                `protobuf:"fixed64,2,opt,name=longitude,proto3" json:"longitude,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer
	service domain.TripService
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
	}

	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (g *gRPCHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pickup := req.GetStartLocation()
	destination := req.GetEndLocation()

	pickupCoor := &types.Coordinate{
		Latitude:  pickup.Latitude,
		Longitude: pickup.Longitude,
	}

	destinationCoor := &types.Coordinate{
		Latitude:  destination.Latitude,
		Longitude: destination.Longitude,
	}

	t, err := g.service.GetRoute(ctx, pickupCoor, destinationCoor)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get route: %v\n", err)
	}

	return &pb.PreviewTripResponse{
		Route:     t.ToProto(),
		RideFares: []*pb.RideFare{},
	}, nil
}

func (g *gRPCHandler) StartTrip(ctx context.Context, in *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	return nil, nil
}
