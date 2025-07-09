package entity

import (
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/biter777/countries"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Connection struct {
	ID                      uuid.UUID          `gorm:"type:uuid;primaryKey"                                                      json:"id"`
	Line                    int                `gorm:"type:TINYINT;not null"                                                     json:"line"`
	DepartureCountry        string             `gorm:"type:varchar(56);not null"                                                 json:"departureCountry"`
	DestinationCountry      string             `gorm:"type:varchar(56);not null"                                                 json:"destinationCountry"`
	DepartureTime           time.Time          `gorm:"not null"                                                                  json:"departureTime"`
	ArrivalTime             time.Time          `gorm:"not null"                                                                  json:"arrivalTime"`
	EstimatedDuration       int                `gorm:"-"                                                                         json:"estimatedDuration"`
	GoogleMapsConnectionURL string             `gorm:"not null"                                                                  json:"googleMapsConnectionURL"`
	BusID                   uuid.UUID          `gorm:"type:uuid;not null"                                                        json:"-"`
	Bus                     Bus                `gorm:"foreignKey:BusID"                                                          json:"bus"`
	Stops                   []Stop             `                                                                                 json:"stops"`
	CreatedAt               time.Time          `gorm:"not null"                                                                  json:"createdAt"`
	Updates                 []ConnectionUpdate `gorm:"not null"                                                                  json:"updates"`
	Type                    connectionType     `gorm:"type:enum('Comertial','Special Asignment','Break Down Retun')"             json:"type"`
}

type connectionType string

func ParseConectionType(v string) (connectionType, bool) {
	switch v {
	case ComertialConnectionType, SpecialAsignmentConnectionType, BreakDownRetunConnectionType:
		return connectionType(v), true
	default:
		return "", false
	}
}

type connectionStatus string
type ConnectionUpdate struct {
	ConnectionID uuid.UUID        `json:"-"          gorm:"type:uuid; not null"`
	CreatedAt    time.Time        `json:"createAt"   gorm:"not null" json:"createAt"`
	Status       connectionStatus `json:"status"     gorm:"type:enum('Registered','Canceled','Sold','Started','Finished','Stopped','Renewed','Could Not Be Finished');not null" `
	Comment      string           `json:"commnet"    gorm:"type:varchar(500)"`
}

const (
	ComertialConnectionType        = "Comertial"
	SpecialAsignmentConnectionType = "Special Asignment"
	BreakDownRetunConnectionType   = "Break down"

	RegisteredConnectionStatus       = "Registered"
	CanceledConnectionStatus         = "Canceled"
	SoldConnectionStatus             = "Sold"
	StartedConnectionStatus          = "Started"
	FinishedConnectionStatus         = "Finished"
	StoppedConnectionStatus          = "Stopped"
	RenewedConnectionStatus          = "Renewed"
	CouldNotBeFinishConnectionStatus = "Could Not Be Finished"
)

type Stop struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey"                         json:"id"`
	TicketID uuid.UUID  `gorm:"not null"                                     json:"-"`
	Ticket   Ticket     `gorm:"foreignKey:TicketID"                          json:"ticket"`
	RouteID  uuid.UUID  `gorm:"not null"                                     json:"-"`
	Type     stopType   `gorm:"type:enum('Pick-up','Drop-off')"              json:"type"`
	Status   stopStatus `gorm:"type:enum('Confirmed','Missed','Completed')"  json:"status"`
}

type stopType string
type stopStatus string

const (
	PickUpStopType  stopType = "Pick-up"
	DropOffStopType stopType = "Drop-off"

	ConfirmedStopStatus stopStatus = "Confirmed"
	MissedStopStatus    stopStatus = "Missed"
	CompletedStopStatus stopStatus = "Completed"
)

func MigrateConnection(db *gorm.DB) error {
	return db.AutoMigrate(
		&Connection{},
		&ConnectionUpdate{},
		&Stop{},
	)

}

func (c *Connection) Validate() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams

	if countries.ByName(c.DepartureCountry) == 0 {
		params.SetInvalidParam("departureCoutry", "Non-existing-country.")
	}

	if countries.ByName(c.DestinationCountry) == 0 {
		params.SetInvalidParam("destinationCountry", "Non-existing-country.")
	}

	if c.DepartureTime.Before(time.Now()) {
		params.SetInvalidParam("DepartureTime", "Past time.")
	}

	return params
}

func (c *Connection) PrepareNew() {
	c.ID = uuid.New()
	c.Updates = []ConnectionUpdate{{
		ConnectionID: c.ID,
		Status:       RegisteredConnectionStatus,
	}}
}

//
//
//
//
//
//
//
//
//

type ConnectionSimplified struct {
	ID                 uuid.UUID `json:"id"`
	DepartureCountry   string    `json:"departureCountry"`
	DestinationCountry string    `json:"destinationCountry"`
	DepartureTime      time.Time `json:"departureTime"`
	ArrivalTime        time.Time `json:"arrivalTime"`
	EstimatedDuration  int       `json:"estimatedDuration"`
}

func (c *Connection) Simplify() ConnectionSimplified {
	return ConnectionSimplified{
		ID:                 c.ID,
		DepartureCountry:   c.DepartureCountry,
		DestinationCountry: c.DestinationCountry,
		DepartureTime:      c.DepartureTime,
		ArrivalTime:        c.ArrivalTime,
		EstimatedDuration:  c.EstimatedDuration,
	}
}

type CustomerConnection struct {
	ConnectionSimplified
	GoogleMapsConnectionURL string `json:"googleMapsConnectionURL"`
	Bus                     Bus    `json:"bus"`
	Stops                   []Stop `json:"stops"`
}

func (c *Connection) ToCustomer() CustomerConnection {
	return CustomerConnection{
		ConnectionSimplified:    c.Simplify(),
		GoogleMapsConnectionURL: c.GoogleMapsConnectionURL,
		Bus:                     c.Bus,
		Stops:                   c.Stops,
	}
}
