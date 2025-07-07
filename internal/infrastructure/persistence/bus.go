package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"
	"time"

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
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	Available(ctx context.Context, id uuid.UUID, departureTime, arrivalTime time.Time, departureCountry, destinationCountry string) (bool, error)
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
			Preload("Rows.Seats").
			Preload("Images").
			First(&bus),
		"non-existing-bus")
}

func (bds *busMySQL) GetBuses(ctx context.Context, p dbutil.Pagination) ([]entity.Bus, hypermedia.Links, error) {
	return dbutil.Paginate[entity.Bus](ctx, bds.db, p, "Rows.Seats")
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

func (bds *busMySQL) Availability(
	ctx context.Context,
	id uuid.UUID,
	departureTime, arrivalTime time.Time,
	departureCountry string,
) (available struct {
	busy                   bool
	lastDestinationCountry string
	isNew                  bool
	isActive               bool
}, err error) {

	err = bds.db.WithContext(ctx).Raw(`
		SELECT 
			EXISTS (
				SELECT 1 
				FROM trips 
				WHERE bus_id = ? 
				AND (
					(departure_time BETWEEN ? AND ?) 
					OR (arrival_time BETWEEN ? AND ?)
				)
			) AS busy,

			(
				SELECT destination_country 
				FROM trips 
				WHERE bus_id = ? 
				AND arrival_time < ? 
				ORDER BY arrival_time DESC
				LIMIT 1
			) AS lastDestinationCountry,

			(
				SELECT COUNT(*) 
				FROM trips 
				WHERE bus_id = ?
			) = 0 AS isNew,

			(
				SELECT is_active 
				FROM buses 
				WHERE id = ?
			) AS isActive;
	`, id, departureTime, arrivalTime, departureTime, arrivalTime, id, departureTime, id, id).
		Scan(&available).Error

	return available, err
}

// ------------------------Repos Initialization Functions--------------
func NewBus(db *gorm.DB) Bus {
	return &busMySQL{db}
}
