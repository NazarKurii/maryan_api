package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Connection interface {
	GetByID(ctx context.Context, id uuid.UUID) (entity.Connection, error)
	GetConnections(ctx context.Context, pagination dbutil.Pagination) ([]entity.Connection, int, error)
	ChangeDepartureTime(ctx context.Context, id uuid.UUID, departureTime time.Time) error
	ChangeGoogleMapsURL(ctx context.Context, id uuid.UUID, url string) error
	GetCurrentBusID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	ChangeBus(ctx context.Context, id, currentBusID, replasingBusID uuid.UUID) error
	RegisterUpdate(ctx context.Context, update *entity.ConnectionUpdate) error
	ChangeType(ctx context.Context, id uuid.UUID, connectionType entity.ConnectionType) error
}

type connectionMySQL struct {
	db *gorm.DB
}

func (ds connectionMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Connection, error) {
	var connection = entity.Connection{ID: id}
	return connection, dbutil.PossibleFirstError(dbutil.Preload(ds.db.WithContext(ctx), entity.PreloadConnection()...).First(&connection), "non-existing-connection")
}

func (ds connectionMySQL) GetConnections(ctx context.Context, pagination dbutil.Pagination) ([]entity.Connection, int, error) {
	return dbutil.Paginate[entity.Connection](ctx, ds.db, pagination, entity.PreloadConnection()...)
}

func (ds connectionMySQL) ChangeDepartureTime(ctx context.Context, id uuid.UUID, departureTime time.Time) error {
	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Where("id = ?", id).Update("departure_time", departureTime), "non-existing-connection")
}

func (ds connectionMySQL) ChangeGoogleMapsURL(ctx context.Context, id uuid.UUID, url string) error {
	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Where("id = ?", id).Update("google_maps_url", url), "non-existing-connection")
}

func (ds connectionMySQL) GetCurrentBusID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var currentBusID uuid.UUID
	return currentBusID, dbutil.PossibleRawsAffectedError(
		ds.db.WithContext(ctx).
			Model(&entity.Connection{}).
			Where("id = ?", id).Select("bus_id").
			Scan(&currentBusID),
		"non-existing-connection",
	)
}

func (ds connectionMySQL) ChangeBus(ctx context.Context, id, currentBusID, replasingBusID uuid.UUID) error {
	return dbutil.PossibleForeignKeyError(
		ds.db.WithContext(ctx).
			Where("id = ?", id).
			Updates(&entity.Connection{BusID: replasingBusID, ReplacedBusID: currentBusID}),
		"non-existing-connection",
		"non-exisitng-bus",
		"invalid-id",
	)
}

func (ds connectionMySQL) RegisterUpdate(ctx context.Context, update *entity.ConnectionUpdate) error {
	return dbutil.PossibleForeignKeyCreateError(ds.db.WithContext(ctx).Create(update), "non-existing-connection", "connection-update-data")
}

func (ds connectionMySQL) ChangeType(ctx context.Context, id uuid.UUID, connectionType entity.ConnectionType) error {
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

func (ds stopMySQL) Create(ctx context.Context, stop *entity.Stop) error {
	return dbutil.PossibleCreateError(ds.db.WithContext(ctx).Create(stop), "stop-data")
}

func (ds stopMySQL) Delete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Delete(&entity.Stop{ID: id}), "non-existing-stop")
}

func (ds stopMySQL) RegisterUpdate(ctx context.Context, update *entity.StopUpdate) error {
	return dbutil.PossibleForeignKeyCreateError(ds.db.WithContext(ctx).Create(update), "non-existing-connection", "stop-update-data")
}

func NewStop(db *gorm.DB) Stop {
	return stopMySQL{db}
}
