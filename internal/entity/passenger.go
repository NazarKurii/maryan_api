package entity

import (
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Passenger struct {
	ID          uuid.UUID      `gorm:"type:uuid; primaryKey;"         json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;"                     json:"-"`
	Name        string         `gorm:"type:varchar(255); not null"    json:"name"`
	Surname     string         `gorm:"type:varchar(255); not null"    json:"surname"`
	DateOfBirth time.Time      `gorm:"not null"                       json:"dateOfBirth"`
	CreatedAt   time.Time      `gorm:"not null"                       json:"-"`
	DeletedAt   gorm.DeletedAt `                                      json:"-"`
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

	if p.Name == "" {
		params.SetInvalidParam("name", "Must not be empty.")
	}

	if p.Surname == "" {
		params.SetInvalidParam("surname", "Must not be empty.")
	}

	if p.DateOfBirth.IsZero() || p.DateOfBirth.After(time.Now()) {
		params.SetInvalidParam("dateOfBirth", "Invalid date of birth.")
	}

	return params
}

func MigratePassenger(db *gorm.DB) error {
	return db.AutoMigrate(
		&Passenger{},
	)
}
