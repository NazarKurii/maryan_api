package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/pagination"

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
	GetPassengers(ctx context.Context, cfg pagination.CfgCondition) ([]entity.Passenger, int, error)
}

type passengerRepoMySQL struct {
	ds dataStore.Passenger
}

func (p *passengerRepoMySQL) Create(ctx context.Context, passenger *entity.Passenger) error {
	return p.ds.Create(ctx, passenger)
}

func (p *passengerRepoMySQL) Update(ctx context.Context, passenger *entity.Passenger) error {
	return p.ds.Update(ctx, passenger)
}

func (p *passengerRepoMySQL) ForseDelete(ctx context.Context, id uuid.UUID) error {
	return p.ds.ForseDelete(ctx, id)
}

func (p *passengerRepoMySQL) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return p.ds.SoftDelete(ctx, id)
}

func (p *passengerRepoMySQL) Status(ctx context.Context, id uuid.UUID) (bool, bool, error) {
	return p.ds.Status(ctx, id)
}

func (p *passengerRepoMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Passenger, error) {
	return p.ds.GetByID(ctx, id)
}

func (p *passengerRepoMySQL) GetPassengers(ctx context.Context, cfg pagination.CfgCondition) ([]entity.Passenger, int, error) {
	return p.ds.GetPassengers(ctx, cfg)
}

func NewPassengerRepoMysql(db *gorm.DB) Passenger {
	return &passengerRepoMySQL{dataStore.NewPassenger(db)}
}
