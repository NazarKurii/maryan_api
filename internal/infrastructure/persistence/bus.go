package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Bus interface {
	RegistrationNumberExists(ctx context.Context, registrationNumber string) (bool, error)
	Create(ctx context.Context, bus *entity.Bus) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Bus, error)
	GetBuses(ctx context.Context, p dbutil.Pagination) ([]entity.Bus, int, error, bool)
	Delete(ctx context.Context, id uuid.UUID) error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	ChangeLeadDriver(ctx context.Context, busID uuid.UUID, driverID uuid.UUID) error
	ChangeAssistantDriver(ctx context.Context, busID uuid.UUID, driverID uuid.UUID) error
	GetAvailable(ctx context.Context, dates []time.Time, pagination dbutil.Pagination) ([]entity.Bus, int, error, bool)
	SetSchedule(ctx context.Context, schedule []entity.BusAvailability) error
	IsAvailable(ctx context.Context, id uuid.UUID, dates []time.Time) (bool, error)
	GetAll(ctx context.Context) ([]entity.Bus, error)
}

type busMySQL struct {
	db *gorm.DB
}

func (bds *busMySQL) Create(ctx context.Context, bus *entity.Bus) error {
	return dbutil.PossibleCreateError(bds.db.WithContext(ctx).Create(&bus), "invalid-bus-params")
}

func (bds *busMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Bus, error) {
	var bus = entity.Bus{ID: id}
	return bus, dbutil.PossibleFirstError(
		bds.db.WithContext(ctx).
			Preload("Structure.Positions").
			Preload(clause.Associations).
			First(&bus),
		"non-existing-bus")
}

func (bds *busMySQL) GetBuses(ctx context.Context, p dbutil.Pagination) ([]entity.Bus, int, error, bool) {
	return dbutil.Paginate[entity.Bus](ctx, bds.db, p, clause.Associations)
}

func (bds *busMySQL) Delete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		bds.db.WithContext(ctx).
			Delete(&entity.Bus{ID: id}),
		"non-existing-bus")
}

func (bds *busMySQL) IsActive(ctx context.Context, id uuid.UUID) (bool, error) {
	var isActive bool
	return isActive, dbutil.PossibleRawsAffectedError(
		bds.db.WithContext(ctx).
			Model(&entity.Bus{}).
			Where("id = ?", id).
			Select("is_active").
			Scan(&isActive),
		"non-existing-bus")
}

func (bds *busMySQL) MakeActive(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		bds.db.WithContext(ctx).
			Model(&entity.Bus{}).
			Where("id = ?", id).
			Update("is_active", true),
		"non-existing-bus")
}

func (bds *busMySQL) MakeInactive(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		bds.db.WithContext(ctx).
			Model(&entity.Bus{}).
			Where("id = ?", id).
			Update("is_active", false),
		"non-existing-bus")
}

func (bds *busMySQL) RegistrationNumberExists(ctx context.Context, registrationNumber string) (bool, error) {
	var exists bool
	err := bds.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM buses WHERE registration_number = ?)", registrationNumber).
		Scan(&exists).Error
	return exists, err
}

func (bds *busMySQL) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := bds.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM buses WHERE id = ?)", id).
		Scan(&exists).Error
	return exists, err
}

func (bds *busMySQL) GetAvailable(ctx context.Context, dates []time.Time, pagination dbutil.Pagination) ([]entity.Bus, int, error, bool) {
	return dbutil.Paginate[entity.Bus](ctx, bds.db.
		Table("buses").
		Select("DISTINCT buses.*").
		Joins("JOIN bus_availabilities ON bus_availabilities.bus_id = buses.id").
		Where("bus_availabilities.date NOT IN (?)", dates), pagination)
}

func (dbs *busMySQL) ChangeLeadDriver(ctx context.Context, busID uuid.UUID, driverID uuid.UUID) error {
	return dbutil.PossibleForeignKeyError(dbs.db.WithContext(ctx).Table("buses").Where("id = ?", busID).Update("lead_driver", driverID), "non-existing-bus", "non-existing-driver", "invalid-id")
}

func (dbs *busMySQL) ChangeAssistantDriver(ctx context.Context, busID uuid.UUID, driverID uuid.UUID) error {
	return dbutil.PossibleForeignKeyError(dbs.db.WithContext(ctx).Table("buses").Where("id = ?", busID).Update("assistant_driver", driverID), "non-existing-bus", "non-existing-driver", "invalid-id")
}

func (dbs *busMySQL) SetSchedule(ctx context.Context, schedule []entity.BusAvailability) error {
	return dbutil.PossibleForeignKeyCreateError(dbs.db.WithContext(ctx).Create(schedule), "non-existing-bus", "bus-schedule-data")
}

func (dbs *busMySQL) IsAvailable(ctx context.Context, id uuid.UUID, dates []time.Time) (bool, error) {
	var available bool

	err := dbs.db.WithContext(ctx).Raw("SELECT EXISTS (SELECT FROM buses AS b JOIN bus_availabilities AS ba ON ba.bus_id = b.id)  WHERE b.id = ? AND ba.date NOT IN (?) ", id, dates).Scan(&available).Error
	if err != nil {
		return false, rfc7807.DB(err.Error())
	}

	return available, nil
}

func (dbs *busMySQL) GetAll(ctx context.Context) ([]entity.Bus, error) {
	var buses []entity.Bus

	return buses, dbutil.PossibleRawsAffectedError(dbs.db.WithContext(ctx).Find(&buses), "no-buses-yet")
}

// ------------------------Repos Initialization Functions--------------
func NewBus(db *gorm.DB) Bus {
	return &busMySQL{db}
}
