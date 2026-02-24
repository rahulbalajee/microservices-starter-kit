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
	drivers []*driverRecord
	mu      sync.RWMutex
}

type driverRecord struct {
	Driver *pb.Driver
	// Index int
	// TODO: route
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) FindAvailableDrivers(packageType string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var matchingDrivers []string

	for _, driver := range s.drivers {
		if driver.Driver.PackageSlug == packageType {
			matchingDrivers = append(matchingDrivers, driver.Driver.Id)
		}
	}

	if len(matchingDrivers) == 0 {
		return []string{}
	}

	return matchingDrivers
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

	s.drivers = append(s.drivers, &driverRecord{Driver: driver})

	return driver, nil
}

func (s *Service) UnregisterDriver(driverId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, driverRecord := range s.drivers {
		if driverRecord != nil && driverRecord.Driver != nil && driverRecord.Driver.Id == driverId {
			// s.drivers = slices.append(s.drivers[:i], s.drivers[i+1:]...)
			s.drivers = slices.Delete(s.drivers, i, i+1)
			return
		}
	}
}
