package dataStore

import (
	"maryan_api/internal/entity"
	"maryan_api/pkg/log"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	entity.MigrateUser(db)

	log.Migrate(db)

	return nil
}
