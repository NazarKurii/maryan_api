package entity

import (
	googleMaps "maryan_api/internal/infrastructure/clients/google/maps"
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Adress struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey"         json:"id"`
	User_ID         uuid.UUID      `gorm:"type:uuid; not null"          json:"-"`
	Country         string         `gorm:"type:varchar(56);not null"    json:"country"`
	City            string         `gorm:"type:varchar(56);not null"    json:"city"`
	Street          string         `gorm:"type:varchar(255);not null"   json:"street"`
	HouseNumber     int            `gorm:"type:smallint; not null"      json:"houseNumber"`
	ApartmentNumber int            `gorm:"type:smallint"                json:"apartmentNumber"`
	GoogleMapsLink  string         `gorm:"type:varchar(255);not null"   json:"googleMapsLink"`
	CreatedAt       time.Time      `gorm:"not null"                     json:"-"`
	DeletedAt       gorm.DeletedAt `                                    json:"-"`
}

func (a Adress) Validate() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams

	if a.Country == "" {
		params.SetInvalidParam("country", "Invalid country.")
	}
	if a.City == "" {
		params.SetInvalidParam("city", "Invalid city.")
	}
	if a.Street == "" {
		params.SetInvalidParam("street", "Invalid street.")
	}
	if a.HouseNumber <= 0 {
		params.SetInvalidParam("houseNumber", "Invalid house number.")
	}

	if err := googleMaps.VerifyAdressLink(a.GoogleMapsLink); err != nil {
		params.SetInvalidParam("GoogleMapsLink", err.Error())
	}

	return params
}

func (a *Adress) Prepare(userID uuid.UUID) error {
	params := a.Validate()

	if params != nil {
		return rfc7807.BadRequest("invalid-adress-data", "Invalid Adress Data Error", "Provided asress data is not valid.", params...)
	}

	a.User_ID = userID
	a.ID = uuid.New()
	return nil
}

func MigrateAdress(db *gorm.DB) error {
	return db.AutoMigrate(
		&Passenger{},
	)
}
