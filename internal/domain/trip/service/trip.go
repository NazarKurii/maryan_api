package service

import (
	"context"
	"maryan_api/internal/domain/trip/repo"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/google/uuid"
)

type Trip interface {
	Create(ctx context.Context, trip entity.Trip) (uuid.UUID, error)
	GetByID(ctx context.Context, idStr string) (entity.Trip, error)
	GetTripsDriver(ctx context.Context, paginationStr dbutil.PaginationStr, driverID uuid.UUID) ([]entity.Trip, hypermedia.Links, error)
	GetTripsCustomer(ctx context.Context, paginationStr dbutil.PaginationStr, customerID uuid.UUID) ([]entity.Trip, hypermedia.Links, error)
	GetTripsAdmin(ctx context.Context, paginationStr dbutil.PaginationStr) ([]entity.Trip, hypermedia.Links, error)
	ChangeDriver(ctx context.Context, idStr, driverIDStr string) error
	ChangeBus(ctx context.Context, idStr, busIDStr string) error
	ChangeDepartureTime(ctx context.Context, idStr string, time time.Time) error
	Cancel(ctx context.Context, idStr string) error
	Start(ctx context.Context, idStr string) error
	Finish(ctx context.Context, idStr string) error
	MakeSold(ctx context.Context, idStr string) error
}

type tripService struct {
	tripRepo   repo.Trip
	busRepo    repo.Bus
	driverRepo repo.Driver
}

func (t *tripService) Create(ctx context.Context, trip entity.Trip) (uuid.UUID, error) {
	params := trip.Validate()

	exists, err := t.busRepo.Exists(ctx, trip.BusID)
	if err != nil {
		return uuid.Nil, err
	} else if !exists {
		params.SetInvalidParam("busID", "Non-existing bus.")
	}

	isBusy, err := t.busRepo.IsBusy(ctx, trip.BusID, trip.DepartureTime, trip.ArrivalTime, trip.DepartureCountry, trip.DestinationCountry)

	exists, err = t.driverRepo.Exists(ctx, trip.DriverID)
	if err != nil {
		return uuid.Nil, err
	} else if !exists {
		params.SetInvalidParam("DriverID", "Non-existing driver.")
	}

	if params != nil {
		return uuid.Nil, rfc7807.BadRequest("trip-data", "Trip Data Error", "Inva;id params.", params...)
	}

	trip.ID = uuid.New()

	err = t.tripRepo.Create(ctx, &trip)
	if err != nil {
		return uuid.Nil, err
	}

	return trip.ID, nil
}

func (t *tripService) GetByID(ctx context.Context, idStr string) (entity.Trip, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return entity.Trip{}, rfc7807.BadRequest("invalid-uuid", "Invalid UUID Error", err.Error())
	}

	return t.tripRepo.GetByID(ctx, id)
}

func (t *tripService) getTrips(ctx context.Context, paginationStr dbutil.PaginationStr, condition dbutil.Condition) ([]entity.Trip, hypermedia.Links, error) {
	cfg, err := paginationStr.ParseWithCondition(condition, "date")
	if err != nil {
		return nil, nil, err
	}

	return t.tripRepo.GetTrips(ctx, cfg)
}

func (t *tripService) GetTripsDriver(ctx context.Context, paginationStr dbutil.PaginationStr, driverID uuid.UUID) ([]entity.Trip, hypermedia.Links, error) {
	return t.getTrips(ctx, paginationStr, dbutil.Condition{
		Where:  "trips.driver_id = ?",
		Values: []any{driverID},
	})
}

func (t *tripService) GetTripsCustomer(ctx context.Context, paginationStr dbutil.PaginationStr, customerID uuid.UUID) ([]entity.Trip, hypermedia.Links, error) {
	return t.getTrips(ctx, paginationStr, dbutil.Condition{
		Where:  "tickets.customer_id = ?",
		Values: []any{customerID},
	})
}

func (t *tripService) GetTripsAdmin(ctx context.Context, paginationStr dbutil.PaginationStr, driverID uuid.UUID) ([]entity.Trip, hypermedia.Links, error) {
	return t.getTrips(ctx, paginationStr, dbutil.Condition{
		Where:  "trips.driver_id in (?)",
		Values: []any{driverID},
	})
}

func (t *tripService) ChangeBus(ctx context.Context, idstr, busIDStr string) error {
	var params rfc7807.InvalidParams
	id, err := uuid.Parse(idstr)
	if err != nil {
		params.SetInvalidParam("id", err.Error())
	}

	busID, err := uuid.Parse(busIDStr)
	if err != nil {
		params.SetInvalidParam("busID", err.Error())
	}

	if params != nil {
		return rfc7807.BadRequest("invalid-change-trip-bus-data", "Change Trip Bus Data Error", "Invalid params.", params...)
	}

	exists, err := t.busRepo.Exists(ctx, busID)
	if err != nil {
		return err
	}

	if !exists {
		return rfc7807.BadRequest("non-existing-bus", "Non-existing Bus Error", "There is no bus assosiated with provided id.")
	}

	return t.tripRepo.ChangeBus(ctx, id, busID)
}

func (t *tripService) ChangeDepartureTime(ctx context.Context, idStr string, timeStr string) error {
	var params rfc7807.InvalidParams

	time, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		params.SetInvalidParam("time", err.Error())
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		params.SetInvalidParam("id", err.Error())
	}

	return t.tripRepo.ChangeDepartureTime(ctx, id, time)
}

func (t *tripService) Cancel(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return rfc7807.BadRequest("invalid-uuid", "Invalid UUID Error", err.Error())
	}

	return t.tripRepo.Cancel(ctx, id)
}

func (t *tripService) Start(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return rfc7807.BadRequest("invalid-uuid", "Invalid UUID Error", err.Error())
	}

	return t.tripRepo.Start(ctx, id)
}

func (t *tripService) Finish(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return rfc7807.BadRequest("invalid-uuid", "Invalid UUID Error", err.Error())
	}

	return t.tripRepo.Finish(ctx, id)
}

func (t *tripService) MakeSold(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return rfc7807.BadRequest("invalid-uuid", "Invalid UUID Error", err.Error())
	}

	return t.tripRepo.MakeSold(ctx, id)
}

type Stop interface {
	Complete(ctx context.Context, id uuid.UUID) error
	MakeMissed(ctx context.Context, id uuid.UUID) error
}
