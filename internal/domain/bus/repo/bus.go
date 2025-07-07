package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Bus interface {
	RegistrationNumberExists(ctx context.Context, registrationNumber string) (bool, error)
	Create(ctx context.Context, bus *entity.Bus) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Bus, error)
	GetBuses(ctx context.Context, p dbutil.Pagination) ([]entity.Bus, hypermedia.Links, error)
	Delete(ctx context.Context, id uuid.UUID) error
	IsActive(ctx context.Context, id uuid.UUID) (bool, error)
	MakeActive(ctx context.Context, id uuid.UUID) error
	MakeInactive(ctx context.Context, id uuid.UUID) error
}

type busRepo struct {
	store dataStore.Bus
}

func (b *busRepo) Create(ctx context.Context, bus *entity.Bus) error {
	return b.store.Create(ctx, bus)
}

func (b *busRepo) GetByID(ctx context.Context, id uuid.UUID) (entity.Bus, error) {
	return b.store.GetByID(ctx, id)
}

func (b *busRepo) GetBuses(ctx context.Context, p dbutil.Pagination) ([]entity.Bus, hypermedia.Links, error) {
	return b.store.GetBuses(ctx, p)
}

func (b *busRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return b.store.Delete(ctx, id)
}

func (b *busRepo) IsActive(ctx context.Context, id uuid.UUID) (bool, error) {
	return b.store.IsActive(ctx, id)
}

func (b *busRepo) MakeActive(ctx context.Context, id uuid.UUID) error {
	return b.store.MakeActive(ctx, id)
}

func (b *busRepo) MakeInactive(ctx context.Context, id uuid.UUID) error {
	return b.store.MakeInactive(ctx, id)
}

func (b *busRepo) RegistrationNumberExists(ctx context.Context, registrationNumber string) (bool, error) {
	return b.store.RegistrationNumberExists(ctx, registrationNumber)
}

// ------------------------Repos Initialization Functions--------------
func NewBusRepo(db *gorm.DB) Bus {
	return &busRepo{dataStore.NewBus(db)}
}
