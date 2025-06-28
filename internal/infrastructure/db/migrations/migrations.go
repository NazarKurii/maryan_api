package migrations

import (
	"maryan_api/internal/domains/user"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	user.Migrate(db)
	return nil
}
