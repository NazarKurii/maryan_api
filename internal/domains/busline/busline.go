package busline

// import (
// 	"maryan_api/internal/modules/adress"
// 	"maryan_api/internal/modules/bus"
// 	"maryan_api/internal/modules/passenger"
// 	"time"
// )

// type BusLine struct {
// 	ID                 uint      `gorm:"primaryKey; autoincrement"`
// 	CreatedAt          time.Time `gorm:"not null"`
// 	UpdatedAt          time.Time `gorm:"not null"`
// 	DepartureCountry   string    `gorm:"type:varchar(56);not null"`
// 	DestinationCountry string    `gorm:"type:varchar(56);not null"`
// 	DepartureTime      time.Time `gorm:"not null"`
// 	ArrivalTime        time.Time `gorm:"not null"`
// 	EstimatedDuration  uint16    `gorm:"-"`
// 	BusID              uint      `gorm:"not null"`
// 	Bus                bus.Bus   `gorm:"foreignKey:BusID"`
// 	Stops              []Stop    `gorm:"foreignKey:BusLineID"`
// }

// type Stop struct {
// 	BusLineID uint          `gorm:"not null"`
// 	AdressID  uint          `gorm:"not null"`
// 	Adress    adress.Adress `gorm:"foreignKey:AdressID"`
// }

// type BusLinePassenger struct {
// 	BusLineID   uint                `gorm:"not null"`
// 	PassengerID uint                `gorm:"not null"`
// 	Passenger   passenger.Passenger `gorm:"foreignKey:PassengerID"`
// 	BusLine     BusLine             `gorm:"foreignKey:BusLineID"`
// }
