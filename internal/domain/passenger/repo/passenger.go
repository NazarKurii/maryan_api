package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/dbutil"

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
	GetPassengers(ctx context.Context, pagiantion dbutil.Pagination) ([]entity.Passenger, int, error, bool)
}

type passengerRepo struct {
	ds dataStore.Passenger
}

func (p *passengerRepo) Create(ctx context.Context, passenger *entity.Passenger) error {
	return p.ds.Create(ctx, passenger)
}

func (p *passengerRepo) Update(ctx context.Context, passenger *entity.Passenger) error {
	return p.ds.Update(ctx, passenger)
}

func (p *passengerRepo) ForseDelete(ctx context.Context, id uuid.UUID) error {
	return p.ds.ForseDelete(ctx, id)
}

func (p *passengerRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return p.ds.SoftDelete(ctx, id)
}

func (p *passengerRepo) Status(ctx context.Context, id uuid.UUID) (bool, bool, error) {
	return p.ds.Status(ctx, id)
}

func (p *passengerRepo) GetByID(ctx context.Context, id uuid.UUID) (entity.Passenger, error) {
	return p.ds.GetByID(ctx, id)
}

func (p *passengerRepo) GetPassengers(ctx context.Context, pagiantion dbutil.Pagination) ([]entity.Passenger, int, error, bool) {
	return p.ds.GetPassengers(ctx, pagiantion)
}

func NewPassengerRepoMysql(db *gorm.DB) Passenger {
	return &passengerRepo{dataStore.NewPassenger(db)}
}
