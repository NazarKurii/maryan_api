package entity

import (
	"encoding/json"
	"fmt"
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/biter777/countries"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Trip struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	DepartureCountry   string    `gorm:"type:varchar(56);not null"`
	DestinationCountry string    `gorm:"type:varchar(56);not null"`
	DepartureTime      `gorm:"not null"`
	ArrivalTime        `gorm:"not null"`
	EstimatedDuration  string    `gorm:"-"`
	GoogleMapsTripURL  string    `gorm:"not null"`
	BusID              uuid.UUID `gorm:"not null"`
	Bus                Bus       `gorm:"foreignKey:BusID"`
	Driver             User      `gorm:"foreignKey:DriverID"`
	DriverID           uuid.UUID `gorm:"not null"`
	Stops              []Stop
	CreatedAt          time.Time `gorm:"not null"`
	UpdatedAt          time.Time `gorm:"not null"`
	CanceledAt         time.Time `gorm:"index"`
	SoldAt             time.Time
	FinishedAt         time.Time
	StartedAt          time.Time
}

type DepartureTime struct {
	DepartureTime time.Time
}
type ArrivalTime struct {
	ArrivalTime time.Time
}
type timeJSON struct {
	year      int
	month     int
	monthName string
	day       int
	hour      int
	minute    int
}

func (d DepartureTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(timeJSON{
		year:      d.DepartureTime.Year(),
		month:     int(d.DepartureTime.Month()),
		monthName: d.DepartureTime.Month().String(),
		day:       d.DepartureTime.Day(),
		hour:      d.DepartureTime.Hour(),
		minute:    d.DepartureTime.Minute(),
	})
}

func (d *DepartureTime) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return fmt.Errorf("invalid departure time format: %w", err)
	}
	d.DepartureTime, err = time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		return fmt.Errorf("cannot parse departure time: %w", err)
	}
	return nil
}

func (d ArrivalTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(timeJSON{
		year:      d.ArrivalTime.Year(),
		month:     int(d.ArrivalTime.Month()),
		monthName: d.ArrivalTime.Month().String(),
		day:       d.ArrivalTime.Day(),
		hour:      d.ArrivalTime.Hour(),
		minute:    d.ArrivalTime.Minute(),
	})
}

func (d *ArrivalTime) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return fmt.Errorf("invalid departure time format: %w", err)
	}
	d.ArrivalTime, err = time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		return fmt.Errorf("cannot parse departure time: %w", err)
	}
	return nil
}

type Ticket struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID `gorm:"type:uuid;not null"`
	TripID          uuid.UUID `gorm:"type:uuid;not null"`
	PassengerID     uuid.UUID `gorm:"type:uuid;not null"`
	Passenger       Passenger `gorm:"foreignKey: passengerID"`
	PickUpAdressID  uuid.UUID `gorm:"type:uuid;not null"`
	PickUpAdress    Adress    `gorm:"foreignKey: PickUpAdressID"`
	DropOffAdressID uuid.UUID `gorm:"type:uuid;not null"`
	DropOffAdress   Adress    `gorm:"foreignKey: DropOffAdressID"`
	Payment         Payment
	CreatedAt       time.Time `gorm:"not null"`
	CanceledAt      time.Time
	CompletedAt     time.Time
}

type Refaund struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TicketID    uuid.UUID `gorm:"type:uuid;not null"`
	Ticket      Ticket    `gorm:"foreignKey:TicketID"`
	CreatedAt   time.Time `gorm:"not null"`
	CompletedAt time.Time
}

type Payment struct {
	TicketID  uuid.UUID     `gorm:"type:uuid;not null"`
	Price     int           `gorm:"type:MEDIUMINT;not null"`
	Method    paymentMethod `gorm:"type:enum('Apple Pay', 'Card', 'Cash', 'Google Pay');not null"`
	CreatedAt time.Time     `gorm:"not null"`
}

type paymentMethod string

const (
	ApplePayPaymentMethod  = "Apple Pay"
	CardPaymentMethod      = "Card"
	CashPaymentMethod      = "Cash"
	GooglePayPaymentMethod = "Google Pay"
)

type Stop struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TicketID uuid.UUID  `gorm:"not null"`
	Ticket   Ticket     `gorm:"foreignKey:TicketID"`
	RouteID  uuid.UUID  `gorm:"not null"`
	Type     stopType   `gorm:"type:enum('Pick-up','Drop-off')"`
	Status   stopStatus `gorm:"type:enum('Confirmed','Missed','Completed')"`
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

func MigrateTrip(db *gorm.DB) error {
	return db.AutoMigrate(
		&Trip{},
		&Ticket{},
		&Payment{},
		&Stop{},
		&Refaund{},
	)

}

func (t *Trip) Validate() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams

	if countries.ByName(t.DepartureCountry) == 0 {
		params.SetInvalidParam("departureCoutry", "Non-existing-country.")
	}

	if countries.ByName(t.DestinationCountry) == 0 {
		params.SetInvalidParam("destinationCountry", "Non-existing-country.")
	}

	if t.DepartureTime.DepartureTime.Before(time.Now()) {
		params.SetInvalidParam("DepartureTime", "Past time.")
	}

	return params
}
