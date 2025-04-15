package entity

import (
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
)

type Passenger struct {
	ID        uuid.UUID      `gorm:"type:binary(16); primaryKey;"         json:"id"`
	UserID    uuid.UUID      `gorm:"type:binary(16);"                     json:"-"`
	FirstName string         `gorm:"type:varchar(255); not null"    json:"firstName"`
	LastName  string         `gorm:"type:varchar(255); not null"    json:"lastName"`
	CreatedAt time.Time      `gorm:"not null"                       json:"-"`
	DeletedAt gorm.DeletedAt `                                      json:"-"`
}

type PassengerSimplified struct {
	ID        uuid.UUID `json:"binary(16)"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
}

type NewPassenger struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (p NewPassenger) Parse() Passenger {
	return Passenger{
		FirstName: p.FirstName,
		LastName:  p.LastName,
	}
}

func (p *Passenger) Prepare(userID uuid.UUID) rfc7807.InvalidParams {
	params := p.Validate()

	if params == nil {
		p.ID = uuid.New()
		p.UserID = userID
	}

	return params
}

func (p *Passenger) Validate() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams

	if p.FirstName == "" {
		params.SetInvalidParam("name", "Must not be empty.")
	}

	if p.LastName == "" {
		params.SetInvalidParam("surname", "Must not be empty.")
	}

	return params
}

func MigratePassenger(db *gorm.DB) error {
	return db.AutoMigrate(
		&Passenger{},
	)
}
