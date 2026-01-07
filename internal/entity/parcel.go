package entity

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/nazarkurii/marshrutka_api/config"
	rfc7807 "github.com/nazarkurii/marshrutka_api/pkg/problem"

	"github.com/asaskevich/govalidator"
	"github.com/d3code/uuid"
	"github.com/nyaruka/phonenumbers"
	"gorm.io/gorm"
)

type Parcel struct {
	ID                  uuid.UUID      `gorm:"type:binary(16);primaryKey"         json:"id"`
	UserID              uuid.UUID      `gorm:"type:binary(16);not null"           json:"userId"`
	ConnectionID        uuid.UUID      `gorm:"type:binary(16);not null"           json:"connectionID"`
	SenderPhoneNumber   string         `gorm:"type:varchar(15);not null"                                                  json:"senderPhoneNumber"`
	SenderEmail         string         `gorm:"type:varchar(255);not null"                          json:"senderEmail"`
	RecieverPhoneNumber string         `gorm:"type:varchar(15);not null"                                                  json:"recieverPhoneNumber"`
	RecieverEmail       string         `gorm:"type:varchar(255);not null"                          json:"recieverEmail"`
	SenderName          string         `gorm:"type:varchar(255);not null"                          json:"senderFirstName"`
	SenderLastName      string         `gorm:"type:varchar(255);not null"                          json:"senderLastName"`
	RecieverFirstName   string         `gorm:"type:varchar(255);not null"                          json:"recieverFirstName"`
	RecieverLastName    string         `gorm:"type:varchar(255);not null"                          json:"recieverLastName"`
	PickUpAdressID      uuid.UUID      `gorm:"type:binary(16);not null"           json:"-"`
	PickUpAdress        Address        `gorm:"foreignKey:PickUpAdressID"    json:"pickUpAddress"`
	DropOffAdressID     uuid.UUID      `gorm:"type:binary(16);not null"           json:"-"`
	DropOffAdress       Address        `gorm:"foreignKey:DropOffAdressID"   json:"dropOffAddress"`
	CreatedAt           time.Time      `gorm:"not null"                     json:"createdAt"`
	CompletedAt         sql.NullTime   `                                    json:"completedAt"`
	Payment             ParcelPayment  `gorm:"foreignKey:ParcelID"    `
	DeletedAt           gorm.DeletedAt `                                    json:"deletedAt"`
	LuggageVolume       uint           `gorm:"type:INT UNSIGNED;not null"`
	Width               int            `gorm:"type:SMALLINT UNSIGNED;not null"  json:"width"`
	Height              int            `gorm:"type:SMALLINT UNSIGNED;not null" json:"height"`
	Length              int            `gorm:"type:SMALLINT UNSIGNED;not null" json:"length"`
	Weight              int            `gorm:"type:SMALLINT UNSIGNED;not null" json:"weight"`
	Type                ParcelType     `gorm:"type:enum('Documents','Package'); not null" json:"type"`
	QRCode              []byte         `gorm:"type:blob;not null" json:"qrCode"`
}

type ParcelPayment struct {
	ParcelID  uuid.UUID     `gorm:"type:binary(16);not null"                                                json:"packadeId"`
	Price     int           `gorm:"type:MEDIUMINT;not null"                                           json:"price"`
	Method    paymentMethod `gorm:"type:enum('Apple Pay','Card','Cash','Google Pay');not null"        json:"method"`
	CreatedAt time.Time     `gorm:"not null"                                                          json:"createdAt"`
	SessionID string        `gorm:"type:varchar(500);not null"                                                          json:"sessionID"`
	Succeeded bool          `gorm:"not null"                                                          json:"succeeded"`
}

func MigratePackage(db *gorm.DB) error {
	return db.AutoMigrate(
		&Parcel{},
		&ParcelPayment{},
	)

}

type ParcelType string

const (
	PackageParcelType   ParcelType = "Package"
	DocumentsParcelType ParcelType = "Documents"
)

type FindParcelConnectionsRequest struct {
	From   string
	To     string
	Year   string
	Month  string
	Length string
	Height string
	Width  string
}

type CustomerParcel struct {
	Parcel     Parcel             `json:"parcel"`
	Connection CustomerConnection `json:"connection"`
}

type FindParcelConnectionsRequestParsed struct {
	From   uuid.UUID
	To     uuid.UUID
	Length int
	Height int
	Width  int
	Year   int
	Month  time.Month
}

func (fpcr FindParcelConnectionsRequest) Parse() (FindParcelConnectionsRequestParsed, error) {
	var params rfc7807.InvalidParams

	from, _, err := config.ParseCountry(fpcr.From)
	if err != nil {
		params.SetInvalidParam("from", err.Error())
	}

	to, _, err := config.ParseCountry(fpcr.To)
	if err != nil {
		params.SetInvalidParam("to", err.Error())
	}

	if params != nil {
		return FindParcelConnectionsRequestParsed{}, rfc7807.BadRequest("invalid-params", "Invalid Params Error", "Provided params are not valid.", params...)
	}

	year, err := strconv.Atoi(fpcr.Year)
	if err != nil {
		params.SetInvalidParam("year", err.Error())
	} else if year < config.MustParseToLocalByUUID(time.Now(), from).Year() {
		params.SetInvalidParam("year", fmt.Sprintf("Has to be current or future"))
	}

	monthNumber, err := strconv.Atoi(fpcr.Month)
	if err != nil {
		params.SetInvalidParam("month", err.Error())
	} else if monthNumber < 0 || monthNumber > 12 {
		params.SetInvalidParam("month", "Has to be between 1-12.")
	}

	height, err := strconv.Atoi(fpcr.Height)
	if err != nil {
		params.SetInvalidParam("height", err.Error())
	}
	width, err := strconv.Atoi(fpcr.Width)
	if err != nil {
		params.SetInvalidParam("width", err.Error())
	}

	length, err := strconv.Atoi(fpcr.Length)
	if err != nil {
		params.SetInvalidParam("length", err.Error())
	}

	return FindParcelConnectionsRequestParsed{
		From:   from,
		To:     to,
		Length: length,

		Height: height,
		Width:  width,
		Year:   year,
		Month:  time.Month(monthNumber),
	}, nil

}

type PurchaseParcelRequest struct {
	RecieverFirstName   string     `json:"recieverFirstName"`
	RecieverLastName    string     `json:"recieverLastName"`
	SenderFirstName     string     `json:"senderFirstName"`
	SenderLastName      string     `json:"senderLastName"`
	RecieverEmail       string     `json:"recieverEmail"`
	RecieverPhoneNumber string     `json:"recieverPhoneNumber"`
	SenderEmail         string     `json:"senderEmail"`
	SenderPhoneNumber   string     `json:"senderPhoneNumber"`
	DropOffAdress       NewAddress `json:"dropOffAdress"`
	PickUpAdress        NewAddress `json:"pickUpAdress"`
	Width               int        `json:"width"`
	Length              int        `json:"length"`
	Height              int        `json:"height"`
	Weight              int        `json:"weight"`
	Type                string     `json:"type"`
}
type ContactInfo struct {
	FirstName   string
	LastName    string
	Email       string
	PhoneNumber string
}

type PurchaseParcelRequestParsed struct {
	Recievier     ContactInfo
	Sender        ContactInfo
	DropOffAdress NewAddress
	PickUpAdress  NewAddress
	ConnectionID  uuid.UUID
	Width         int
	Length        int
	Height        int
	Weight        int
	Type          ParcelType
}

func (ppr PurchaseParcelRequest) Parse(connectionIdStr string) (PurchaseParcelRequestParsed, rfc7807.InvalidParams) {
	var params rfc7807.InvalidParams

	if len(ppr.RecieverFirstName) < 1 {
		params.SetInvalidParam("recieverFirstName", "Has not to be blank.")
	}

	if len(ppr.RecieverLastName) < 1 {
		params.SetInvalidParam("recieverLastName", "Has not to be blank.")
	}

	if len(ppr.SenderFirstName) < 1 {
		params.SetInvalidParam("senderFirstName", "Has not to be blank.")
	}

	if len(ppr.SenderLastName) < 1 {
		params.SetInvalidParam("senderLastName", "Has not to be blank.")
	}

	if !govalidator.IsEmail(ppr.RecieverEmail) {
		params.SetInvalidParam("recieverEmail", "Invalid email.")
	}

	if !govalidator.IsEmail(ppr.SenderEmail) {
		params.SetInvalidParam("senderEmail", "Invalid email.")
	}

	recieverPhoneNumber, err := phonenumbers.Parse(ppr.RecieverPhoneNumber, "UA")
	if err != nil {
		params.SetInvalidParam("recieverPhoneNumber", err.Error())
	}

	senderPhoneNumber, err := phonenumbers.Parse(ppr.SenderPhoneNumber, "UA")
	if err != nil {
		params.SetInvalidParam("senderPhoneNumber", err.Error())
	}

	connectionID, err := uuid.Parse(connectionIdStr)
	if err != nil {
		params.SetInvalidParam("connectionID", err.Error())
	}

	if ppr.Weight > 50000 || ppr.Weight < 1000 {
		params.SetInvalidParam("weight", "Has to be less than or equal to 50.")
	}

	parcelType, ok := defineParcelType(ppr.Type)
	if !ok {
		params.SetInvalidParam("type", "Invalid parcell type.")
	}

	if params != nil {
		return PurchaseParcelRequestParsed{}, params
	}

	return PurchaseParcelRequestParsed{

		Recievier: ContactInfo{
			FirstName:   ppr.RecieverFirstName,
			LastName:    ppr.RecieverLastName,
			Email:       ppr.RecieverEmail,
			PhoneNumber: phonenumbers.Format(recieverPhoneNumber, phonenumbers.E164),
		},

		Sender: ContactInfo{
			FirstName:   ppr.SenderFirstName,
			LastName:    ppr.SenderLastName,
			Email:       ppr.RecieverEmail,
			PhoneNumber: phonenumbers.Format(senderPhoneNumber, phonenumbers.E164),
		},
		DropOffAdress: ppr.DropOffAdress,
		PickUpAdress:  ppr.PickUpAdress,
		ConnectionID:  connectionID,
		Width:         ppr.Width,
		Length:        ppr.Length,
		Height:        ppr.Height,
		Weight:        ppr.Weight,
		Type:          parcelType,
	}, nil
}

func defineParcelType(parcelType string) (ParcelType, bool) {
	switch parcelType {
	case "package":
		return PackageParcelType, true
	case "documents":
		return DocumentsParcelType, true
	default:
		return ParcelType(""), false
	}
}
