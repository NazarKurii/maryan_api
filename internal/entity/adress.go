package entity

import (
	googleMaps "maryan_api/internal/infrastructure/clients/google/maps"
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
)

type Address struct {
	ID              uuid.UUID      `gorm:"type:binary(16);primaryKey" json:"id,omitempty"`
	UserID          uuid.UUID      `gorm:"type:binary(16);not null"  json:"-"`
	CountryID       uuid.UUID      `gorm:"type:binary(16);not null"  json:"-"`
	Country         Country        `gorm:"foreignKey:CountryID;references:ID" json:"country"`
	City            string         `gorm:"type:varchar(56);not null" json:"city"`
	Street          string         `gorm:"type:varchar(255);not null" json:"street"`
	HouseNumber     string         `gorm:"type:varchar(15);not null" json:"houseNumber"`
	ApartmentNumber string         `gorm:"type:varchar(15)" json:"apartmentNumber"`
	GoogleMapsLink  string         `gorm:"type:varchar(255);not null" json:"googleMapsLink"`
	CreatedAt       time.Time      `gorm:"not null" json:"-"`
	DeletedAt       gorm.DeletedAt `json:"-"`
}

type Country struct {
	ID   uuid.UUID `gorm:"type:binary(16);primaryKey"                       `
	Name string    `gorm:"type:varchar(50);not null"`
}

type NewAddress struct {
	City            string `json:"city"`
	Street          string `json:"street"`
	HouseNumber     string `json:"houseNumber"`
	ApartmentNumber string `json:"apartmentNumber"`
}

func (address NewAddress) ToAddress(countryID uuid.UUID) Address {
	return Address{
		CountryID:       countryID,
		City:            address.City,
		Street:          address.Street,
		HouseNumber:     address.HouseNumber,
		ApartmentNumber: address.ApartmentNumber,
	}

}
func (a Address) Validate() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams

	if a.City == "" {
		params.SetInvalidParam("city", "Invalid city.")
	}
	if a.Street == "" {
		params.SetInvalidParam("street", "Invalid street.")
	}
	if a.HouseNumber == "" {
		params.SetInvalidParam("houseNumber", "Invalid house number.")
	}

	if err := googleMaps.VerifyAdressLink(a.GoogleMapsLink); err != nil {
		params.SetInvalidParam("GoogleMapsLink", err.Error())
	}

	return params
}

func (a *Address) Prepare(userID uuid.UUID) error {
	params := a.Validate()

	if params != nil {
		return rfc7807.BadRequest("invalid-Address-data", "Invalid Address Data Error", "Provided asress data is not valid.", params...)
	}

	a.UserID = userID
	a.ID = uuid.New()
	return nil
}

func MigrateAddress(db *gorm.DB) error {
	return db.AutoMigrate(
		&Country{},
		&Address{},
	)
}
