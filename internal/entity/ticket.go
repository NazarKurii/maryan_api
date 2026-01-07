package entity

import (
	"context"
	"database/sql"
	"net/http"
	"slices"

	"time"

	rfc7807 "github.com/nazarkurii/marshrutka_api/pkg/problem"

	"github.com/nazarkurii/marshrutka_api/config"

	"github.com/asaskevich/govalidator"
	"github.com/d3code/uuid"
	"gorm.io/gorm"
)

type Ticket struct {
	ID              uuid.UUID      `gorm:"type:binary(16);primaryKey"         json:"id"`
	UserID          uuid.UUID      `gorm:"type:binary(16);not null"           json:"userId"`
	ConnectionID    uuid.UUID      `gorm:"type:binary(16);not null"           json:"connectionID"`
	Seats           []TicketSeat   `gorm:"constraint:OnDelete:CASCADE" json:"seats"`
	PhoneNumber     string         `gorm:"type:varchar(15);not null"                                                  json:"phoneNumber"`
	Email           string         `gorm:"type:varchar(255);not null"                          json:"email"`
	Passengers      []Passenger    `gorm:"foreignKey:TicketID;onstraint:OnDelete:CASCADE"      json:"passengers"`
	PickUpAdressID  uuid.UUID      `gorm:"type:binary(16);not null"           json:"-"`
	PickUpAdress    Address        `gorm:"foreignKey:PickUpAdressID;onstraint:OnDelete:CASCADE"    json:"pickUpAddress"`
	DropOffAdressID uuid.UUID      `gorm:"type:binary(16);not null"           json:"-"`
	DropOffAdress   Address        `gorm:"foreignKey:DropOffAdressID;onstraint:OnDelete:CASCADE"   json:"dropOffAddress"`
	CreatedAt       time.Time      `gorm:"not null"                     json:"createdAt"`
	CompletedAt     sql.NullTime   `                                    json:"completedAt"`
	Payment         TicketPayment  `gorm:"foreignKey:TicketID;onstraint:OnDelete:CASCADE"    `
	DeletedAt       gorm.DeletedAt `                                    json:"deletedAt"`
	LuggageVolume   luggage        `gorm:"type:MEDIUMINT UNSIGNED;not null"`
	QRCode          []byte         `gorm:"type:blob;not null" json:"qrCode"`
}

type luggage uint

const (
	Backpack     luggage = 24000
	SmallLuggage luggage = 60000
	LargeLuggage luggage = 200000
)

type TicketSeat struct {
	TicketID uuid.UUID `gorm:"type:binary(16);not null;constraint:OnDelete:CASCADE"          json:"-"`
	SeatID   uuid.UUID `gorm:"type:binary(16);not null"           json:"-"`
	Seat     Seat      `gorm:"foreignKey:SeatID"   json:"seat"`
}
type CustomerTicket struct {
	Ticket     Ticket               `json:"ticket"`
	Connection ConnectionSimplified `json:"connection"`
	Expired    bool                 `json:"expired"`
}

type TicketPayment struct {
	TicketID  uuid.UUID     `gorm:"type:binary(16);not null;primaryKey;constraint:OnDelete:CASCADE"                                              json:"ticketId"`
	Price     int           `gorm:"type:MEDIUMINT;not null"                                           json:"price"`
	Method    paymentMethod `gorm:"type:enum('Apple Pay','Card','Cash','Google Pay');not null"        json:"method"`
	CreatedAt time.Time     `gorm:"not null"                                                          json:"createdAt"`
	SessionID string        `gorm:"type:varchar(500);not null"                                                          json:"sessionID"`
	Succeeded bool          `gorm:"not null"                                                          json:"succeeded"`
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
		&TicketSeat{},
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
	Backpacks     int            `json:"backpacks"`
	SmallLuggage  int            `json:"smallLuggage"`
	LargeLuggage  int            `json:"largeLuggage"`
}

func (t NewTicketJSON) LuggagePrice() int {

	var backPacksTotalPrice, largeLuggageTotalPrice int

	luggage := config.GetLoggageConfig()
	passengersNumber := len(t.Passengers)

	if t.Backpacks > passengersNumber {
		backPacksTotalPrice = (t.Backpacks - passengersNumber) * int(luggage.Small.Price)
	}

	if t.LargeLuggage > passengersNumber {
		largeLuggageTotalPrice = (t.LargeLuggage - passengersNumber) * int(luggage.Large.Price)
	}

	return backPacksTotalPrice + largeLuggageTotalPrice + t.SmallLuggage*int(luggage.Medium.Price)

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

func (t NewTicketJSON) Validate(connection Connection, takenSeats []uuid.UUID, ticketID uuid.UUID, luggageVolumeLeft uint) ([]TicketSeat, error) {

	if connection.DepartureTime.Before(time.Now().UTC()) {
		return nil, rfc7807.BadRequest("unavailable-connection0", "Unavailavble Connection Error", "Connection has alredy departed.")
	}

	seatsLength := len(t.SeatIDs)

	if seatsLength != len(t.Passengers) {
		return nil, rfc7807.BadRequest("seats-passengers", "Seats Passengers Error", "The seats number and the passengers number have to be equal.")
	}

	var totalSeats int
	for _, seat := range connection.Bus.Seats {
		if seat.Number != 0 {
			totalSeats++
		}
	}

	if totalSeats-len(takenSeats) < seatsLength {
		return nil, rfc7807.New(http.StatusConflict, "available-seats", "Available Seats Error", "There are no anough available seats for this many passengers")
	}

	if t.LuggageVolume() > luggage(luggageVolumeLeft) {
		return nil, rfc7807.New(http.StatusConflict, "luggage-volume", "Luggage Volume Error", "There is not enough space left to fit provided luggage volume.")
	}

	var seats = make([]TicketSeat, seatsLength)
	for i, seat := range t.SeatIDs {
		if slices.Contains(takenSeats, seat) {
			return nil, rfc7807.New(http.StatusConflict, "taken-seat", "Taken Seat Error", seat.String()+" is already taken.")
		} else {
			seats[i] = TicketSeat{
				TicketID: ticketID,
				SeatID:   seat,
			}
		}
	}
	return seats, nil
}

func (t NewTicketJSON) ParseAdresses(ctx context.Context, client *http.Client, departureCountryID, destinationCountryID uuid.UUID) (*Address, *Address, error) {
	pickUpAdress := t.PickUpAdress.ToAddress(departureCountryID)
	err := pickUpAdress.Prepare(ctx, client)
	if err != nil {
		return nil, nil, err
	}

	dropOffAddress := t.DropOffAdress.ToAddress(destinationCountryID)
	err = dropOffAddress.Prepare(ctx, client)
	if err != nil {
		return nil, nil, err
	}
	return &pickUpAdress, &dropOffAddress, nil

}

func (t NewTicketJSON) ParsePassengers(ticketID uuid.UUID) ([]Passenger, error) {
	var passengers []Passenger
	for _, newPassenger := range t.Passengers {
		passenger := newPassenger.Parse()
		params := passenger.Prepare(ticketID)
		if params != nil {
			return nil, rfc7807.BadRequest("passenger-invalid-data", "Passenger Data Error", "Provided data is not valid.", params...)
		}

		passengers = append(passengers, passenger)
	}
	return passengers, nil
}

func (t NewTicketJSON) LuggageVolume() luggage {
	return luggage(t.SmallLuggage)*(SmallLuggage) + luggage(t.Backpacks)*(Backpack) + luggage(t.LargeLuggage)*(LargeLuggage)
}
