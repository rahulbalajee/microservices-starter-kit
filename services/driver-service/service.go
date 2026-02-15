package main

import (
	math "math/rand/v2"
	pb "ride-sharing/shared/proto/driver"
	"ride-sharing/shared/util"
	"slices"
	"sync"

	"github.com/mmcloughlin/geohash"
)

type Service struct {
	drivers []*driverInMap
	mu      sync.RWMutex
}

type driverInMap struct {
	Driver *pb.Driver
	// Index int
	// TODO: route
}

func NewService() *Service {
	return &Service{
		drivers: make([]*driverInMap, 0),
	}
}

func (s *Service) RegisterDriver(driverId string, packageSlug string) (*pb.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	randomIdx := math.IntN(len(PredefinedRoutes))
	randomRoute := PredefinedRoutes[randomIdx]

	geohash := geohash.Encode(randomRoute[0][0], randomRoute[0][1])

	driver := &pb.Driver{
		Geohash:        geohash,
		Location:       &pb.Location{Latitude: randomRoute[0][0], Longitude: randomRoute[0][1]},
		Name:           "Chuck Norris",
		Id:             driverId,
		PackageSlug:    packageSlug,
		ProfilePicture: util.GetRandomAvatar(randomIdx),
		CarPlate:       GenerateRandomPlate(),
	}

	s.drivers = append(s.drivers, &driverInMap{Driver: driver})

	return driver, nil
}

func (s *Service) UnregisterDriver(driverId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, driverInMap := range s.drivers {
		if driverInMap != nil && driverInMap.Driver != nil && driverInMap.Driver.Id == driverId {
			// s.drivers = slices.append(s.drivers[:i], s.drivers[i+1:]...)
			s.drivers = slices.Delete(s.drivers, i, i+1)
			return
		}
	}
}
