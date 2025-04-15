package migrations

import (
	"maryan_api/internal/domains/user"
	"maryan_api/pkg/log"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&user.User{},

		&log.Log{},
		// &busline.BusLine{},
		// &busline.BusLinePassenger{},
		// &busline.Stop{},
		// &passenger.Passenger{},
		// &ticket.Ticket{},
	)
}
