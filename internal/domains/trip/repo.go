package trip

import (
	"maryan_api/pkg/dbutil"
	rfc7807 "maryan_api/pkg/problem"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// -----------------Bus--------------------
type busRepo interface {
	registrationNumberExists(registrationNumber string) (bool, error)
	create(bus *Bus) error
	get(id uuid.UUID) (Bus, error)
	getBuses(pageNumber, pageSize int) ([]Bus, int, error)
	delete(id uuid.UUID) error
	isActive(id uuid.UUID) (bool, error)
	makeActive(id uuid.UUID) error
	makeInactive(id uuid.UUID) error
}

type busRepoMySql struct {
	db *gorm.DB
}

func (brms *busRepoMySql) create(bus *Bus) error {
	err := dbutil.PossibleCreateError(brms.db.Create(&bus), "invalid-bus-params")
	if err != nil {
		return err
	}

	return nil

}

func (brms *busRepoMySql) get(id uuid.UUID) (Bus, error) {
	var bus Bus
	return bus, dbutil.PossibleFirstError(brms.db.Preload("Rows.Seats").Preload("Images").First(&bus), "non-existing-bus")
}

func (brms *busRepoMySql) getBuses(pageNumber, pageSize int) ([]Bus, int, error) {
	pageNumber--
	var buses []Bus

	err := dbutil.PossibleRawsAffectedError(brms.db.Limit(pageSize).Offset(pageNumber*pageSize).Preload("Images").Find(&buses), "non-existing-page")
	if err != nil {
		return nil, 0, err
	}

	var totalBuses int64

	err = brms.db.Model(&Bus{}).Count(&totalBuses).Error
	if err != nil || totalBuses == 0 {
		return nil, 0, rfc7807.DB("Could not count buses.")
	}

	return buses, int(math.Ceil(float64(totalBuses) / float64(pageSize))), nil

}

func (brms *busRepoMySql) delete(id uuid.UUID) error {
	return dbutil.PossibleDeleteError(brms.db.Delete(&Bus{ID: id}), "non-existing-bus")
}

func (brms *busRepoMySql) isActive(id uuid.UUID) (bool, error) {
	var isActive bool
	return isActive, dbutil.PossibleRawsAffectedError(brms.db.Model(&Bus{}).Where("id = ?", id).Select("is_active").Scan(&isActive), "non-existing-bus")
}

func (brms *busRepoMySql) makeActive(id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(brms.db.Model(&Bus{}).Where("id = ?", id).Update("is_active", true), "non-existing-bus")
}

func (brms *busRepoMySql) makeInactive(id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(brms.db.Model(&Bus{}).Where("id = ?", id).Update("is_active", false), "non-existing-bus")
}

func (brms *busRepoMySql) registrationNumberExists(registrationNumber string) (bool, error) {
	var exists bool
	if err := brms.db.Raw("SELECT EXISTS(SELECT 1 FROM buses WHERE registration_number = ?)", registrationNumber).Error; err != nil {
		return false, err
	}
	return exists, nil
}

type passengerRepo interface {
}

type passengerRepoMySQL struct {
	db *gorm.DB
}

// ------------------------Repos Initialization Functions--------------
func newBusRepoMySql(db *gorm.DB) busRepo {
	return &busRepoMySql{db}
}

func newPassengerRepoMysql(db *gorm.DB) passengerRepo {
	return passengerRepoMySQL{db}
}
