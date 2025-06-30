package migrations

import (
	"maryan_api/internal/domains/trip"
	"maryan_api/internal/domains/user"
	"maryan_api/pkg/log"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	trip.Migrate(db)
	user.Migrate(db)
	log.Migrate(db)
	return nil
}
