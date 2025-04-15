package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/dbutil"
	"time"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
)

type Trip interface {
	Create(ctx context.Context, trip *entity.Trip) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Trip, error)
	GetTrips(ctx context.Context, pagination dbutil.Pagination) ([]entity.Trip, int, error, bool)
	DeleteEverythingForTest(ctx context.Context) error
	RegisterUpdate(ctx context.Context, update *entity.TripUpdate) error
	TestInsert(ctx context.Context, trips []*entity.Trip) error
}

type tripRepo struct {
	ds dataStore.Trip
}

func (r *tripRepo) Create(ctx context.Context, trip *entity.Trip) error {
	return r.ds.Create(ctx, trip)
}

func (r *tripRepo) GetByID(ctx context.Context, id uuid.UUID) (entity.Trip, error) {
	return r.ds.GetByID(ctx, id)
}

func (r *tripRepo) GetTrips(ctx context.Context, pagination dbutil.Pagination) ([]entity.Trip, int, error, bool) {
	return r.ds.GetTrips(ctx, pagination)
}

func (r *tripRepo) RegisterUpdate(ctx context.Context, update *entity.TripUpdate) error {
	return r.ds.RegisterUpdate(ctx, update)
}

func (r *tripRepo) DeleteEverythingForTest(ctx context.Context) error {

	return r.ds.DeleteEverythingForTest(ctx)
}

func (r *tripRepo) TestInsert(ctx context.Context, trips []*entity.Trip) error {

	return r.ds.Test(ctx, trips)
}

func NewTrip(db *gorm.DB) Trip {
	return &tripRepo{ds: dataStore.NewTrip(db)}
}

type Bus interface {
	IsAvailable(ctx context.Context, id uuid.UUID, dates []time.Time) (bool, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	GetAll(ctx context.Context) ([]entity.Bus, error)
}

type busRepo struct {
	ds dataStore.Bus
}

func (r busRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.ds.Exists(ctx, id)
}

func (r busRepo) IsAvailable(ctx context.Context, id uuid.UUID, dates []time.Time) (bool, error) {
	return r.ds.IsAvailable(ctx, id, dates)
}

func (r busRepo) GetAll(ctx context.Context) ([]entity.Bus, error) {
	return r.ds.GetAll(ctx)
}

func NewBus(db *gorm.DB) Bus {
	return &busRepo{dataStore.NewBus(db)}
}

type Countries interface {
	GetAll(ctx context.Context) ([]entity.Country, error)
}

type countreisRepo struct {
	ds dataStore.Country
}

func (c countreisRepo) GetAll(ctx context.Context) ([]entity.Country, error) {
	return c.ds.GetAll(ctx)
}

func NewCountry(db *gorm.DB) Countries {
	return &countreisRepo{dataStore.NewCountry(db)}
}
