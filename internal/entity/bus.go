package entity

import (
	"database/sql"
	"encoding/json"
	"fmt"
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Bus struct {
	ID                 uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	Model              string       `gorm:"type:varchar(255); not null" json:"model"`
	Images             BusImages    `gorm:"not null" json:"imageURL"`
	IsActive           bool         `gorm:"not null" json:"isActive"`
	RegistrationNumber string       `gorm:"type:varchar(8); not null; unique" json:"registrantionNumber"`
	Year               int          `gorm:"type:smallint; not null" json:"year"`
	GpsTrackerID       string       `gorm:"type:varchar(255); not null" json:"gpsTrackerID"`
	Seats              []Seat       `gorm:"foreignKey:BusID" json:"rows"`
	Structure          []Row        `gorm:"not null"`
	CreatedAt          time.Time    `gorm:"not null" json:"createdAt"`
	UpdatedAt          time.Time    `gorm:"not null" json:"updatedAt"`
	DeletedAt          sql.NullTime `gorm:"index" json:"deletedAt"`
}

type Seat struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	BusID  uuid.UUID `gorm:"type:uuid"`
	Number int       `gorm:"type:tinyint;not null" json:"number"`
	Type   seatType  `gorm:"type:enum('Driver','Window', 'Single', 'Single-Window');not null" json:"type"`
}

type seatType string

const (
	SingleSeatType       = "Single"
	SingleWindowSeatType = "Single-Window"
	WindowSeatType       = "Window"
	DriverSeatType       = "Driver"
)

type Row struct {
	BusID     uuid.UUID      `gorm:"type:uuid"`
	Positions []SeatPosition `gorm:"not null"`
}

type SeatPosition struct {
	Number  int  `gorm:"not null"`
	Empty   bool `gorm:"not null"`
	Postion int  `gorm:"not null"`
}

type BusImage struct {
	BusID uuid.UUID `gorm:"type:uuid;not null" json:"busID"`
	Url   string    `gorm:"type:varchar(255);not null" json:"url"`
}

type BusImages []BusImage

func (bi *BusImages) MarshalJSON() ([]byte, error) {
	var urls = make([]string, len(*bi))
	for i, biu := range *bi {
		urls[i] = biu.Url
	}

	return json.Marshal(urls)
}

func (bi *BusImages) UnmarshalJSON(data []byte) error {
	var urls []string
	err := json.Unmarshal(data, &urls)
	if err != nil {
		return err
	}

	for _, url := range urls {
		*bi = append(*bi, BusImage{Url: url})
	}

	return nil
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
		var correctSeat bool = true

		if seat.Number < 1 {
			params.SetInvalidParam(fmt.Sprintf("seat (index:%d)", i), "Invalid seat number.")
			correctSeat = false
		}

		if seatNumbers[seat.Number] == 1 {
			params.SetInvalidParam(fmt.Sprintf("seat (index:%d)", i), "Invalid seat number (Repeated).")
			correctSeat = false
		} else {
			seatNumbers[seat.Number]++
		}

		switch seat.Type {
		case DriverSeatType, WindowSeatType, SingleSeatType, SingleWindowSeatType:
			if correctSeat {
				b.Seats[i].BusID = b.ID
				b.Seats[i].ID = uuid.New()
			}
		default:
			params.SetInvalidParam(fmt.Sprintf("seat (index:%d)", i), "Invalid seat type.")
			continue
		}
	}

	for i, row := range b.Structure {
		var correctRow bool = true
		for j, seat := range row.Positions {
			switch {
			case seat.Number == 0 && !seat.Empty:
				params.SetInvalidParam(fmt.Sprintf("Structure row(index:%d) seatPosition(index:%d)", i, j), "Seat is not empty, but seat number is 0")
				correctRow = false
			case seat.Number != 0 && seat.Empty:
				params.SetInvalidParam(fmt.Sprintf("Structure row(index:%d) seatPosition(index:%d)", i, j), "Seat is  empty, but seat number is not 0")
				correctRow = false
			default:
				if seatNumbers[seat.Number] == 0 {
					params.SetInvalidParam(fmt.Sprintf("Structure row(index:%d) seatPosition(index:%d)", i, j), "Seat number is not unique")
					correctRow = false
				} else {
					seatNumbers[seat.Number]--
				}
			}
		}
		if correctRow {
			b.Structure[i].BusID = b.ID
		}
	}

	return params
}

func MigrateBus(db *gorm.DB) error {
	return db.AutoMigrate(
		&Bus{},
		&Row{},
		&Seat{},
		&Row{},
		&SeatPosition{},
	)
}
