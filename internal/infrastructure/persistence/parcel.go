package dataStore

import (
	"context"
	"fmt"
	"time"

	"github.com/nazarkurii/marshrutka_api/config"
	"github.com/nazarkurii/marshrutka_api/internal/entity"
	"github.com/nazarkurii/marshrutka_api/pkg/dbutil"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Parsel interface {
	GetConnectionsByMonth(ctx context.Context, fromCountry, toCountry uuid.UUID, monthNumber, year int) ([]entity.Connection, error)
	CreateParcelStops(ctx context.Context, paymentSessionID string) error
	Create(ctx context.Context, parcel *entity.Parcel) error
	GetParcels(ctx context.Context, pagination dbutil.Pagination) ([]entity.Parcel, []entity.Connection, int, error, bool)
	RemoveParcelStops(ctx context.Context, paymentSessionID string) error
	PaymentSucceeded(ctx context.Context, paymentSessionID string) error
	DeleteParcels(ctx context.Context, paymentSessionID string) error
}

type parselMysql struct {
	db *gorm.DB
}

func (pds *parselMysql) GetConnectionsByMonth(ctx context.Context, fromCountry, toCountry uuid.UUID, monthNumber, year int) ([]entity.Connection, error) {
	var connections []entity.Connection

	err := dbutil.PossibleDbError(dbutil.Preload(pds.db, entity.PreloadConnection()...).WithContext(ctx).Table("connections").
		Joins("LEFT JOIN buses ON connections.bus_id = buses.id").
		Joins("LEFT JOIN (SELECT SUM(luggage_volume) AS luggage_volume, connection_id FROM parcels GROUP BY connection_id) parcels ON parcels.connection_id = connections.id").
		Joins("LEFT JOIN (SELECT SUM(luggage_volume) AS luggage_volume, connection_id FROM tickets GROUP BY connection_id) tickets ON tickets.connection_id = connections.id").
		Joins("LEFT JOIN (SELECT COUNT(id) AS number, bus_id FROM seats GROUP BY bus_id) bus_seats ON bus_seats.bus_id = buses.id").
		Joins(`LEFT JOIN (
                    SELECT connections.id AS connection_id, COUNT(ticket_seats.seat_id) AS number
                    FROM ticket_seats
                    JOIN tickets ON tickets.id = ticket_seats.ticket_id
                    JOIN connections ON connections.id = tickets.connection_id
                    GROUP BY connections.id
               ) passengers ON passengers.connection_id = connections.id`).
		Where(
			` connections.departure_country_id = ? 
     AND connections.destination_country_id = ? 
     AND connections.sell_before > ? AND YEAR(departure_time) = ? AND MONTH(departure_time)= ?`,
			fromCountry,
			toCountry,
			config.MustParseToLocalByUUID(time.Now(), fromCountry).UTC(),
			year, monthNumber,
		).Order("connections.departure_time ASC").Find(&connections))

	if err != nil {
		return nil, err
	}

	for i, connection := range connections {
		var takenLuggageVolume uint
		var passengersNumber int
		for _, stop := range connection.Stops {
			if stop.LocationType == entity.PickUpStopType {
				if stop.Type == entity.PassengerStopType {
					passengersNumber++
				}
				takenLuggageVolume += uint(stop.Ticket.LuggageVolume) + uint(stop.Parcel.LuggageVolume)

			}
		}

		luggage := config.GetLoggageConfig()

		connections[i].LuggageVolumeLeft = uint(connection.Bus.LuggageVolume) - takenLuggageVolume - uint((len(connection.Bus.Seats)-passengersNumber)*(luggage.Small.Volume+luggage.Large.Volume))
	}
	return connections, nil
}

func (ds *parselMysql) CreateParcelStops(ctx context.Context, paymentSessionID string) error {
	var parcel entity.Parcel
	err := dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).
		Where("id IN  (SELECT parcel_id FROM parcel_payments WHERE session_id = ?)", paymentSessionID).
		Find(&parcel), "non-existing-session ")
	if err != nil {
		return err
	}

	var stops = make([]*entity.Stop, 2)

	id := uuid.New()
	stops[0] = &entity.Stop{
		ID: id,
		ParcelID: uuid.NullUUID{
			parcel.ID,
			true,
		},
		ConnectionID: parcel.ConnectionID,
		LocationType: entity.PickUpStopType,
		Type:         entity.ParcelStopType,
		Updates: []entity.StopUpdate{
			{
				StopID: id,
				Status: entity.ConfirmedStopStatus,
			},
		},
	}

	id = uuid.New()
	stops[1] = &entity.Stop{
		ID: id,
		ParcelID: uuid.NullUUID{
			parcel.ID,
			true,
		},
		ConnectionID: parcel.ConnectionID,
		LocationType: entity.DropOffStopType,
		Type:         entity.ParcelStopType,
		Updates: []entity.StopUpdate{
			{
				StopID: id,
				Status: entity.ConfirmedStopStatus,
			},
		},
	}

	return dbutil.PossibleCreateError(ds.db.WithContext(ctx).Create(stops), "non-existing-connection")
}

func (ds *parselMysql) Create(ctx context.Context, parcel *entity.Parcel) error {
	return dbutil.PossibleCreateError(ds.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(parcel), "parcel-data")
}

func (ds *parselMysql) GetParcels(ctx context.Context, pagination dbutil.Pagination) ([]entity.Parcel, []entity.Connection, int, error, bool) {

	parcels, total, err, empty := dbutil.Paginate[entity.Parcel](ctx, ds.db, pagination, clause.Associations)
	if err != nil && empty {
		return nil, nil, 0, err, true
	}

	var connectionIDs = make([]uuid.UUID, len(parcels))
	for i, parcel := range parcels {
		connectionIDs[i] = parcel.ConnectionID
	}

	var connections []entity.Connection

	return parcels, connections, total, dbutil.PossibleDbError(
		ds.db.WithContext(ctx).
			Preload(clause.Associations).
			Where(
				"id IN (?)", connectionIDs,
			).
			Group("id").
			Find(&connections),
	), false
}

func (ds *parselMysql) PaymentSucceeded(ctx context.Context, paymentSessionID string) error {
	return dbutil.PossibleRawsAffectedError(ds.db.Table("parcel_payments").Where("session_id = ?", paymentSessionID).Update("succeeded", true))
}

//cs_test_a1CrwsqWkRwxQ3RGa7YPrNeryQ3zqEEORgWmyFtg8WNX1xPziGQ1BhoQtx

func (ds *parselMysql) RemoveParcelStops(ctx context.Context, paymentSessionID string) error {
	err := dbutil.PossibleRawsAffectedError(ds.db.Unscoped().Table("stop_updates").Where("stop_id IN (SELECT id FROM stops WHERE parcel_id IN (SELECT parcel_id FROM parcel_payments WHERE session_id = ?))", paymentSessionID).Delete(&entity.Stop{}), "non-existing-data")
	if err != nil {
		return err
	}

	return dbutil.PossibleRawsAffectedError(ds.db.Unscoped().Table("stops").Where("parcel_id IN (SELECT parcel_id FROM parcel_payments WHERE session_id = ?)", paymentSessionID).Delete(&entity.Stop{}), "non-existing-data")
}

func (ds *parselMysql) DeleteParcels(ctx context.Context, paymentSessionID string) error {
	return ds.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var parcelIDs []uuid.UUID
		err := dbutil.PossibleRawsAffectedError(
			tx.Table("parcels").
				Select("id").
				Where("id IN (SELECT parcel_id FROM parcel_payments WHERE session_id = ?)", paymentSessionID).
				Find(&parcelIDs),
		)

		if err != nil {
			return err
		} else if len(parcelIDs) == 0 {
			return fmt.Errorf("non-existing-session")
		}

		var addressIDs []uuid.UUID
		err = dbutil.PossibleRawsAffectedError(tx.Raw(`
			SELECT pick_up_adress_id FROM parcels WHERE id IN (?) 
			UNION 
			SELECT drop_off_adress_id FROM parcels WHERE id IN (?)
		`, parcelIDs, parcelIDs).Scan(&addressIDs))
		if err != nil {
			return err
		}

		err = dbutil.PossibleRawsAffectedError(tx.Where("parcel_id IN (?)", parcelIDs).Unscoped().Delete(&entity.ParcelPayment{}))
		if err != nil {
			return err
		}

		err = dbutil.PossibleRawsAffectedError(tx.Where("id IN (?)", parcelIDs).Unscoped().Delete(&entity.Parcel{}))
		if err != nil {
			return err
		}

		return dbutil.PossibleRawsAffectedError(tx.Where("id IN (?)", addressIDs).Unscoped().Delete(&entity.Address{}))
	})
}
func NewParsel(db *gorm.DB) Parsel {
	return &parselMysql{db}
}
