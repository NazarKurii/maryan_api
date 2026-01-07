package dataStore

import (
	"github.com/nazarkurii/marshrutka_api/internal/entity"
	"github.com/nazarkurii/marshrutka_api/internal/valueobject"
	"github.com/nazarkurii/marshrutka_api/pkg/log"

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
	errCheck(entity.MigratePackage(db))
	errCheck(entity.MigrateTrip(db))
	errCheck(valueobject.MigrateVerifications(db))
	errCheck(log.Migrate(db))
	errCheck(entity.MigrateTicket(db))

	errCheck(entity.MigrateConnection(db))
	// testdata.CreateTestData(db)
	return nil
}
