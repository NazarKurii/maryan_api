package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/dbutil"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Bus interface {
	RegistrationNumberExists(ctx context.Context, registrationNumber string) (bool, error)
	Create(ctx context.Context, bus *entity.Bus) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Bus, error)
	GetBuses(ctx context.Context, p dbutil.Pagination) ([]entity.Bus, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ChangeLeadDriver(ctx context.Context, busID uuid.UUID, driverID uuid.UUID) error
	ChangeAssistantDriver(ctx context.Context, busID uuid.UUID, driverID uuid.UUID) error
	GetAvailable(ctx context.Context, dates []time.Time, pagination dbutil.Pagination) ([]entity.Bus, int, error)
	SetSchedule(ctx context.Context, schedule []entity.BusAvailability) error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type Driver interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type driverRepo struct {
	store dataStore.Driver
}

func (d *driverRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return d.store.UserExists(ctx, id)
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

func (b *busRepo) GetBuses(ctx context.Context, p dbutil.Pagination) ([]entity.Bus, int, error) {
	return b.store.GetBuses(ctx, p)
}

func (b *busRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return b.store.Delete(ctx, id)
}

func (b *busRepo) RegistrationNumberExists(ctx context.Context, registrationNumber string) (bool, error) {
	return b.store.RegistrationNumberExists(ctx, registrationNumber)
}

func (b *busRepo) ChangeLeadDriver(ctx context.Context, busID uuid.UUID, driverID uuid.UUID) error {
	return b.store.ChangeLeadDriver(ctx, busID, driverID)
}

func (b *busRepo) ChangeAssistantDriver(ctx context.Context, busID uuid.UUID, driverID uuid.UUID) error {
	return b.store.ChangeAssistantDriver(ctx, busID, driverID)
}

func (b *busRepo) GetAvailable(ctx context.Context, dates []time.Time, pagination dbutil.Pagination) ([]entity.Bus, int, error) {
	return b.store.GetAvailable(ctx, dates, pagination)
}

func (b *busRepo) SetSchedule(ctx context.Context, schedule []entity.BusAvailability) error {
	return b.store.SetSchedule(ctx, schedule)
}

func (b *busRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return b.store.Exists(ctx, id)
}

// ------------------------Repos Initialization Functions--------------
func NewBusRepo(db *gorm.DB) Bus {
	return &busRepo{dataStore.NewBus(db)}
}

func NewDriverRepo(db *gorm.DB) Driver {
	return &driverRepo{dataStore.NewDriver(db)}
}
