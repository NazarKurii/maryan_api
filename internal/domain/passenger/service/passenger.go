package service

import (
	"maryan_api/internal/domain/passenger/repo"
	"net/http"
)

type PassengerService interface {
}

type passengerServiceImpl struct {
	repo   repo.PassengerRepo
	client *http.Client
}

func NewPassengerService(passenger repo.PassengerRepo, client *http.Client) PassengerService {
	return &passengerServiceImpl{passenger, client}
}
