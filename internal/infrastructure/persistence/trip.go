package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Trip interface {
	Create(ctx context.Context, trip *entity.Trip) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Trip, error)
	GetTrips(ctx context.Context, pagination dbutil.Pagination) ([]entity.Trip, int, error)
	RegisterUpdate(ctx context.Context, update *entity.TripUpdate) error
}

type tripMySQL struct {
	db *gorm.DB
}

func (ds tripMySQL) Create(ctx context.Context, trip *entity.Trip) error {
	return dbutil.PossibleCreateError(ds.db.WithContext(ctx).Create(trip), "trip-data")
}

func (ds tripMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Trip, error) {
	var trip = entity.Trip{ID: id}
	return trip, dbutil.PossibleFirstError(dbutil.Preload(ds.db.WithContext(ctx), entity.PreloadTrip()...).First(&trip), "non-existing-trip")
}

func (ds tripMySQL) GetTrips(ctx context.Context, pagination dbutil.Pagination) ([]entity.Trip, int, error) {
	return dbutil.Paginate[entity.Trip](ctx, ds.db, pagination, entity.PreloadTrip()...)
}

func (ds tripMySQL) RegisterUpdate(ctx context.Context, update *entity.TripUpdate) error {
	return dbutil.PossibleForeignKeyCreateError(ds.db.WithContext(ctx).Create(update), "non-exisitng-trip", "trip-update-data")
}

func NewTrip(db *gorm.DB) Trip {
	return &tripMySQL{db}
}
