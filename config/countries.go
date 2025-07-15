package config

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var countries = map[string]uuid.UUID{}

func LoadCountries(db *gorm.DB) {
	var countriesSlice []struct {
		ID   uuid.UUID
		Name string
	}

	err := db.Table("countries").Select("id", "name").Scan(&countriesSlice).Error
	if err != nil {
		panic(err)
	}

	for _, country := range countriesSlice {
		countries[country.Name] = country.ID
	}
}

func CountryExists(country string) (uuid.UUID, bool) {
	id, ok := countries[country]
	return id, ok
}

func ParseToLocal(date time.Time, country string) (time.Time, error) {
	location, err := countryToLocation(country)
	if err != nil {
		return time.Time{}, err
	}

	return date.In(location), nil
}

func countryToLocation(country string) (*time.Location, error) {
	mapping := map[string]string{
		"Germany":   "Europe/Berlin",
		"Poland":    "Europe/Warsaw",
		"Czechia":   "Europe/Prague",
		"Estonia":   "Europe/Tallinn",
		"Latvia":    "Europe/Riga",
		"Lithuania": "Europe/Vilnius",
		"Slovakia":  "Europe/Bratislava",
		"Hungary":   "Europe/Budapest",
		"Ukraine":   "Europe/Kyiv", // official spelling since Go 1.15+
	}

	if tz, ok := mapping[country]; ok {
		return time.LoadLocation(tz)
	}
	return nil, fmt.Errorf("unsupported country: %s", country)
}
