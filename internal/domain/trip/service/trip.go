package service

import (
	"context"
	"maryan_api/internal/domain/trip/repo"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"
	rfc7807 "maryan_api/pkg/problem"
	"maryan_api/pkg/timeutil"
	"slices"
	"time"

	"github.com/d3code/uuid"
)

type Trip interface {
	CreateTestTrips(ctx context.Context) error
	Create(ctx context.Context, trip entity.Trip) (uuid.UUID, error)
	GetByID(ctx context.Context, id string) (entity.Trip, error)
	GetTrips(ctx context.Context, pagination dbutil.PaginationStr) ([]entity.TripSimplified, hypermedia.Links, error)
	RegisterUpdate(ctx context.Context, update entity.TripUpdate) error
}

type tripService struct {
	tripRepo  repo.Trip
	busRepo   repo.Bus
	countries repo.Countries
}

func (s *tripService) CreateTestTrips(ctx context.Context) error {
	err := s.tripRepo.DeleteEverythingForTest(ctx)
	if err != nil {
		return err
	}

	countries, err := s.countries.GetAll(ctx)
	if err != nil {
		return err
	}

	ukraine := entity.Country{
		ID:   uuid.MustParse("2296d2ff-77b8-11f0-8b6f-cef5641437d8"),
		Name: "Ukraine",
	}

	countries = slices.Delete(countries, slices.Index(countries, ukraine), slices.Index(countries, ukraine)+1)

	buses, err := s.busRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	var trips = make([]*entity.Trip, 0, 8*53)

	tomorrow := time.Now().Add(time.Hour*24 - time.Hour*time.Duration(time.Now().Hour()))

	for i, country := range countries {
		for j := 0; j < 53; j++ {
			tripID := uuid.New()
			outbondConnectionID := uuid.New()
			returnConnectionID := uuid.New()
			trips = append(trips, &entity.Trip{
				ID: tripID,
				Updates: []entity.TripUpdate{{
					TripID: tripID,
					Status: entity.TripStatusRegistered,
				}},
				OutboundConnection: entity.Connection{
					ID:                   outbondConnectionID,
					Line:                 (i + 1) * 100,
					DepartureCountryID:   ukraine.ID,
					DestinationCountryID: country.ID,
					DepartureTime:        tomorrow.Add(time.Duration(int(time.Hour) * 24 * j * 7)),
					ArrivalTime:          tomorrow.Add(time.Duration(int(time.Hour)*24*j*7 + int(time.Hour)*24*2)),
					GoogleMapsURL:        "",
					BusID:                buses[i].ID,
					Updates: []entity.ConnectionUpdate{
						{
							ConnectionID: outbondConnectionID,
							Status:       entity.RegisteredConnectionStatus,
						},
					},
					Type: entity.ComertialConnectionType,
				},
				ReturnConnection: entity.Connection{
					ID:                   returnConnectionID,
					Line:                 (i + 1) * 100,
					DepartureCountryID:   country.ID,
					DestinationCountryID: ukraine.ID,
					DepartureTime:        tomorrow.Add(time.Duration(int(time.Hour)*24*j*7 + int(time.Hour)*24*3)),
					ArrivalTime:          tomorrow.Add(time.Duration(int(time.Hour)*24*j*7 + int(time.Hour)*24*5)),
					GoogleMapsURL:        "",
					BusID:                buses[i].ID,
					Updates: []entity.ConnectionUpdate{
						{
							ConnectionID: returnConnectionID,
							Status:       entity.RegisteredConnectionStatus,
						},
					},
					Type: entity.ComertialConnectionType,
				},
			})
		}
	}

	return s.tripRepo.TestInsert(ctx, trips)
}

func (s *tripService) Create(ctx context.Context, trip entity.Trip) (uuid.UUID, error) {
	params := trip.Validate()
	if params != nil {
		return uuid.Nil, rfc7807.BadRequest("trip-data", "Trip Data Error", "Provided data is not valid.", params...)
	}

	busExists, err := s.busRepo.Exists(ctx, trip.OutboundConnection.BusID)
	if err != nil {
		return uuid.Nil, err
	} else if !busExists {
		return uuid.Nil, rfc7807.BadRequest("non-existing-bus", "Non-existing Bus Error", "There is no bus assosiated with provided id.")
	}

	available, err := s.busRepo.IsAvailable(ctx, trip.OutboundConnection.BusID, timeutil.DatesBetween(trip.OutboundConnection.DepartureTime, trip.ReturnConnection.ArrivalTime))
	if err != nil {
		return uuid.Nil, err
	} else if !available {
		return uuid.Nil, rfc7807.BadRequest("unavailable-bus", "Unavailable Bus Error", "The bus is unavailble during the trip time")
	}

	trip.PreapareNew()

	if err := s.tripRepo.Create(ctx, &trip); err != nil {
		return uuid.Nil, err
	}

	return trip.ID, nil
}

func (s *tripService) GetByID(ctx context.Context, idStr string) (entity.Trip, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return entity.Trip{}, rfc7807.BadRequest("invalid-id", "Invalid ID Error", err.Error())
	}

	return s.tripRepo.GetByID(ctx, id)
}

func (s *tripService) GetTrips(ctx context.Context, paginationStr dbutil.PaginationStr) ([]entity.TripSimplified, hypermedia.Links, error) {
	pagination, err := paginationStr.Parse(
		[]string{},
	)

	if err != nil {
		return nil, nil, err
	}

	trips, total, err, empty := s.tripRepo.GetTrips(ctx, pagination)
	if err != nil || empty {
		return nil, nil, err
	}

	var tripsSimplified = make([]entity.TripSimplified, len(trips))

	for i, trip := range trips {
		tripsSimplified[i] = trip.Simplify()
	}

	return tripsSimplified, hypermedia.Pagination(paginationStr, total), nil
}

func (s *tripService) RegisterUpdate(ctx context.Context, update entity.TripUpdate) error {
	err := update.Validate()
	if err != nil {
		return err
	}

	return s.tripRepo.RegisterUpdate(ctx, &update)
}

func NewTripService(trip repo.Trip, bus repo.Bus, countries repo.Countries) Trip {
	return &tripService{trip, bus, countries}
}
