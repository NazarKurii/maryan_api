package entity

import (
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Trip struct {
	ID                   uuid.UUID    `gorm:"type:binary(16);primaryKey"                         json:"id"`
	OutboundConnectionID uuid.UUID    `gorm:"type:binary(16);not null"                           json:"-"`
	OutboundConnection   Connection   `gorm:"foreignKey:OutboundConnectionID;references:ID"      json:"outboundConnection"`
	ReturnConnectionID   uuid.UUID    `gorm:"type:binary(16);not null"                           json:"-"`
	ReturnConnection     Connection   `gorm:"foreignKey:ReturnConnectionID;references:ID"        json:"returnConnection"`
	Updates              []TripUpdate `                                                    json:"updates"`
}

type tripStatus string

const (
	TripStatusRegistered        tripStatus = "Registered"
	TripStatusCanceled          tripStatus = "Canceled"
	TripStatusChangedBus        tripStatus = "Changed Bus"
	TripStatusStarted           tripStatus = "Started"
	TripStatusOutboundDone      tripStatus = "Outbound Done"
	TripStatusBreakDown         tripStatus = "Break Down"
	TripStatusBrokenBusFixed    tripStatus = "Broken Bus Fixed"
	TripStatusBrokenBusReplaced tripStatus = "Broken Bus Replaced"
	TripStatusFinished          tripStatus = "Finished"
)

type TripUpdate struct {
	TripID    uuid.UUID  `json:"-"          gorm:"type:binary(16);not null"`
	Status    tripStatus `json:"status"     gorm:"type:enum('Registered','Canceled','Changed Bus','Started','Outbound Done','Break Down','Broken Bus Fixed','Broken Bus Replaced','Finished');not null"`
	CreatedAt time.Time  `json:"createdAt"  gorm:"not null"`
	Comment   string     `json:"comment"    gorm:"type:varchar(500)"`
}

func (tu *TripUpdate) Validate() error {

	switch tu.Status {
	case TripStatusRegistered,
		TripStatusCanceled,
		TripStatusChangedBus,
		TripStatusStarted,
		TripStatusOutboundDone,
		TripStatusBreakDown,
		TripStatusBrokenBusFixed,
		TripStatusBrokenBusReplaced,
		TripStatusFinished:
	default:
		return rfc7807.BadRequest("invalid-trip-status", "Invalid Trip Status Error", "Trip status provided is not valid.")
	}

	return nil
}

type TripSimplified struct {
	ID                 uuid.UUID            `json:"id"`
	OutboundConnection ConnectionSimplified `json:"outboundConnection"`
	ReturnConnection   ConnectionSimplified `json:"returnConnection"`
}

func (t Trip) Simplify() TripSimplified {
	return TripSimplified{
		ID:                 t.ID,
		OutboundConnection: t.OutboundConnection.Simplify(),
		ReturnConnection:   t.ReturnConnection.Simplify(),
	}
}

func MigrateTrip(db *gorm.DB) error {
	return db.AutoMigrate(
		&Trip{},
		&TripUpdate{},
	)
}

func PreloadTrip() []string {
	return []string{
		clause.Associations,

		"OutboundConnection.Bus",
		"OutboundConnection.Bus.Images",
		"OutboundConnection.Bus.LeadDriver",
		"OutboundConnection.Bus.AssistantUser",
		"OutboundConnection.Bus.Seats",
		"OutboundConnection.Bus.Structure",
		"OutboundConnection.Bus.Structure.Positions",

		"OutboundConnection.ReplacedBus",
		"OutboundConnection.ReplacedBus.Images",
		"OutboundConnection.ReplacedBus.LeadDriver",
		"OutboundConnection.ReplacedBus.AssistantUser",
		"OutboundConnection.ReplacedBus.Seats",
		"OutboundConnection.ReplacedBus.Structure",
		"OutboundConnection.ReplacedBus.Structure.Positions",

		"OutboundConnection.Stops",
		"OutboundConnection.Stops.Ticket",
		"OutboundConnection.Stops.Updates",
		"OutboundConnection.Updates",

		"ReturnConnection.Bus",
		"ReturnConnection.Bus.Images",
		"ReturnConnection.Bus.LeadDriver",
		"ReturnConnection.Bus.AssistantUser",
		"ReturnConnection.Bus.Seats",
		"ReturnConnection.Bus.Structure",
		"ReturnConnection.Bus.Structure.Positions",

		"ReturnConnection.ReplacedBus",
		"ReturnConnection.ReplacedBus.Images",
		"ReturnConnection.ReplacedBus.LeadDriver",
		"ReturnConnection.ReplacedBus.AssistantUser",
		"ReturnConnection.ReplacedBus.Seats",
		"ReturnConnection.ReplacedBus.Structure",
		"ReturnConnection.ReplacedBus.Structure.Positions",

		"ReturnConnection.Stops",
		"ReturnConnection.Stops.Ticket",
		"ReturnConnection.Stops.Updates",
		"ReturnConnection.Updates",
	}
}

func (t Trip) Validate() rfc7807.InvalidParams {
	params := t.OutboundConnection.Validate()
	params = append(params, t.ReturnConnection.Validate()...)

	if t.OutboundConnection.BusID != t.ReturnConnection.BusID {
		params.SetInvalidParam("busIDs", "Connections have to posses the same bus id.")
	}

	if t.OutboundConnection.DestinationCountry != t.ReturnConnection.DepartureCountry {
		params.SetInvalidParam("connections", "The destination country of the outbound connection has be the same as the departure coutry of return connection.")
	}

	if t.OutboundConnection.DepartureCountry.Name != "Ukraine" {
		params.SetInvalidParam("outbondConnection", "Departure coutry has to be 'Ukraine'.")
	}

	if t.ReturnConnection.DestinationCountry.Name != "Ukraine" {
		params.SetInvalidParam("returnConnection", "Destination coutry has to be 'Ukraine'.")
	}

	diff := t.ReturnConnection.ArrivalTime.Sub(t.OutboundConnection.DepartureTime)
	if diff < 0 {
		params.SetInvalidParam("dates", "Return arrival time cannot be before Outbound departure time")
	}

	if diff.Hours() < 48 {
		params.SetInvalidParam("dates", "Difference between return arrival time and  Outbound departure time cannot be les than 30 hours.")
	}

	return params
}

func (t *Trip) PreapareNew() {
	t.OutboundConnection.PrepareNew()
	t.ReturnConnection.PrepareNew()
	t.ID = uuid.New()

	t.Updates = []TripUpdate{{
		TripID: t.ID,
		Status: TripStatusRegistered,
	}}

}
