package dataStore

import "gorm.io/gorm"

type Driver interface {
	User
}

type driverMySQL struct {
	userMySQL
}

func NewDriver(db *gorm.DB) Driver {
	return &driverMySQL{userMySQL{db}}
}
