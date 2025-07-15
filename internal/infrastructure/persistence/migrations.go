package dataStore

import (
	"maryan_api/internal/entity"
	"maryan_api/internal/valueobject"
	"maryan_api/pkg/log"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	errCheck(entity.MigrateUser(db))
	errCheck(entity.MigrateBus(db))
	errCheck(entity.MigratePassenger(db))
	errCheck(entity.MigrateAddress(db))
	errCheck(entity.MigrateTrip(db))
	errCheck(valueobject.MigrateVerifications(db))
	errCheck(log.Migrate(db))
	errCheck(entity.MigrateTicket(db))

	errCheck(entity.MigrateConnection(db))
	return nil
}
