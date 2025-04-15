package entity

import (
	"encoding/json"
	"fmt"
	rfc7807 "maryan_api/pkg/problem"
	"slices"
	"time"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
)

type Bus struct {
	ID                 uuid.UUID      `gorm:"type:binary(16);primaryKey"                       `
	Model              string         `gorm:"type:varchar(255);not null"                 `
	Images             []BusImage     `gorm:"foreignKey:BusID"                                    `
	RegistrationNumber string         `gorm:"type:varchar(8);not null;unique"            `
	Year               int            `gorm:"type:smallint;not null"                     `
	GpsTrackerID       string         `gorm:"type:varchar(255);not null"                 `
	LeadDriver         User           `gorm:"foreignKey:LeadDriverID;references:ID"      `
	LeadDriverID       uuid.NullUUID  `gorm:"type:binary(16);unique"                  `
	AssistantDriver    User           `gorm:"foreignKey:AssistantDriverID;references:ID" `
	AssistantDriverID  uuid.NullUUID  `gorm:"type:binary(16);unique"                  `
	Seats              []Seat         `gorm:"foreignKey:BusID"                           `
	Structure          []Row          `gorm:"foreignKey:BusID"                                   `
	CreatedAt          time.Time      `gorm:"not null"                                   `
	UpdatedAt          time.Time      `gorm:"not null"                                   `
	DeletedAt          gorm.DeletedAt `gorm:"index"                                      `
}

//

type Seat struct {
	ID        uuid.UUID     `gorm:"type:binary(16);primaryKey;"                                                       `
	BusID     uuid.UUID     `gorm:"type:binary(16)"                                                                   `
	Number    int           `gorm:"type:tinyint;not null"                                                       `
	Type      seatType      `gorm:"type:enum('Window', 'Single', 'Single-Window', 'Aisle', 'Middle');not null"  `
	Direction seatDirection `gorm:"type:enum('Forward', 'Backward');not null"                                   `
}

type seatType string

const (
	SeatTypeSingle       seatType = "Single"
	SeatTypeSingleWindow seatType = "Single-Window"
	SeatTypeWindow       seatType = "Window"
	SeatTypeAisle        seatType = "Aisle"
	SeatTypeMidlle       seatType = "Middle"
	seatTypeError        seatType = "Error"
)

func defineSeatType(typeStr string) (seatType, bool) {
	switch seatType(typeStr) {
	case SeatTypeSingle:
		return SeatTypeSingle, true
	case SeatTypeSingleWindow:
		return SeatTypeSingleWindow, true
	case SeatTypeWindow:
		return SeatTypeWindow, true
	case SeatTypeAisle:
		return SeatTypeAisle, true
	case SeatTypeMidlle:
		return SeatTypeMidlle, true
	default:
		return seatTypeError, false
	}
}

type seatDirection string

const (
	SeatDirectionForward  seatDirection = "Forward"
	SeatDirectionBackward seatDirection = "Backward"
	seatDirectionError    seatDirection = "Error"
)

func defineSeatDirection(directionStr string) (seatDirection, bool) {
	switch seatDirection(directionStr) {
	case SeatDirectionForward:
		return SeatDirectionForward, true
	case SeatDirectionBackward:
		return SeatDirectionBackward, true
	default:
		return seatDirectionError, false

	}
}

type Row struct {
	ID        uuid.UUID      `gorm:"type:binary(16);primaryKey"    `
	BusID     uuid.UUID      `gorm:"type:binary(16), not null"     `
	Number    int            `gorm:"type:TINYINT; not null"  `
	Positions []SeatPosition `gorm:"foreignKey:RowID"        `
}

type SeatPosition struct {
	RowID      uuid.UUID         `gorm:"type:binary(16), not null"                          `
	SeatNumber int               `gorm:"type:TINYINT; not null"                       `
	Type       seatPossitionType `gorm:"type:enum('Space', 'Table', 'Seat');not null" `
	Position   int               `gorm:"type:TINYINT; not null"                       `
}

type seatPossitionType string

const (
	SeatPossitionTypeSpace seatPossitionType = "Space"
	SeatPossitionTypeTable seatPossitionType = "Table"
	SeatPossitionTypeSeat  seatPossitionType = "Seat"
	seatPossitionTypeError seatPossitionType = "Error"
)

func defineSeatPositionType(typeStr string) (seatPossitionType, bool) {
	switch seatPossitionType(typeStr) {
	case SeatPossitionTypeSeat:
		return SeatPossitionTypeSeat, true
	case SeatPossitionTypeTable:
		return SeatPossitionTypeTable, true
	case SeatPossitionTypeSpace:
		return SeatPossitionTypeSpace, true
	default:
		return seatPossitionTypeError, false
	}
}

type BusImage struct {
	BusID uuid.UUID `gorm:"type:binary(16);not null"            `
	Url   string    `gorm:"type:varchar(255);not null"    `
}

func (b BusImage) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Url)
}

type BusAvailability struct {
	BusID   uuid.UUID             `gorm:"type:binary(16); not null"                                                  json:"-"`
	Status  busAvailabilityStatus `gorm:"type:enum('Other','Broken','Busy'); not null"                         json:"status"`
	Date    time.Time             `gorm:"not null;default:CURRENT_TIME_STAMP"                                  json:"date"`
	Comment string                `gorm:"type:varchar(500)"                                                    json:"comment;omitempty"`
}

type busAvailabilityStatus string

const (
	BusAvailabilityStatusBroken busAvailabilityStatus = "Broken"
	BusAvailabilityStatusOther  busAvailabilityStatus = "Other"
	BusAvailabilityStatusBusy   busAvailabilityStatus = "Busy"
)

func (ba busAvailabilityStatus) IsValid() bool {
	switch ba {
	case BusAvailabilityStatusBroken, BusAvailabilityStatusOther, BusAvailabilityStatusBusy:
		return true
	default:
		return false
	}
}

func (b *Bus) Prepare() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams
	if b.Model == "" {
		params.SetInvalidParam("model", "Invalid bus model.")
	}

	if b.Year < 1990 {
		params.SetInvalidParam("year", "Invalid production year.")
	}

	if b.RegistrationNumber == "" {
		params.SetInvalidParam("registrationNumber", "Invalid bus registration number.")
	}

	if len(b.Seats) == 0 {
		params.SetInvalidParam("seats", "Invalid bus seats number, must be greater than 0.")
	}

	if len(b.Images) != 0 {
		b.Images = nil
	}

	b.ID = uuid.New()
	var seatNumbers = map[int]int{}

	for i, seat := range b.Seats {

		if seat.Number < 1 {
			params.SetInvalidParam(fmt.Sprintf("seat (index:%d)", i), "Invalid seat number.")
		}

		if seatNumbers[seat.Number] == 1 {
			params.SetInvalidParam(fmt.Sprintf("seat (index:%d)", i), "Invalid seat number (Repeated).")
		} else {
			seatNumbers[seat.Number]++
		}

		if params == nil {
			b.Seats[i].BusID = b.ID
			b.Seats[i].ID = uuid.New()
		}

	}

	for i, row := range b.Structure {
		if params == nil {
			b.Structure[i].ID = uuid.New()
		}
		for j, seat := range row.Positions {
			switch {
			case seat.SeatNumber == 0 && seat.Type == SeatPossitionTypeSeat:
				params.SetInvalidParam(fmt.Sprintf("Structure row(index:%d) seatPosition(index:%d)", i, j), "Seat is not empty, but seat number is 0")
			case seat.SeatNumber != 0 && (seat.Type == SeatPossitionTypeSpace || seat.Type == SeatPossitionTypeTable):
				params.SetInvalidParam(fmt.Sprintf("Structure row(index:%d) seatPosition(index:%d)", i, j), "Seat is  empty, but seat number is not 0")
			default:
				if seat.Type == SeatPossitionTypeSeat {
					if seatNumbers[seat.SeatNumber] == 0 {
						params.SetInvalidParam(fmt.Sprintf("Structure row(index:%d) seatPosition(index:%d)", i, j), "Seat number is not unique")
					} else {
						seatNumbers[seat.SeatNumber]--
					}
				}

				if params == nil {
					b.Structure[i].Positions[j].RowID = b.Structure[i].ID
				}
			}
		}

		b.Structure[i].BusID = b.ID

	}

	return params
}

func MigrateBus(db *gorm.DB) error {
	return db.AutoMigrate(
		&Bus{},
		&Seat{},
		&Row{},
		&SeatPosition{},
		&BusImage{},
	)
}

type CustomerBus struct {
	Model              string                   `json:"model"`
	Images             []string                 `json:"imageURLs"`
	RegistrationNumber string                   `json:"registrationNumber"`
	Year               int                      `json:"year"`
	Structure          [][]ResponseCustomerSeat `json:"structure"`
}

type ResponseCustomerSeat struct {
	ResponseSeat
	Taken bool `json:"taken"`
}

type ResponseSeat struct {
	ID        uuid.UUID `json:"id"`
	Number    int       `json:"number"`
	Type      string    `json:"type"`
	Direction string    `json:"direction"`
}

func (b Bus) ToCustomerBus(takenSeatsIDs []uuid.UUID) CustomerBus {
	var imageUrls = make([]string, len(b.Images))
	for i, image := range b.Images {
		imageUrls[i] = image.Url
	}

	return CustomerBus{
		Model:              b.Model,
		Images:             imageUrls,
		RegistrationNumber: b.RegistrationNumber,
		Year:               b.Year,
		Structure:          b.responseCustomerStructure(takenSeatsIDs),
	}
}

func (b Bus) responseCustomerStructure(takenSeatsIDs []uuid.UUID) [][]ResponseCustomerSeat {
	var structure = make([][]ResponseCustomerSeat, len(b.Structure))

	for _, row := range b.Structure {
		structure[row.Number] = make([]ResponseCustomerSeat, len(row.Positions))
		for _, rowObject := range row.Positions {

			if rowObject.Type == SeatPossitionTypeTable || rowObject.Type == SeatPossitionTypeSpace {
				structure[row.Number][rowObject.Position] = ResponseCustomerSeat{
					ResponseSeat: ResponseSeat{Type: string(rowObject.Type)},
				}
			} else {
				seatIndex := slices.IndexFunc(b.Seats, func(seat Seat) bool {
					return seat.Number == rowObject.SeatNumber
				})

				structure[row.Number][rowObject.Position] = ResponseCustomerSeat{
					ResponseSeat: ResponseSeat{Type: string(b.Seats[seatIndex].Type),
						Number:    b.Seats[seatIndex].Number,
						ID:        b.Seats[seatIndex].ID,
						Direction: string(b.Seats[seatIndex].Direction)},
					Taken: slices.ContainsFunc(takenSeatsIDs, func(id uuid.UUID) bool { return id == b.Seats[seatIndex].ID }),
				}
			}

		}
	}
	return structure
}

func (b Bus) responseStructure() [][]ResponseSeat {
	var structure = make([][]ResponseSeat, len(b.Structure))

	for _, row := range b.Structure {
		structure[row.Number] = make([]ResponseSeat, len(row.Positions))
		for _, rowObject := range row.Positions {

			if rowObject.Type == SeatPossitionTypeTable || rowObject.Type == SeatPossitionTypeSpace {
				structure[row.Number][rowObject.Position] = ResponseSeat{
					Type: string(rowObject.Type),
				}
			} else {
				seatIndex := slices.IndexFunc(b.Seats, func(seat Seat) bool {
					return seat.Number == rowObject.SeatNumber
				})

				structure[row.Number][rowObject.Position] = ResponseSeat{
					Type:      string(b.Seats[seatIndex].Type),
					Number:    b.Seats[seatIndex].Number,
					ID:        b.Seats[seatIndex].ID,
					Direction: string(b.Seats[seatIndex].Direction),
				}
			}

		}
	}
	return structure
}

type EmployeeBus struct {
	ID                 uuid.UUID        `json:"id"`
	Model              string           `json:"model"`
	ImageUrls          []string         `json:"imageURLs"`
	RegistrationNumber string           `json:"registrationNumber"`
	Year               int              `json:"year"`
	GpsTrackerID       string           `json:"gpsTrackerID"`
	LeadDriver         User             `json:"leadDriver"`
	AssistantDriver    User             `json:"assistantDriver"`
	Structure          [][]ResponseSeat `json:"structure"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
	DeletedAt          gorm.DeletedAt   `json:"deletedAt"`
}

func (b Bus) ToEmployeeBus() EmployeeBus {
	var imageUrls = make([]string, len(b.Images))
	for i, image := range b.Images {
		imageUrls[i] = image.Url
	}

	return EmployeeBus{
		ID:                 b.ID,
		Model:              b.Model,
		ImageUrls:          imageUrls,
		RegistrationNumber: b.RegistrationNumber,
		Year:               b.Year,
		Structure:          b.responseStructure(),
		LeadDriver:         b.LeadDriver,
		AssistantDriver:    b.AssistantDriver,
		CreatedAt:          b.CreatedAt,
		UpdatedAt:          b.UpdatedAt,
		DeletedAt:          b.DeletedAt,
	}
}

type NewBus struct {
	Model              string        `gorm:"type:varchar(255);not null"                   json:"model"`
	RegistrationNumber string        `gorm:"type:varchar(8);not null;unique"              json:"registrationNumber"`
	Year               int           `gorm:"type:smallint;not null"                       json:"year"`
	GpsTrackerID       string        `gorm:"type:varchar(255);not null"                   json:"gpsTrackerID"`
	LeadDriverID       uuid.NullUUID `gorm:"type:uuid;not null"                           json:"leadDriverID"`
	AssistantDriverID  uuid.NullUUID `gorm:"type:uuid;not null"                           json:"assistantDriverID"`
	Structure          [][]NewSeat   `gorm:"not null"                                     json:"structure"`
}

type NewSeat struct {
	Number    int    `json:"number"`
	Type      string `json:"type"`
	Direction string `json:"direction"`
}

func (nb NewBus) Parse() (Bus, rfc7807.InvalidParams) {
	var bus = Bus{
		Model:              nb.Model,
		RegistrationNumber: nb.RegistrationNumber,
		Year:               nb.Year,
		GpsTrackerID:       nb.GpsTrackerID,
		LeadDriverID:       nb.LeadDriverID,
		AssistantDriverID:  nb.AssistantDriverID,
		Structure:          make([]Row, len(nb.Structure)),
	}

	var InvalidParams rfc7807.InvalidParams

	for rowIndex, newRow := range nb.Structure {
		bus.Structure[rowIndex].Number = rowIndex
		for seatIndex, newSeat := range newRow {
			seatPositionType, ok := defineSeatPositionType(newSeat.Type)
			if ok {
				if newSeat.Number != 0 {
					InvalidParams.SetInvalidParam(fmt.Sprintf("structure[%d]", seatIndex), "If the structure object is not a seat its number has to be 0.")
				} else if InvalidParams == nil {
					bus.Structure[rowIndex].Positions = append(bus.Structure[rowIndex].Positions, SeatPosition{
						Type:     seatPositionType,
						Position: seatIndex,
					})
				}
			} else if seatType, ok := defineSeatType(newSeat.Type); !ok {
				InvalidParams.SetInvalidParam(fmt.Sprintf("structure[%d]", seatIndex), "Non-existing seat type '"+newSeat.Type+"'.")
			} else {

				if newSeat.Number == 0 {
					InvalidParams.SetInvalidParam(fmt.Sprintf("structure[%d]", seatIndex), "Seat number cannot be 0.")

				}
				newSeatDirection, ok := defineSeatDirection(newSeat.Direction)

				if !ok {
					InvalidParams.SetInvalidParam(fmt.Sprintf("structure[%d]", seatIndex), "Seat direction has to be either 'Forward' or 'Backward'.")

				}

				if InvalidParams != nil {
					continue
				}

				bus.Structure[rowIndex].Positions = append(bus.Structure[rowIndex].Positions, SeatPosition{
					SeatNumber: newSeat.Number,
					Type:       SeatPossitionTypeSeat,
					Position:   seatIndex,
				})

				bus.Seats = append(bus.Seats, Seat{
					Number:    newSeat.Number,
					Type:      seatType,
					Direction: newSeatDirection,
				})
			}
		}

	}

	return bus, InvalidParams
}

// BusID     uuid.UUID     `gorm:"type:uuid"                                                                   `
// 	Number    int           `gorm:"type:tinyint;not null"                                                       `
// 	Type      seatType      `gorm:"type:enum('Window', 'Single', 'Single-Window', 'Aisle', 'Midlle');not null"  `
// 	Direction seatDirection `gorm:"type:enum('Forward', 'Backward');not null"

//
