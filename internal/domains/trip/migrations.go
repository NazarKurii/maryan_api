package trip

import "gorm.io/gorm"

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&Bus{},
		&BusImage{},
		&Row{},
		&Seat{},
	)

	if err != nil {
		panic(err)
	}
}
