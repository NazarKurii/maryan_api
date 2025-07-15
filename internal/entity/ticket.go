package entity

import (
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ticket struct {
	ID              uuid.UUID      `gorm:"type:binary(16);primaryKey"         json:"id"`
	UserID          uuid.UUID      `gorm:"type:binary(16);not null"           json:"userId"`
	ConnectionID    uuid.UUID      `gorm:"type:binary(16);not null"           json:"connectionID"`
	Seat            Seat           `json:"seat"`
	SeatID          uuid.UUID      `gorm:"type:binary(16);not null"           json:"-"`
	PhoneNumber     string         `gorm:"type:varchar(15);not null"                                                  json:"phoneNumber"`
	Email           string         `gorm:"type:varchar(255);not null"                          json:"email"`
	PassengerID     uuid.UUID      `gorm:"type:binary(16);not null"           json:"-"`
	Passenger       Passenger      `gorm:"foreignKey:PassengerID"       json:"passenger"`
	PickUpAdressID  uuid.UUID      `gorm:"type:binary(16);not null"           json:"-"`
	PickUpAdress    Address        `gorm:"foreignKey:PickUpAdressID"    json:"pickUpAddress"`
	DropOffAdressID uuid.UUID      `gorm:"type:binary(16);not null"           json:"-"`
	DropOffAdress   Address        `gorm:"foreignKey:DropOffAdressID"   json:"dropOffAddress"`
	CreatedAt       time.Time      `gorm:"not null"                     json:"createdAt"`
	CompletedAt     time.Time      `                                    json:"completedAt"`
	TicketPayment   TicketPayment  `gorm:"foreignKey:TicketID"    `
	DeletedAt       gorm.DeletedAt `                                    json:"deletedAt"`
}

type CustomerTicket struct {
	Ticket     Ticket               `json:"ticket"`
	Connection ConnectionSimplified `json:"connection"`
}

type TicketPayment struct {
	TicketID  uuid.UUID     `gorm:"type:binary(16);not null"                                                json:"ticketId"`
	Price     int           `gorm:"type:MEDIUMINT;not null"                                           json:"price"`
	Method    paymentMethod `gorm:"type:enum('Apple Pay','Card','Cash','Google Pay');not null"        json:"method"`
	CreatedAt time.Time     `gorm:"not null"                                                          json:"createdAt"`
	SessionID string        `gorm:"type:varchar(500);not null"                                                          json:"sessionID"`
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
		&TicketPayment{},
	)

}

type NewTicketJSON struct {
	ConnectionID  uuid.UUID      `json:"connectionId"`
	SeatIDs       []uuid.UUID    `json:"seatIDs"`
	Passengers    []NewPassenger `json:"passengers"`
	DropOffAdress NewAddress     `json:"dropOffAdress"`
	PickUpAdress  NewAddress     `json:"pickUpAdress"`
	Email         string         `json:"email"`
	PhoneNumber   string         `json:"phoneNumber"`
}

func (t NewTicketJSON) ParseContaanctInfo() (email string, phoneNumber string, err error) {
	var params rfc7807.InvalidParams
	if !govalidator.IsEmail(t.Email) {
		params.SetInvalidParam("email", "Contains invalid characters or is not an email.")
	} else {
		email = t.Email
	}

	phoneNumber, err = fomratPhoneNumber(t.PhoneNumber)
	if err != nil {
		params.SetInvalidParam("phoneNumber", err.Error())
	}

	err = rfc7807.BadRequest("invalid-data", "Invalid Data Error", "The provided data is not valid.", params...)

	return
}
