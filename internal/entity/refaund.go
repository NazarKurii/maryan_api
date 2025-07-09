package entity

import (
	"time"

	"github.com/google/uuid"
)

type Refaund struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TicketID    uuid.UUID `gorm:"type:uuid;not null"   json:"-"`
	Ticket      Ticket    `gorm:"foreignKey:TicketID"  json:"ticket"`
	CreatedAt   time.Time `gorm:"not null"             json:"createdAt"`
	CompletedAt time.Time `                            json:"completedAt"`
}

func NewRefaund(ticketID uuid.UUID) Refaund {
	return Refaund{
		ID:       uuid.New(),
		TicketID: ticketID,
	}
}
