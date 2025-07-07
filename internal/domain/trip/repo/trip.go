package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"

	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Trip interface {
	Create(ctx context.Context, trip *entity.Trip) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Trip, error)
	GetTrips(ctx context.Context, p dbutil.CondtionPagination) ([]entity.Trip, hypermedia.Links, error)
	ChangeDriver(ctx context.Context, id, driverID uuid.UUID) error
	ChangeBus(ctx context.Context, id, busID uuid.UUID) error
	ChangeDepartureTime(ctx context.Context, id uuid.UUID, time time.Time) error
	Cancel(ctx context.Context, id uuid.UUID) error
	Start(ctx context.Context, id uuid.UUID) error
	Finish(ctx context.Context, id uuid.UUID) error
	MakeSold(ctx context.Context, id uuid.UUID) error
}

type Bus interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type Driver interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type Stop interface {
	Complete(ctx context.Context, id uuid.UUID) error
	MakeMissed(ctx context.Context, id uuid.UUID) error
}

type tripRepo struct {
	ds dataStore.Trip
}

func (t *tripRepo) Create(ctx context.Context, trip *entity.Trip) error {
	return t.ds.Create(ctx, trip)
}

func (t *tripRepo) GetByID(ctx context.Context, id uuid.UUID) (entity.Trip, error) {
	return t.ds.GetByID(ctx, id)
}

func (t *tripRepo) GetTrips(ctx context.Context, p dbutil.CondtionPagination) ([]entity.Trip, hypermedia.Links, error) {
	return t.ds.GetTrips(ctx, p)
}

func (t *tripRepo) ChangeDriver(ctx context.Context, id, driverID uuid.UUID) error {
	return t.ds.ChangeDriver(ctx, id, driverID)
}

func (t *tripRepo) ChangeBus(ctx context.Context, id, busID uuid.UUID) error {
	return t.ds.ChangeBus(ctx, id, busID)
}

func (t *tripRepo) ChangeDepartureTime(ctx context.Context, id uuid.UUID, time time.Time) error {
	return t.ds.ChangeDepartureTime(ctx, id, time)
}

func (t *tripRepo) Cancel(ctx context.Context, id uuid.UUID) error {
	return t.ds.Cancel(ctx, id)
}

func (t *tripRepo) Start(ctx context.Context, id uuid.UUID) error {
	return t.ds.Start(ctx, id)
}

func (t *tripRepo) Finish(ctx context.Context, id uuid.UUID) error {
	return t.ds.Finish(ctx, id)
}

func (t *tripRepo) MakeSold(ctx context.Context, id uuid.UUID) error {
	return t.ds.MakeSold(ctx, id)
}

func NewTrip(db *gorm.DB) Trip {
	return &tripRepo{dataStore.NewTrip(db)}
}

type driverRepo struct {
	ds dataStore.User
}

func (d *driverRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return d.ds.UserExistsByID(ctx, id)
}

func NewDriver(db *gorm.DB) Driver {
	return &driverRepo{dataStore.NewUser(db)}
}

type busRepo struct {
	ds dataStore.Bus
}

func (d *busRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return d.ds.Exists(ctx, id)
}

func NewBus(db *gorm.DB) Bus {
	return &busRepo{dataStore.NewBus(db)}
}

type stopRepo struct {
	ds dataStore.Stop
}

func (s stopRepo) Complete(ctx context.Context, id uuid.UUID) error {
	return s.ds.Complete(ctx, id)
}
func (s stopRepo) MakeMissed(ctx context.Context, id uuid.UUID) error {
	return s.ds.MakeMissed(ctx, id)
}

func NewStop(db *gorm.DB) Stop {
	return &stopRepo{dataStore.NewStop(db)}
}
