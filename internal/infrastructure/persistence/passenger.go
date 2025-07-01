package dataStore

import "gorm.io/gorm"

type PassengerDataStore interface {
}

type passengerMySQL struct {
	db *gorm.DB
}

func NewPassengerRepoMysql(db *gorm.DB) PassengerDataStore {
	return passengerMySQL{db}
}
