package entity

import (
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/google/uuid"
)

type Passenger struct {
	ID          uuid.UUID `gorm:"type:uuid; primaryKey;" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;" json:"-"`
	Name        string    `gorm:"type:varchar(255); not null" json:"name"`
	Surname     string    `gorm:"type:varchar(255); not null" json:"surname"`
	DateOfBirth time.Time `gorm:"not null" json:"dateOfBirth"`
	Adress      Adress
}

type Adress struct {
	ID              uuid.UUID `gorm:"primaryKey;autoincrement"`
	PassengerID     uuid.UUID `gorm:"type:uuid; not null" json:"-"`
	Country         string    `gorm:"type:varchar(56);not null" json:"country"`
	City            string    `gorm:"type:varchar(56);not null" json:"city"`
	Street          string    `gorm:"type:varchar(255);not null" json:"street"`
	HouseNumber     int       `gorm:"type:smallint; not null" json:"houseNumber"`
	ApartmentNumber int       `gorm:"type:smallint" json:"apartmentNumber"`
	GoogleMapsLink  string    `gorm:"type:varchar(255);not null"`
	CreatedAt       time.Time `gorm:"not null"`
	DeletedAt       time.Time `gorm:"index"`
}

func (p *Passenger) prepare() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams

	if p.Name == "" {
		params.SetInvalidParam("name", "Must not be empty.")
	}

	if p.Surname == "" {
		params.SetInvalidParam("surname", "Must not be empty.")
	}

	if p.DateOfBirth.IsZero() || p.DateOfBirth.After(time.Now()) {
		params.SetInvalidParam("dateOfBirth", "Invalid date of birth.")
	}

	if (p.Adress == Adress{}) {
		params.SetInvalidParam("address", "Missing address.")
	} else {
		if p.Adress.Country == "" {
			params.SetInvalidParam("address.country", "Invalid country.")
		}
		if p.Adress.City == "" {
			params.SetInvalidParam("address.city", "Invalid city.")
		}
		if p.Adress.Street == "" {
			params.SetInvalidParam("address.street", "Invalid street.")
		}
		if p.Adress.HouseNumber <= 0 {
			params.SetInvalidParam("address.houseNumber", "Invalid house number.")
		}
	}

	// Assign IDs
	p.ID = uuid.New()
	p.Adress.ID = uuid.New()
	p.Adress.PassengerID = p.ID

	return params
}
