package dataStore

import (
	"maryan_api/internal/entity"
	"maryan_api/internal/valueobject"
	"maryan_api/pkg/log"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	entity.MigrateUser(db)
	entity.MigrateBus(db)
	entity.MigratePassenger(db)

	valueobject.MigrateVerifications(db)

	log.Migrate(db)
	return nil
}
