package service

import (
	"context"
	"maryan_api/internal/domain/passenger/repo"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"

	rfc7807 "maryan_api/pkg/problem"
	"net/http"

	"github.com/google/uuid"
)

type Passenger interface {
	Create(ctx context.Context, passenger entity.Passenger, userID uuid.UUID) (uuid.UUID, error)
	Update(ctx context.Context, passenger entity.Passenger) (uuid.UUID, error)
	Delete(ctx context.Context, id string) error
	GetPassenger(ctx context.Context, id string) (entity.Passenger, error)
	GetPassengers(ctx context.Context, paginationStr dbutil.PaginationStr, userID uuid.UUID) ([]entity.PassengerSimplified, hypermedia.Links, error)
}

type passengerServiceImpl struct {
	repo   repo.Passenger
	client *http.Client
}

func (p *passengerServiceImpl) Create(ctx context.Context, passenger entity.Passenger, userID uuid.UUID) (uuid.UUID, error) {
	params := passenger.Prepare(userID)
	if params != nil {
		return uuid.Nil, rfc7807.BadRequest("passenger-invalid-data", "Passenger Data Error", "Provided data is not valid.", params...)
	}

	return passenger.ID, p.repo.Create(ctx, &passenger)
}

func (p *passengerServiceImpl) Update(ctx context.Context, passenger entity.Passenger) (uuid.UUID, error) {
	params := passenger.Validate()
	if params != nil {
		return uuid.Nil, rfc7807.BadRequest("passenger-invalid-data", "Passenger Data Error", "Provided data is not valid.", params...)
	}

	exists, usedByTicket, err := p.repo.Status(ctx, passenger.ID)
	if err != nil {
		return uuid.Nil, err
	} else if !exists {
		return uuid.Nil, rfc7807.BadRequest("non-existing-passenger", "Non-existing Passenger Error", "There is no passenger associated with provided id.")
	}

	if !usedByTicket {
		err = p.repo.Update(ctx, &passenger)
		if err != nil {
			return uuid.Nil, err
		}
	} else {
		err = p.repo.SoftDelete(ctx, passenger.ID)
		if err != nil {
			return uuid.Nil, err
		}

		passenger.ID = uuid.New()

		err = p.repo.Create(ctx, &passenger)
		if err != nil {
			return uuid.Nil, err
		}
	}

	return passenger.ID, nil
}

func (p *passengerServiceImpl) Delete(ctx context.Context, idStr string) error {

	id, err := uuid.Parse(idStr)
	if err != nil {
		return rfc7807.BadRequest("invalid-id", "Invalid ID Error", err.Error())
	}

	exists, usedByTicket, err := p.repo.Status(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return rfc7807.BadRequest("non-existing-passenger", "Non-existing Passenger Error", "There is no passenger associated with provided id.")
	}

	if !usedByTicket {
		err = p.repo.ForseDelete(ctx, id)
		if err != nil {
			return err
		}
	} else {
		err = p.repo.SoftDelete(ctx, id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *passengerServiceImpl) GetPassenger(ctx context.Context, idStr string) (entity.Passenger, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return entity.Passenger{}, rfc7807.BadRequest("invalid-id", "Invalid ID Error", err.Error())
	}
	return p.repo.GetByID(ctx, id)
}

func (p *passengerServiceImpl) GetPassengers(ctx context.Context, paginationStr dbutil.PaginationStr, userID uuid.UUID) ([]entity.PassengerSimplified, hypermedia.Links, error) {
	pagination, err := paginationStr.ParseWithCondition(dbutil.Condition{"user_id IN ?", []any{userID}}, "surname", "name", "date_of_birth")
	if err != nil {
		return nil, nil, err
	}

	passengers, links, err := p.repo.GetPassengers(ctx, pagination)
	if err != nil {
		return nil, nil, err
	}

	var passengersSimplified = make([]entity.PassengerSimplified, len(passengers))

	for i, passenger := range passengers {
		passengersSimplified[i] = passenger.Simplify()
	}

	return passengersSimplified, links, nil
}

func NewPassengerService(repo repo.Passenger, client *http.Client) Passenger {
	return &passengerServiceImpl{repo, client}
}
