package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	rfc7807 "maryan_api/pkg/problem"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Passenger interface {
	Create(ctx context.Context, p *entity.Passenger) error
	Update(ctx context.Context, p *entity.Passenger) error
	ForseDelete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Status(ctx context.Context, id uuid.UUID) (exists bool, usedByTicket bool, err error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Passenger, error)
	GetPassengers(ctx context.Context, p dbutil.Pagination) ([]entity.Passenger, int, error)
}

type passengerMySQL struct {
	db *gorm.DB
}

func (pds *passengerMySQL) Create(ctx context.Context, passenger *entity.Passenger) error {
	return dbutil.PossibleCreateError(
		pds.db.WithContext(ctx).Create(passenger),
		"invalid-passenger-data",
	)
}

func (pds *passengerMySQL) Update(ctx context.Context, passenger *entity.Passenger) error {
	return dbutil.PossibleRawsAffectedError(
		pds.db.WithContext(ctx).Save(passenger),
		"invalid-passenger-data",
	)
}

func (pds *passengerMySQL) ForseDelete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		pds.db.WithContext(ctx).Unscoped().Delete(&entity.Passenger{ID: id}),
		"invalid-passenger-data",
	)
}

func (pds *passengerMySQL) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		pds.db.WithContext(ctx).Delete(&entity.Passenger{ID: id}),
		"invalid-passenger-data",
	)
}

func (pds *passengerMySQL) Status(ctx context.Context, id uuid.UUID) (bool, bool, error) {
	var exists bool
	var usedByTicket bool

	if err := pds.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM passengers WHERE id = ?)", id).
		Scan(&exists).Error; err != nil {
		return false, false, rfc7807.DB(err.Error())
	}

	if err := pds.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM tickets WHERE passenger_id = ?)", id).
		Scan(&usedByTicket).Error; err != nil {
		return false, false, rfc7807.DB(err.Error())
	}

	return exists, usedByTicket, nil
}

func (pds *passengerMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Passenger, error) {
	passenger := entity.Passenger{ID: id}
	return passenger, dbutil.PossibleFirstError(
		pds.db.WithContext(ctx).First(&passenger),
		"non-existing-passenger",
	)
}

func (pds *passengerMySQL) GetPassengers(ctx context.Context, p dbutil.Pagination) ([]entity.Passenger, int, error) {
	return dbutil.Paginate[entity.Passenger](ctx, pds.db, p)
}

func NewPassenger(db *gorm.DB) Passenger {
	return &passengerMySQL{db: db}
}
