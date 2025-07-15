package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Trip struct {
	ID                   uuid.UUID    `gorm:"type:uuid;primaryKey"                         json:"id"`
	OutboundConnectionID uuid.UUID    `gorm:"type:uuid;not null"                           json:"-"`
	OutboundConnection   Connection   `gorm:"foreignKey:OutboundTripID;references:ID"      json:"outboundConnection"`
	ReturnConnectionID   uuid.UUID    `gorm:"type:uuid;not null"                           json:"-"`
	ReturnConnection     Connection   `gorm:"foreignKey:ReturnTripID;references:ID"        json:"returnConnection"`
	Updates              []TripUpdate `                                                    json:"updates"`
}

type tripStatus string

const (
	TripStatusRegistered        = "Registered"
	TripStatusCanceled          = "Canceled"
	TripStatusChangedBus        = "Changed Bus"
	TripStatusStarted           = "Started"
	TripStatusOutboundDone      = "Outbound Done"
	TripStatusBreakDown         = "Break Down"
	TripStatusBrokenBusFixed    = "Broken Bus Fixed"
	TripStatusBrokenBusReplaced = "Broken Bus Replaced"
	TripStatusFinished          = "Finished"
)

type TripUpdate struct {
	TripID    uuid.UUID  `json:"-"          gorm:"type:uuid;not null"`
	Status    tripStatus `json:"status"     gorm:"type:enum('Registered','Canceled','Changed Bus','Started','Outbound Done','Break Down','Broken Bus Fixed','Broken Bus Replaced','Finished');not null"`
	CreatedAt time.Time  `json:"createdAt"  gorm:"not null"`
	Comment   string     `json:"comment"    gorm:"type:varchar(500)"`
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
