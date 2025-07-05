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
	Rows               []Row        `gorm:"foreignKey:BusID" json:"rows"`
	CreatedAt          time.Time    `gorm:"not null" json:"createdAt"`
	UpdatedAt          time.Time    `gorm:"not null" json:"updatedAt"`
	DeletedAt          sql.NullTime `gorm:"index" json:"deletedAt"`
}

type Row struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	BusID  uuid.UUID `gorm:"type:uuid;not null" json:"busID"`
	Number uint8     `gorm:"not null" json:"number"`
	Seats  []Seat    `json:"seats"`
}

type Seat struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	RowID        uuid.UUID `gorm:"type:uuid;not null" json:"rowID"`
	Number       uint8     `gorm:"not null" json:"number"`
	Separated    bool      `gorm:"not null" json:"separated"`
	NextToDriver bool      `gorm:"not null" json:"nextToDriver"`
	Space        bool      `gorm:"not null" json:"space"`
	RowPosition  int       `gorm:"type:tinyint;not null" json:"rowPosition"`
	Window       bool      `gorm:"not null" json:"window"`
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

	if len(b.Rows) == 0 {
		params.SetInvalidParam("seats", "Invalid bus seats number, must be greater than 0.")
	}

	if len(b.Images) != 0 {
		b.Images = nil
	}

	b.ID = uuid.New()

	for i, row := range b.Rows {
		if len(row.Seats) == 0 {
			params.SetInvalidParam(fmt.Sprintf("row (index:%d)", i), "Empty row.")
			continue
		}
		b.Rows[i].ID = uuid.New()
		b.Rows[i].BusID = b.ID

		for j, seat := range row.Seats {

			if seat.Number < 1 && !seat.Space {
				params.SetInvalidParam(fmt.Sprintf("Row(index:%d) seat (index:%d)", j, i), "Invalid seat number.")
			}
			if seat.Number > 0 && seat.Space {
				params.SetInvalidParam(fmt.Sprintf("Row(index:%d) seat (index:%d)", j, i), "Invalid seat number. Must be 0 if the seat is 'space'")
			}

			if seat.RowPosition < 1 {
				params.SetInvalidParam(fmt.Sprintf("Row(index:%d) seat (index:%d)", j, i), "Invalid  seat row position.")
			}
			b.Rows[i].Seats[j].ID = uuid.New()
			b.Rows[i].Seats[j].RowID = b.Rows[i].ID
		}

	}

	return params
}

func MigrateBus(db *gorm.DB) {
	db.AutoMigrate(
		&Bus{},
		&Row{},
		&Seat{},
	)
}
