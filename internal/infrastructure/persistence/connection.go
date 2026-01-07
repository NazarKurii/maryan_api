package dataStore

import (
	"context"
	"fmt"
	"time"

	"github.com/nazarkurii/marshrutka_api/config"
	"github.com/nazarkurii/marshrutka_api/internal/entity"
	"github.com/nazarkurii/marshrutka_api/pkg/dbutil"
	rfc7807 "github.com/nazarkurii/marshrutka_api/pkg/problem"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Connection interface {
	GetByID(ctx context.Context, id uuid.UUID, passengerNumber int) (entity.Connection, []uuid.UUID, error)
	GetConnections(ctx context.Context, pagination dbutil.Pagination) ([]entity.Connection, int, error, bool)
	ChangeDepartureTime(ctx context.Context, id uuid.UUID, departureTime time.Time) error
	ChangeGoogleMapsURL(ctx context.Context, id uuid.UUID, url string) error
	GetCurrentBusID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	// ChangeBus(ctx context.Context, id, currentBusID, replasingBusID uuid.UUID) error
	RegisterUpdate(ctx context.Context, update *entity.ConnectionUpdate) error
	ChangeType(ctx context.Context, id uuid.UUID, connectionType entity.ConnectionType) error
	FindConnections(ctx context.Context, request entity.FindConnectionsRequest) (FoundConnections, error)
}

type connectionMySQL struct {
	db *gorm.DB
}

type FoundConnections struct {
	Connections []entity.Connection
	TicketsLeft []TicketsLeft
	LeftRange   []entity.ConnectionsRange
	RightRange  []entity.ConnectionsRange
}

func (fc FoundConnections) ConnectionsIDs() []uuid.UUID {
	var connectionsIDs = make([]uuid.UUID, len(fc.Connections))
	for i, connection := range fc.Connections {
		connectionsIDs[i] = connection.ID
	}

	return connectionsIDs
}

type TicketsLeft struct {
	ID     uuid.UUID `gorm:"column:id"`
	Number float64   `gorm:"column:tickets_left"`
}

func (ds *connectionMySQL) FindConnections(
	ctx context.Context,
	request entity.FindConnectionsRequest,
) (FoundConnections, error) {
	var foundConnections FoundConnections

	// 1. Base connections
	if err := ds.findBaseConnections(ctx, request, &foundConnections); err != nil {
		return FoundConnections{}, err
	}

	// 2. Tickets left
	if err := ds.findTicketsLeft(ctx, &foundConnections); err != nil {
		return FoundConnections{}, err
	}

	// 3. Left range
	if err := ds.findLeftRange(ctx, request, &foundConnections); err != nil {
		return FoundConnections{}, err
	}

	// 4. Right range
	if err := ds.findRightRange(ctx, request, &foundConnections); err != nil {
		return FoundConnections{}, err
	}

	return foundConnections, nil
}

// --- private helpers ---

func (ds *connectionMySQL) findBaseConnections(
	ctx context.Context,
	request entity.FindConnectionsRequest,
	foundConnections *FoundConnections,
) error {
	return dbutil.PossibleDbError(
		ds.db.WithContext(ctx).
			Preload(clause.Associations).
			Where(
				"DATE(CONVERT_TZ(departure_time, 'UTC', ?)) = ? AND destination_country_id = ? AND departure_country_id = ?",
				config.MustGetLocationFromCountryID(request.From).String(),
				request.Date.Format("2006-01-02"),
				request.To,
				request.From,
			).
			Find(&foundConnections.Connections),
	)
}

func (ds *connectionMySQL) findTicketsLeft(
	ctx context.Context,
	foundConnections *FoundConnections,
) error {
	return dbutil.PossibleDbError(
		ds.db.WithContext(ctx).Raw(`
			SELECT 
				c.id AS id,
				COALESCE((
					SELECT COUNT(s.id)
					FROM buses b
					JOIN seats s ON b.id = s.bus_id
					WHERE b.id = c.bus_id
				), 0)
				-
				COALESCE((
					SELECT COUNT(st.id) / 2
					FROM stops st
					WHERE st.connection_id = c.id
					GROUP BY st.connection_id
				), 0) AS tickets_left
			FROM connections c
			WHERE c.id IN ?
		`, foundConnections.ConnectionsIDs()).
			Scan(&foundConnections.TicketsLeft),
	)
}

func (ds *connectionMySQL) findLeftRange(
	ctx context.Context,
	request entity.FindConnectionsRequest,
	foundConnections *FoundConnections,
) error {
	return dbutil.PossibleDbError(
		ds.db.WithContext(ctx).
			Table("connections").
			Select(
				"DATE(departure_time) AS date",
				"COUNT(id) AS number",
				"MIN(price) AS minPrice",
				"MAX(sell_before) AS sellBefore",
			).
			Where(
				"DATE(CONVERT_TZ(departure_time, 'UTC', ?)) < DATE(?) AND destination_country_id = ? AND departure_country_id = ?",
				config.MustGetLocationFromCountryID(request.From).String(),
				request.Date.Format("2006-01-02"),
				request.To,
				request.From,
			).
			Group("DATE(departure_time)").
			Order("DATE(departure_time) DESC").
			Limit(5).
			Scan(&foundConnections.LeftRange),
	)
}

func (ds *connectionMySQL) findRightRange(
	ctx context.Context,
	request entity.FindConnectionsRequest,
	foundConnections *FoundConnections,
) error {
	return dbutil.PossibleDbError(
		ds.db.WithContext(ctx).
			Table("connections").
			Select(
				"DATE(departure_time) AS date",
				"COUNT(id) AS number",
				"MIN(price) AS minPrice",
				"MAX(sell_before) AS sellBefore",
			).
			Where(
				"DATE(CONVERT_TZ(departure_time, 'UTC', ?)) > ? AND destination_country_id = ? AND departure_country_id = ?",
				config.MustGetLocationFromCountryID(request.From).String(),
				request.Date.Format("2006-01-02"),
				request.To,
				request.From,
			).
			Group("DATE(departure_time)").
			Order("DATE(departure_time) ASC").
			Limit(5).
			Scan(&foundConnections.RightRange),
	)
}

func (ds *connectionMySQL) GetByID(ctx context.Context, id uuid.UUID, passengersNumber int) (entity.Connection, []uuid.UUID, error) {
	var connection = entity.Connection{ID: id}
	err := dbutil.PossibleFirstError(dbutil.Preload(ds.db, entity.PreloadConnection()...).WithContext(ctx).First(&connection), "non-existing-connection")
	if err != nil {
		return entity.Connection{}, nil, err
	}

	var takenSeatsIDs []uuid.UUID
	var takenLuggageVolume uint
	for _, stop := range connection.Stops {
		if stop.LocationType == entity.PickUpStopType {

			if stop.Type == entity.PassengerStopType {

				for _, seat := range stop.Ticket.Seats {
					takenSeatsIDs = append(takenSeatsIDs, seat.SeatID)
				}
			}

			takenLuggageVolume += uint(stop.Ticket.LuggageVolume) + uint(stop.Parcel.LuggageVolume)
		}
	}

	busSeats := len(connection.Bus.Seats)
	takenSeatsLength := len(takenSeatsIDs)
	if busSeats < passengersNumber {
		return entity.Connection{}, nil, rfc7807.BadRequest("too-big-passengers-number", "Too Big Passengers Number Error", fmt.Sprintf("For this connections maximum is %s.", busSeats-takenSeatsLength))
	}
	luggageConfig := config.GetLoggageConfig()
	connection.LuggageVolumeLeft = uint(connection.Bus.LuggageVolume) - takenLuggageVolume - uint((busSeats)-takenSeatsLength+passengersNumber)*(uint(luggageConfig.Small.Volume)+uint(luggageConfig.Large.Volume))

	return connection, takenSeatsIDs, nil
}

func (ds *connectionMySQL) GetConnections(ctx context.Context, pagination dbutil.Pagination) ([]entity.Connection, int, error, bool) {
	return dbutil.Paginate[entity.Connection](
		ctx,
		ds.db.Joins(
			"JOIN buses ON buses.id = connections.bus_id",
			"JOIN tickets ON connections.id = tickets.connection_id",
			"JOIN users ON users.id = connections.user_id OR users.id = tickets.user_id",
			"JOIN passengers ON passengers.id = tickets.passenger_id",
		).Group("connection.id"),
		pagination,
		entity.PreloadConnection()...)
}

func (ds *connectionMySQL) ChangeDepartureTime(ctx context.Context, id uuid.UUID, departureTime time.Time) error {
	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Where("id = ?", id).Update("departure_time", departureTime), "non-existing-connection")
}

func (ds *connectionMySQL) ChangeGoogleMapsURL(ctx context.Context, id uuid.UUID, url string) error {
	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Where("id = ?", id).Update("google_maps_url", url), "non-existing-connection")
}

func (ds *connectionMySQL) GetCurrentBusID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var currentBusID uuid.UUID
	return currentBusID, dbutil.PossibleRawsAffectedError(
		ds.db.WithContext(ctx).
			Model(&entity.Connection{}).
			Where("id = ?", id).Select("bus_id").
			Scan(&currentBusID),
		"non-existing-connection",
	)
}

// func (ds *connectionMySQL) ChangeBus(ctx context.Context, id, currentBusID, replasingBusID uuid.UUID) error {
// 	return dbutil.PossibleForeignKeyError(
// 		ds.db.WithContext(ctx).
// 			Where("id = ?", id).
// 			Updates(&entity.Connection{BusID: replasingBusID, ReplacedBusID: currentBusID}),
// 		"non-existing-connection",
// 		"non-exisitng-bus",
// 		"invalid-id",
// 	)
// }

func (ds *connectionMySQL) RegisterUpdate(ctx context.Context, update *entity.ConnectionUpdate) error {
	return dbutil.PossibleForeignKeyCreateError(ds.db.WithContext(ctx).Create(update), "non-existing-connection", "connection-update-data")
}

func (ds *connectionMySQL) ChangeType(ctx context.Context, id uuid.UUID, connectionType entity.ConnectionType) error {
	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Where("id = ?", id).Update("type", connectionType.Val), "non-existing-connection")
}

func NewConnection(db *gorm.DB) Connection {
	return &connectionMySQL{db}
}

type Stop interface {
	Create(ctx context.Context, stop *entity.Stop) error
	Delete(ctx context.Context, id uuid.UUID) error
	RegisterUpdate(ctx context.Context, update *entity.StopUpdate) error
}

type stopMySQL struct {
	db *gorm.DB
}

func (ds *stopMySQL) Create(ctx context.Context, stop *entity.Stop) error {
	return dbutil.PossibleCreateError(ds.db.WithContext(ctx).Create(stop), "stop-data")
}

func (ds *stopMySQL) Delete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Delete(&entity.Stop{ID: id}), "non-existing-stop")
}

func (ds *stopMySQL) RegisterUpdate(ctx context.Context, update *entity.StopUpdate) error {
	return dbutil.PossibleForeignKeyCreateError(ds.db.WithContext(ctx).Create(update), "non-existing-connection", "stop-update-data")
}

func NewStop(db *gorm.DB) Stop {
	return &stopMySQL{db}
}
