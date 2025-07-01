package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	rfc7807 "maryan_api/pkg/problem"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BusDataStore interface {
	RegistrationNumberExists(registrationNumber string, ctx context.Context) (bool, error)
	Create(bus *entity.Bus, ctx context.Context) error
	GetByID(id uuid.UUID, ctx context.Context) (entity.Bus, error)
	GetBuses(pageNumber, pageSize int, order string, ctx context.Context) ([]entity.Bus, int, error)
	Delete(id uuid.UUID, ctx context.Context) error
	IsActive(id uuid.UUID, ctx context.Context) (bool, error)
	MakeActive(id uuid.UUID, ctx context.Context) error
	MakeInactive(id uuid.UUID, ctx context.Context) error
}

type busMySQL struct {
	db *gorm.DB
}

func (brms *busMySQL) Create(bus *entity.Bus, ctx context.Context) error {
	return dbutil.PossibleCreateError(brms.db.WithContext(ctx).Create(&bus), "invalid-bus-params")
}

func (brms *busMySQL) GetByID(id uuid.UUID, ctx context.Context) (entity.Bus, error) {
	var bus entity.Bus
	return bus, dbutil.PossibleFirstError(brms.db.Preload("Rows.Seats").Preload("Images").First(&bus), "non-existing-bus")
}

func (brms *busMySQL) GetBuses(pageNumber, pageSize int, order string, ctx context.Context) ([]entity.Bus, int, error) {
	pageNumber--
	var buses []entity.Bus

	err := dbutil.PossibleRawsAffectedError(brms.db.WithContext(ctx).Limit(pageSize).Offset(pageNumber*pageSize).Order(order).Preload("Images").Find(&buses), "non-existing-page")
	if err != nil {
		return nil, 0, err
	}

	var totalBuses int64

	err = brms.db.Model(&entity.Bus{}).Count(&totalBuses).Error
	if err != nil || totalBuses == 0 {
		return nil, 0, rfc7807.DB("Could not count buses.")
	}

	return buses, int(math.Ceil(float64(totalBuses) / float64(pageSize))), nil

}

func (brms *busMySQL) Delete(id uuid.UUID, ctx context.Context) error {
	return dbutil.PossibleDeleteError(brms.db.WithContext(ctx).Delete(&entity.Bus{ID: id}), "non-existing-bus")
}

func (brms *busMySQL) IsActive(id uuid.UUID, ctx context.Context) (bool, error) {
	var isActive bool
	return isActive, dbutil.PossibleRawsAffectedError(
		brms.db.
			WithContext(ctx).
			Model(&entity.Bus{}).
			Where("id = ?", id).
			Select("is_active").
			Scan(&isActive),
		"non-existing-bus")
}

func (brms *busMySQL) MakeActive(id uuid.UUID, ctx context.Context) error {
	return dbutil.PossibleRawsAffectedError(brms.db.WithContext(ctx).Model(&entity.Bus{}).Where("id = ?", id).Update("is_active", true), "non-existing-bus")
}

func (brms *busMySQL) MakeInactive(id uuid.UUID, ctx context.Context) error {
	return dbutil.PossibleRawsAffectedError(brms.db.WithContext(ctx).Model(&entity.Bus{}).Where("id = ?", id).Update("is_active", false), "non-existing-bus")
}

func (brms *busMySQL) RegistrationNumberExists(registrationNumber string, ctx context.Context) (bool, error) {
	var exists bool
	if err := brms.db.WithContext(ctx).Raw("SELECT EXISTS(SELECT 1 FROM buses WHERE registration_number = ?)", registrationNumber).Error; err != nil {
		return false, err
	}
	return exists, nil
}

// ------------------------Repos Initialization Functions--------------
func NewBusDataStore(db *gorm.DB) BusDataStore {
	return &busMySQL{db}
}
