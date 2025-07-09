package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	TripStatusChangedDriver     = "Changed Driver"
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
	Status    tripStatus `json:"status"     gorm:"type:enum('Registered','Canceled','Changed Driver','Changed Bus','Started','Outbound Done','Break Down','Broken Bus Fixed','Broken Bus Replaced','Finished');not null"`
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
