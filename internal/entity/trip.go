package entity

import (
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/google/uuid"
)

type Trip struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	DepartureCountry   string    `gorm:"type:varchar(56);not null"`
	DestinationCountry string    `gorm:"type:varchar(56);not null"`
	DepartureTime      time.Time `gorm:"not null"`
	ArrivalTime        time.Time `gorm:"not null"`
	EstimatedDuration  uint16    `gorm:"-"`
	BusID              uuid.UUID `gorm:"not null"`
	Bus                Bus       `gorm:"foreignKey:BusID"`
	Driver             User      `gorm:"foreignKey:DriverID"`
	DriverID           uuid.UUID `gorm:"not null"`
	Stops              []Stop
	GoogleMapsTripURL  string    `gorm:"not null"`
	CreatedAt          time.Time `gorm:"not null"`
	UpdatedAt          time.Time `gorm:"not null"`
	CanceledAt         time.Time
	FinishedAt         time.Time
	StartedAt          time.Time
}

type Stop struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TicketID    uuid.UUID  `gorm:"not null"`
	Ticket      Adress     `gorm:"foreignKey:AdressID"`
	RouteID     uuid.UUID  `gorm:"not null"`
	PassengerID uuid.UUID  `gorm:"not null"`
	Passenger   Passenger  `gorm:"foreignKey:AdressID"`
	Adress      Adress     `gorm:"foreignKey:AdressID"`
	AdressID    uuid.UUID  `gorm:"not null"`
	Type        stopType   `gorm:"type:enum('Pick-up','Drop-off')"`
	Status      stopStatus `gorm:"type:enum('Confirmed','Missed','Completed')"`
}

type Ticket struct {
	ID              uint                `gorm:"primaryKey; autoincrement"`
	BusLineID       uint                `gorm:" not null"`
	BusLine         busline.BusLine     `gorm:"foreignKey:BusLineID"`
	PassengerID     uint                `gorm:" not null"`
	Passenger       passenger.Passenger `gorm:"foreignKey:PassengerID"`
	PickUpAdressID  uint                `gorm:" not null"`
	PickUpAdress    adress.Adress       `gorm:"foreignKey:PickUpAdressID"`
	DropOffAdressID uint                `gorm:" not null"`
	DropOffAdress   adress.Adress       `gorm:"foreignKey:DropOffAdressID"`
}

type stopType string

const (
	PickUpStopType  stopType = "Pick-up"
	DropOffStopType stopType = "Drop-off"
)

type stopStatus string

const (
	ConfirmedStopStatus stopStatus = "Confirmed"
	MissedStopStatus    stopStatus = "Missed"
	CompletedStopStatus stopStatus = "Completed"
)

func (t Trip) Validate() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams

	if t.DepartureCountry == "" {
		params.SetInvalidParam("departureCountry", "No data has been provided.")
	}
	if t.DepartureCountry == "" {
		params.SetInvalidParam("destinationCountry", "No data has been provided.")
	}
	if t.DepartureTime.Before(time.Now()) {
		params.SetInvalidParam("departureTime", "Provided departure time is not valid (past time provided).")
	}
	if t.BusID == uuid.Nil {
		params.SetInvalidParam("busId", "No data has been provided.")
	}

	if t.DriverID == uuid.Nil {
		params.SetInvalidParam("busId", "No data has been provided.")
	}

	return params
}

func (t *Trip) NewID() {
	t.ID = uuid.New()
}
