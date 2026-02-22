package grpc

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
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
	service   domain.TripService
	publisher *events.TripEventPublisher
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService, publisher *events.TripEventPublisher) *gRPCHandler {
	handler := &gRPCHandler{
		service:   service,
		publisher: publisher,
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

	userID := req.GetUserID()

	route, err := g.service.GetRoute(ctx, pickupCoor, destinationCoor)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get route: %v\n", err)
	}

	estimatedFares := g.service.EstimatePackagesPriceWithRoute(route)

	fares, err := g.service.GenerateTripFares(ctx, estimatedFares, userID, route)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate ride fare: %v\n", err)
	}

	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}

func (g *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	fareID := req.GetRideFareID()
	userID := req.GetUserID()

	fare, err := g.service.GetAndValidateFare(ctx, fareID, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate fare: %v\n", err)
	}

	trip, err := g.service.CreateTrip(ctx, fare)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create the trip: %v\n", err)
	}

	if err = g.publisher.PublishTripCreated(ctx, trip); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish the trip created event: %v\n", err)
	}

	return &pb.CreateTripResponse{TripID: trip.ID.Hex()}, nil
}
