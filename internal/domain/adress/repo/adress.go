package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/dbutil"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
)

type Address interface {
	Create(ctx context.Context, p *entity.Address) error
	Update(ctx context.Context, p *entity.Address) error
	ForseDelete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Status(ctx context.Context, id uuid.UUID) (exists bool, usedByTicket bool, err error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Address, error)
	GetAddresses(ctx context.Context, p dbutil.Pagination) ([]entity.Address, int, error, bool)
}

type AddressRepo struct {
	ds dataStore.Address
}

func (a *AddressRepo) Create(ctx context.Context, Address *entity.Address) error {
	return a.ds.Create(ctx, Address)
}

func (a *AddressRepo) Update(ctx context.Context, Address *entity.Address) error {
	return a.ds.Update(ctx, Address)
}

func (a *AddressRepo) ForseDelete(ctx context.Context, id uuid.UUID) error {
	return a.ds.ForseDelete(ctx, id)
}

func (a *AddressRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return a.ds.SoftDelete(ctx, id)
}

func (a *AddressRepo) Status(ctx context.Context, id uuid.UUID) (bool, bool, error) {
	return a.ds.Status(ctx, id)
}

func (a *AddressRepo) GetByID(ctx context.Context, id uuid.UUID) (entity.Address, error) {
	return a.ds.GetByID(ctx, id)
}

func (a *AddressRepo) GetAddresses(ctx context.Context, p dbutil.Pagination) ([]entity.Address, int, error, bool) {
	return a.ds.GetAddresses(ctx, p)
}

func NewAddressRepo(db *gorm.DB) Address {
	return &AddressRepo{dataStore.NewAddress(db)}
}
