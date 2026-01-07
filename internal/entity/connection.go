package entity

import (
	"strconv"
	"time"

	"github.com/nazarkurii/marshrutka_api/config"
	rfc7807 "github.com/nazarkurii/marshrutka_api/pkg/problem"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Connection struct {
	ID   uuid.UUID `gorm:"type:binary(16);primaryKey" json:"-"`
	Line int       `gorm:"type:SMALLINT;not null" json:"line"`

	Price int `gorm:"type:MEDIUMINT UNSIGNED;not null" json:"price"`

	DepartureCountryID uuid.UUID `gorm:"type:binary(16);not null" json:"-"`
	DepartureCountry   Country   `gorm:"foreignKey:DepartureCountryID;references:ID" json:"departureCountry"`

	DestinationCountryID uuid.UUID `gorm:"type:binary(16);not null" json:"-"`
	DestinationCountry   Country   `gorm:"foreignKey:DestinationCountryID;references:ID" json:"destinationCountry"`

	DepartureTime     time.Time `gorm:"not null" json:"departureTime"`
	ArrivalTime       time.Time `gorm:"not null" json:"arrivalTime"`
	EstimatedDuration int       `gorm:"-" json:"estimatedDuration"`
	GoogleMapsURL     string    `gorm:"not null" json:"googleMapsConnectionURL"`

	BusID uuid.UUID `gorm:"type:binary(16);not null" json:"-"`
	Bus   Bus       `gorm:"foreignKey:BusID" json:"bus"`

	Stops     []Stop             `json:"stops"`
	CreatedAt time.Time          `gorm:"not null" json:"createdAt"`
	Updates   []ConnectionUpdate `gorm:"not null" json:"updates"`

	Type connectionType `gorm:"type:enum('Comertial','Special Asignment','Break Down Return', 'Break Down Replacement'); not null" json:"type"`

	SellBefore        time.Time `gorm:"not null" json:"sellBefore"`
	BackpackPrice     int       `gorm:"type:MEDIUMINT UNSIGNED;not null" json:"backpackPrice"`
	SmallLuggagePrice int       `gorm:"type:MEDIUMINT UNSIGNED;not null" json:"smallLuggagePrice"`
	LargeLuggagePrice int       `gorm:"type:MEDIUMINT UNSIGNED; not null" json:"largeLuggagePrice"`

	MaxWidth          int  `gorm:"type:SMALLINT UNSIGNED;not null"`
	MaxHeight         int  `gorm:"type:SMALLINT UNSIGNED;not null"`
	MaxLength         int  `gorm:"type:SMALLINT UNSIGNED;not null"`
	LuggageVolumeLeft uint `gorm:"-"`
}

func (c *Connection) AfterFind(tx *gorm.DB) (err error) {

	c.EstimatedDuration = int(c.ArrivalTime.Sub(c.DepartureTime).Minutes())

	return
}

type connectionType string
type ConnectionType struct {
	Val connectionType
}

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
	ConnectionID uuid.UUID        `json:"-"          gorm:"type:binary(16); not null"`
	CreatedAt    time.Time        `json:"createAt"   gorm:"not null" json:"createAt"`
	Status       connectionStatus `json:"status"     gorm:"type:enum('Registered','Canceled','Sold','Started','Finished','Stopped','Renewed','Could Not Be Finished','Departure Time Changed');not null" `
	Comment      string           `json:"commnet"    gorm:"type:varchar(500)"`
}

func (tu ConnectionUpdate) Validate() error {
	switch tu.Status {
	case RegisteredConnectionStatus,
		ChangedDepartureTimeConnectionStatus,
		CanceledConnectionStatus,
		SoldConnectionStatus,
		StartedConnectionStatus,
		FinishedConnectionStatus,
		StoppedConnectionStatus,
		RenewedConnectionStatus,
		CouldNotBeFinishConnectionStatus:
	default:
		return rfc7807.BadRequest("invalid-connection-status", "Invalid Connection Status Error", "Connection status provided is not valid.")
	}

	return nil
}

const (
	ComertialConnectionType            = "Comertial"
	SpecialAsignmentConnectionType     = "Special Asignment"
	BreakDownRetunConnectionType       = "Break down Return"
	BreakDownReplacementConnectionType = "Break Down Replacement"

	RegisteredConnectionStatus           = "Registered"
	ChangedDepartureTimeConnectionStatus = "Departure Time Changed"
	CanceledConnectionStatus             = "Canceled"
	SoldConnectionStatus                 = "Sold"
	StartedConnectionStatus              = "Started"
	FinishedConnectionStatus             = "Finished"
	StoppedConnectionStatus              = "Stopped"
	RenewedConnectionStatus              = "Renewed"
	CouldNotBeFinishConnectionStatus     = "Could Not Be Finished"
)

type Stop struct {
	ID           uuid.UUID        `gorm:"type:binary(16);primaryKey"                         json:"id"`
	TicketID     uuid.NullUUID    `gorm:"type:binary(16)"                                     json:"-"`
	Ticket       Ticket           `gorm:"foreignKey:TicketID"                          json:"ticket"`
	ParcelID     uuid.NullUUID    `gorm:"type:binary(16)"                                     json:"-"`
	Parcel       Parcel           `gorm:"foreignKey:ParcelID"                          json:"package"`
	ConnectionID uuid.UUID        `gorm:"type:binary(16);not null"                                     json:"-"`
	Type         stopType         `gorm:"type:enum('Passenger','Parcel')"              json:"type"`
	LocationType stopLocationType `gorm:"type:enum('Pick-up','Drop-off')"              json:"locationType"`
	Updates      []StopUpdate     `gorm:"constraint:OnDelete:CASCADE"                                                 json:"updates"`
}

type stopType string
type StopUpdate struct {
	StopID    uuid.UUID  `gorm:"type:binary(16); not null"                             json:"-" `
	Status    stopStatus `gorm:"type:enum('Confirmed','Missed','Completed')"     json:"status"`
	Comment   string     `gorm:"type:varchar(500)"                               json:"comment"`
	CreatedAt time.Time  `gorm:"not null"                                        json:"createdAt"`
}
type stopStatus string
type stopLocationType string

const (
	PassengerStopType stopType = "Passenger"
	ParcelStopType    stopType = "Parcel"

	PickUpStopType  stopLocationType = "Pick-up"
	DropOffStopType stopLocationType = "Drop-off"

	ConfirmedStopStatus stopStatus = "Confirmed"
	MissedStopStatus    stopStatus = "Missed"
	CompletedStopStatus stopStatus = "Completed"
)

func MigrateConnection(db *gorm.DB) error {
	return db.AutoMigrate(
		&Connection{},
		&ConnectionUpdate{},
		&Stop{},
		&StopUpdate{},
	)

}

func (c *Connection) Validate() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams

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
//cd
//

func (c *Connection) Simplify() ConnectionSimplified {

	return ConnectionSimplified{
		ID:                 c.ID,
		Price:              c.Price,
		DepartureCountry:   c.DepartureCountry.Name,
		DestinationCountry: c.DestinationCountry.Name,
		DepartureTime:      config.MustParseToLocalByUUID(c.DepartureTime, c.DepartureCountryID),
		ArrivalTime:        config.MustParseToLocalByUUID(c.ArrivalTime, c.DestinationCountryID),
		Line:               c.Line,
		EstimatedDuration:  c.EstimatedDuration,
		SellBefore:         c.SellBefore,
	}
}

type CustomerConnection struct {
	ConnectionSimplified
	GoogleMapsConnectionURL string      `json:"googleMapsConnectionURL"`
	Bus                     CustomerBus `json:"bus"`
	Stops                   []Stop      `json:"stops"`
	LuggageVolumeLeft       uint        `json:"luggageVolumeLeft"`
	BackpackPrice           int         `json:"backpackPrice"`
	SmallLuggagePrice       int         `json:"smallLuggagePrice"`
	LargeLuggagePrice       int         `json:"largeLuggagePrice"`
}

type ParcelConnection struct {
	ConnectionSimplified
	Price int `json:"price"`
}

func (c *Connection) ToCustomer(takenSeatsIDs []uuid.UUID, smallLuggagePrice, mediumLuggagePrice, largeLuggagePrice int) CustomerConnection {
	return CustomerConnection{
		ConnectionSimplified:    c.Simplify(),
		GoogleMapsConnectionURL: c.GoogleMapsURL,
		Bus:                     c.Bus.ToCustomerBus(takenSeatsIDs),
		Stops:                   c.Stops,
		LuggageVolumeLeft:       c.LuggageVolumeLeft,
		BackpackPrice:           smallLuggagePrice,
		SmallLuggagePrice:       mediumLuggagePrice,
		LargeLuggagePrice:       largeLuggagePrice,
	}
}

func PreloadConnection() []string {
	return []string{
		clause.Associations,

		"Bus.Images",
		"Bus.LeadDriver",
		"Bus.AssistantDriver",
		"Bus.Seats",
		"Bus.Structure",
		"Bus.Structure.Positions",
	}
}

type FindConnectionsRequestJSON struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Date      string `json:"date"`
	Adults    string `json:"adults"`
	Children  string `json:"children"`
	Teenagers string `json:"teenagers"`
}

func (r FindConnectionsRequestJSON) Parse() (FindConnectionsRequest, rfc7807.InvalidParams) {
	var invalidParams rfc7807.InvalidParams

	adults, err := strconv.Atoi(r.Adults)
	if err != nil {
		invalidParams.SetInvalidParam("adults", err.Error())
	} else if adults < 0 {
		invalidParams.SetInvalidParam("adults", "cannot be less that 0.")
	}
	teenagers, err := strconv.Atoi(r.Teenagers)
	if err != nil {
		invalidParams.SetInvalidParam("infants", err.Error())
	} else if teenagers < 0 {
		invalidParams.SetInvalidParam("infants", "cannot be less that 0.")
	} else if teenagers < 1 && adults < 1 {
		invalidParams.SetInvalidParam("passengers", "There has to be at least one adult or one teenager.")
	}

	children, err := strconv.Atoi(r.Children)
	if err != nil {
		invalidParams.SetInvalidParam("children", err.Error())
	} else if children < 0 {
		invalidParams.SetInvalidParam("children", "cannot be less that 0.")
	} else if children > 0 && adults < 1 {
		invalidParams.SetInvalidParam("children", "cannot be more than 0 if there is no adult.")
	}

	if invalidParams != nil {
		return FindConnectionsRequest{}, invalidParams
	}

	fromID, timeLocation, err := config.ParseCountry(r.From)
	if err != nil {
		invalidParams.SetInvalidParam("from", err.Error())
	}

	toID, _, err := config.ParseCountry(r.To)
	if err != nil {
		invalidParams.SetInvalidParam("from", err.Error())
	}

	date, err := time.ParseInLocation("2006-01-02", r.Date, timeLocation)
	if err != nil {
		invalidParams.SetInvalidParam("date", err.Error())
	}

	return FindConnectionsRequest{
		From:      fromID,
		To:        toID,
		Date:      date,
		Adults:    adults,
		Children:  children,
		Teenagers: teenagers,
	}, nil

}

type FindConnectionsRequest struct {
	From      uuid.UUID
	To        uuid.UUID
	Date      time.Time
	Adults    int
	Children  int
	Teenagers int
}

type FindConnectionsResponse struct {
	Connections []FoundConnection  `json:"connections"`
	LeftRange   []ConnectionsRange `json:"leftRange"`
	RightRange  []ConnectionsRange `json:"rightRange"`
}

type FoundConnection struct {
	ConnectionSimplified
	TicketsLeft int  `json:"ticketsLeft"`
	Fits        bool `json:"fits"`
	Available   bool `json:"available"`
}
type ConnectionSimplified struct {
	ID                 uuid.UUID `json:"id"`
	Price              int       `json:"price"`
	Line               int       `json:"line"`
	DepartureCountry   string    `json:"departureCountry"`
	DestinationCountry string    `json:"destinationCountry"`
	DepartureTime      time.Time `json:"departureTime"`
	ArrivalTime        time.Time `json:"arrivalTime"`
	EstimatedDuration  int       `json:"estimatedDuration"`
	SellBefore         time.Time `json:"sellBefore"`
}

type ConnectionParcel struct {
	ConnectionSimplified
	Price             int  `json:"price"`
	LuggageVolumeLeft uint `json:"luggageVolumeLeft"`
	MaxWidth          uint `json:"maxWidth"`
	MaxHeight         uint `json:"maxHeight"`
	MaxLength         uint `json:"maxLength"`
	Usable            bool `json:"usable"`
	DayNumber         int  `json:"dayNumber"`
	DayMonth          int  `json:"dayMonth"`
	IsCurrentMonth    bool `json:"isCurrentMonth"`
}
type ConnectionsRange struct {
	Date       time.Time `gorm:"column:date" json:"date"`
	Available  bool      `json:"available"`
	SellBefore time.Time `gorm:"column:sellBefore" json:"-"`
	Number     int       `gorm:"column:number" json:"number"`
	MinPrice   int       `gorm:"column:minPrice" json:"minPrice"`
}

func (c *Connection) ToParcelConnection(usable bool, dayNumber, dayMonth int, isCurrentMonth bool, price int) ConnectionParcel {

	return ConnectionParcel{
		Price:                price,
		ConnectionSimplified: c.Simplify(),
		LuggageVolumeLeft:    c.LuggageVolumeLeft,
		MaxWidth:             c.Bus.MaxWidth,
		MaxHeight:            c.Bus.MaxHeight,
		MaxLength:            c.Bus.MaxLength,
		Usable:               usable,
		DayNumber:            dayNumber,
		DayMonth:             dayMonth,
		IsCurrentMonth:       isCurrentMonth,
	}
}
