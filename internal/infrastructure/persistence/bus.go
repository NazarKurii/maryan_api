package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Bus interface {
	RegistrationNumberExists(ctx context.Context, registrationNumber string) (bool, error)
	Create(ctx context.Context, bus *entity.Bus) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Bus, error)
	GetBuses(ctx context.Context, p dbutil.Pagination) ([]entity.Bus, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	ChangeLeadDriver(ctx context.Context, busID uuid.UUID, driverID uuid.UUID) error
	ChangeAssistantDriver(ctx context.Context, busID uuid.UUID, driverID uuid.UUID) error
	GetAvailable(ctx context.Context, dates []time.Time, pagination dbutil.Pagination) ([]entity.Bus, int, error)
	SetSchedule(ctx context.Context, schedule []entity.BusAvailability) error
}

type busMySQL struct {
	db *gorm.DB
}

func (bds busMySQL) Create(ctx context.Context, bus *entity.Bus) error {
	return dbutil.PossibleCreateError(bds.db.WithContext(ctx).Create(&bus), "invalid-bus-params")
}

func (bds busMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Bus, error) {
	var bus = entity.Bus{ID: id}
	return bus, dbutil.PossibleFirstError(
		bds.db.WithContext(ctx).
			Preload("Rows.Seats").
			Preload(clause.Associations).
			First(&bus),
		"non-existing-bus")
}

func (bds busMySQL) GetBuses(ctx context.Context, p dbutil.Pagination) ([]entity.Bus, int, error) {
	return dbutil.Paginate[entity.Bus](ctx, bds.db.Select("id", "model", "images", "year", "lead_driver", "assistant_driver", "seats"), p, clause.Associations)
}

func (bds busMySQL) Delete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		bds.db.WithContext(ctx).
			Delete(&entity.Bus{ID: id}),
		"non-existing-bus")
}

func (bds busMySQL) IsActive(ctx context.Context, id uuid.UUID) (bool, error) {
	var isActive bool
	return isActive, dbutil.PossibleRawsAffectedError(
		bds.db.WithContext(ctx).
			Model(&entity.Bus{}).
			Where("id = ?", id).
			Select("is_active").
			Scan(&isActive),
		"non-existing-bus")
}

func (bds busMySQL) MakeActive(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		bds.db.WithContext(ctx).
			Model(&entity.Bus{}).
			Where("id = ?", id).
			Update("is_active", true),
		"non-existing-bus")
}

func (bds busMySQL) MakeInactive(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		bds.db.WithContext(ctx).
			Model(&entity.Bus{}).
			Where("id = ?", id).
			Update("is_active", false),
		"non-existing-bus")
}

func (bds busMySQL) RegistrationNumberExists(ctx context.Context, registrationNumber string) (bool, error) {
	var exists bool
	err := bds.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM buses WHERE registration_number = ?)", registrationNumber).
		Scan(&exists).Error
	return exists, err
}

func (bds busMySQL) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := bds.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM buses WHERE id = ?)", id).
		Scan(&exists).Error
	return exists, err
}

func (bds busMySQL) GetAvailable(ctx context.Context, dates []time.Time, pagination dbutil.Pagination) ([]entity.Bus, int, error) {
	return dbutil.Paginate[entity.Bus](ctx, bds.db.
		Table("buses").
		Select("DISTINCT buses.*").
		Joins("JOIN bus_availabilities ON bus_availabilities.bus_id = buses.id").
		Where("bus_availabilities NOT IN (?)", dates), pagination)
}

func (dbs busMySQL) ChangeLeadDriver(ctx context.Context, busID uuid.UUID, driverID uuid.UUID) error {
	return dbutil.PossibleForeignKeyError(dbs.db.WithContext(ctx).Table("buses").Where("id = ?", busID).Update("lead_driver", driverID), "non-existing-bus", "non-existing-driver", "invalid-id")
}

func (dbs busMySQL) ChangeAssistantDriver(ctx context.Context, busID uuid.UUID, driverID uuid.UUID) error {
	return dbutil.PossibleForeignKeyError(dbs.db.WithContext(ctx).Table("buses").Where("id = ?", busID).Update("assistant_driver", driverID), "non-existing-bus", "non-existing-driver", "invalid-id")
}

func (dbs busMySQL) SetSchedule(ctx context.Context, schedule []entity.BusAvailability) error {
	return dbutil.PossibleForeignKeyCreateError(dbs.db.WithContext(ctx).Create(schedule), "non-existing-bus", "bus-schedule-data")
}

// ------------------------Repos Initialization Functions--------------
func NewBus(db *gorm.DB) Bus {
	return &busMySQL{db}
}
