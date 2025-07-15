package dataStore

import (
	"context"

	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"

	"gorm.io/gorm"
)

type Country interface {
	GetAll(ctx context.Context) ([]entity.Country, error)
}

type countryMySQL struct {
	db *gorm.DB
}

func (cds *countryMySQL) GetAll(ctx context.Context) ([]entity.Country, error) {
	var countries []entity.Country

	return countries, dbutil.PossibleRawsAffectedError(cds.db.Find(&countries), "no-countries-yet")
}

func NewCountry(db *gorm.DB) Country {
	return &countryMySQL{db}
}
