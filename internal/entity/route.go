package entity

import (
	"time"
)

type Route struct {
	ID                 uint      `gorm:"primaryKey; autoincrement"`
	CreatedAt          time.Time `gorm:"not null"`
	UpdatedAt          time.Time `gorm:"not null"`
	DepartureCountry   string    `gorm:"type:varchar(56);not null"`
	DestinationCountry string    `gorm:"type:varchar(56);not null"`
	DepartureTime      time.Time `gorm:"not null"`
	ArrivalTime        time.Time `gorm:"not null"`
	EstimatedDuration  uint16    `gorm:"-"`
	BusID              uint      `gorm:"not null"`
	Bus                Bus       `gorm:"foreignKey:BusID"`
	Stops              []Stop    `gorm:"foreignKey:BusLineID"`
}

type Stop struct {
	BusLineID uint   `gorm:"not null"`
	AdressID  uint   `gorm:"not null"`
	Adress    Adress `gorm:"foreignKey:AdressID"`
}
