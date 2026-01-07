package entity

import (
	"context"
	"net/http"
	"time"

	googleMaps "github.com/nazarkurii/marshrutka_api/internal/infrastructure/clients/google/maps"
	rfc7807 "github.com/nazarkurii/marshrutka_api/pkg/problem"

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
	GoogleMapsID    string         `gorm:"type:varchar(255);not null" json:"googleMapsID"`
	FormatedAdress  string         `gorm:"type:varchar(500);not null" json:"formatedAdress"`
	CreatedAt       time.Time      `gorm:"not null" json:"-"`
	DeletedAt       gorm.DeletedAt `json:"-"`
}

type Country struct {
	ID   uuid.UUID `gorm:"type:binary(16);primaryKey"                       `
	Name string    `gorm:"type:varchar(50);not null; UNIQUE"`
}

type NewAddress struct {
	City            string `json:"city"`
	Street          string `json:"street"`
	HouseNumber     string `json:"houseNumber"`
	ApartmentNumber string `json:"apartmentNumber"`
	GoogleMapsID    string `json:"googleMapsID"`
	FormatedAdress  string ` json:"formatedAdress"`
}

func (address NewAddress) ToAddress(countryID uuid.UUID) Address {
	return Address{
		CountryID:       countryID,
		City:            address.City,
		Street:          address.Street,
		HouseNumber:     address.HouseNumber,
		ApartmentNumber: address.ApartmentNumber,
		GoogleMapsID:    address.GoogleMapsID,
		FormatedAdress:  address.FormatedAdress,
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

	return params
}

func (a *Address) Prepare(context context.Context, client *http.Client) error {
	params := a.Validate()

	if params != nil {
		return rfc7807.BadRequest("invalid-Address-data", "Invalid Address Data Error", "Provided asress data is not valid.", params...)
	}

	err := googleMaps.VerifyAdressID(context, client, a.GoogleMapsID)
	if err != nil {
		return err
	}

	a.ID = uuid.New()
	return nil
}

func MigrateAddress(db *gorm.DB) error {
	return db.AutoMigrate(
		&Country{},
		&Address{},
	)
}
