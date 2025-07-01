package repo

import "gorm.io/gorm"

type PassengerRepo interface {
}

type passengerRepoMySQL struct {
	db *gorm.DB
}

func NewPassengerRepoMysql(db *gorm.DB) PassengerRepo {
	return passengerRepoMySQL{db}
}
