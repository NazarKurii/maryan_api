package testdata

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nazarkurii/marshrutka_api/config"
	"github.com/nazarkurii/marshrutka_api/internal/entity"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
)

// SET foreign_key_checks = 0;
// SET @schema = DATABASE();

// SET SESSION group_concat_max_len = 1000000;

// SELECT CONCAT('DROP TABLE ', GROUP_CONCAT(CONCAT('`', table_name, '`'))) INTO @drop_sql
// FROM information_schema.tables
// WHERE table_schema = @schema;

// PREPARE stmt FROM @drop_sql;
// EXECUTE stmt;
// DEALLOCATE PREPARE stmt;
// SET foreign_key_checks = 1;

func CreateTestData(db *gorm.DB) {
	db.Exec(`
INSERT INTO countries (id, name) VALUES
(UUID_TO_BIN(UUID()), 'Poland'),
(UUID_TO_BIN(UUID()), 'Germany'),
(UUID_TO_BIN(UUID()), 'Czechia'),
(UUID_TO_BIN(UUID()), 'Estonia'),
(UUID_TO_BIN(UUID()), 'Latvia'),
(UUID_TO_BIN(UUID()), 'Lithuania'),
(UUID_TO_BIN(UUID()), 'Slovakia'),
(UUID_TO_BIN(UUID()), 'Hungary'),
(UUID_TO_BIN(UUID()), 'Ukraine');
`)
	config.LoadCountries(db)
	drivers := entity.TestDrivers()
	err := db.Create(drivers).Error
	if err != nil {
		panic(err)
	}

	buses := entity.TestBuses()

	for i, _ := range buses {
		buses[i].LeadDriverID = uuid.NullUUID{drivers[i*2].ID, true}
		buses[i].AssistantDriverID = uuid.NullUUID{drivers[(i*2)+1].ID, true}
	}

	err = db.Create(buses).Error
	if err != nil {
		panic(err)
	}

	var line = 100
	countries, ukraineID := config.GetCountries()
	var busIndex int
	fmt.Println(len(countries))
	for _, countryID := range countries {
		if countryID == ukraineID {
			continue
		}

		departureTime := time.Now()
		var trips = make([]entity.Trip, 50)
		for i := 0; i < 50; i++ {

			tripiID := uuid.New()
			outboundConnectionID := uuid.New()
			returnConnectionID := uuid.New()
			trips[i] = entity.Trip{
				ID: tripiID,
				OutboundConnection: entity.Connection{
					ID:                   outboundConnectionID,
					Line:                 line,
					Price:                (rand.Intn(250) + 100) * 100,
					DepartureCountryID:   ukraineID,
					DestinationCountryID: countryID,
					DepartureTime:        departureTime.UTC(),
					ArrivalTime:          departureTime.Add(time.Hour*15 + time.Hour*time.Duration(rand.Intn(20))).UTC(),
					BusID:                buses[busIndex].ID,
					Updates: []entity.ConnectionUpdate{{
						ConnectionID: outboundConnectionID,
						Status:       entity.RegisteredConnectionStatus,
					}},
					Type:              entity.ComertialConnectionType,
					SellBefore:        departureTime.Add(-time.Hour * 24).UTC(),
					BackpackPrice:     int(config.GetLoggageConfig().Small.Price),
					SmallLuggagePrice: int(config.GetLoggageConfig().Medium.Price),
					LargeLuggagePrice: int(config.GetLoggageConfig().Large.Price),
					LuggageVolumeLeft: uint(buses[busIndex].LuggageVolume),
					MaxLength:         int(buses[busIndex].MaxLength),
					MaxHeight:         int(buses[busIndex].MaxHeight),
					MaxWidth:          int(buses[busIndex].MaxWidth),
				},
				ReturnConnection: entity.Connection{
					ID:                   returnConnectionID,
					Line:                 line,
					Price:                (rand.Intn(250) + 100) * 100,
					DepartureCountryID:   countryID,
					DestinationCountryID: ukraineID,
					DepartureTime:        departureTime.Add(time.Hour * 60).UTC(),
					ArrivalTime:          departureTime.Add(time.Hour*60 + time.Hour*15 + time.Hour*time.Duration(rand.Intn(20))).UTC(),
					BusID:                buses[busIndex].ID,
					Updates: []entity.ConnectionUpdate{{
						ConnectionID: returnConnectionID,
						Status:       entity.RegisteredConnectionStatus,
					}},
					Type:              entity.ComertialConnectionType,
					SellBefore:        departureTime.Add(time.Hour * 36).UTC().UTC(),
					BackpackPrice:     int(config.GetLoggageConfig().Small.Price),
					SmallLuggagePrice: int(config.GetLoggageConfig().Medium.Price),
					LargeLuggagePrice: int(config.GetLoggageConfig().Large.Price),
					LuggageVolumeLeft: uint(buses[busIndex].LuggageVolume),
				},
				Updates: []entity.TripUpdate{{
					TripID: tripiID,
					Status: entity.TripStatusRegistered,
				}},
			}

			departureTime = departureTime.Add(time.Hour * 24 * 10)
		}
		err = db.Create(trips).Error
		if err != nil {
			panic(err)
		}

		line += 100
		busIndex++
		departureTime = time.Now()
	}
}
