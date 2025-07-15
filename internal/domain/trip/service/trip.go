package service

import (
	"maryan_api/internal/domain/trip/repo"
)

type Trip interface {
}

type tripService struct {
	tripRepo   repo.Trip
	busRepo    repo.Bus
	driverRepo repo.Driver
}
