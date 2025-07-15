package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ticket struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey"         json:"id"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null"           json:"userId"`
	ConnectionID    uuid.UUID      `gorm:"type:uuid;not null"           json:"connectionID"`
	PassengerID     uuid.UUID      `gorm:"type:uuid;not null"           json:"-"`
	Passenger       Passenger      `gorm:"foreignKey:PassengerID"       json:"passenger"`
	PickUpAdressID  uuid.UUID      `gorm:"type:uuid;not null"           json:"-"`
	PickUpAdress    Adress         `gorm:"foreignKey:PickUpAdressID"    json:"pickUpAddress"`
	DropOffAdressID uuid.UUID      `gorm:"type:uuid;not null"           json:"-"`
	DropOffAdress   Adress         `gorm:"foreignKey:DropOffAdressID"   json:"dropOffAddress"`
	Payment         Payment        `                                    json:"payment"`
	CreatedAt       time.Time      `gorm:"not null"                     json:"createdAt"`
	CompletedAt     time.Time      `                                    json:"completedAt"`
	DeletedAt       gorm.DeletedAt `                                    json:"deletedAt"`
}

type Payment struct {
	TicketID  uuid.UUID     `gorm:"type:uuid;not null"                                                json:"ticketId"`
	Price     int           `gorm:"type:MEDIUMINT;not null"                                           json:"price"`
	Method    paymentMethod `gorm:"type:enum('Apple Pay','Card','Cash','Google Pay');not null"        json:"method"`
	CreatedAt time.Time     `gorm:"not null"                                                          json:"createdAt"`
}

type paymentMethod string

const (
	PaymentMethodApplePay  = "Apple Pay"
	PaymentMethodCard      = "Card"
	PaymentMethodCash      = "Cash"
	PaymentMethodGooglePay = "Google Pay"
)

func DefinePaymentMethod(v string) (paymentMethod, bool) {
	switch paymentMethod(v) {
	case PaymentMethodApplePay, PaymentMethodCard, PaymentMethodCash, PaymentMethodGooglePay:
		return paymentMethod(v), true
	default:
		return "", false
	}
}

func MigrateTicket(db *gorm.DB) error {
	return db.AutoMigrate(
		&Ticket{},
		&Payment{},
	)

}
