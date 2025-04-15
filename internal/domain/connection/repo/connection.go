package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/dbutil"

	"time"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
)

type Connection interface {
	GetByID(ctx context.Context, id uuid.UUID) (entity.Connection, []uuid.UUID, error)
	GetConnections(ctx context.Context, pagination dbutil.Pagination) ([]entity.Connection, int, error, bool)
	ChangeDepartureTime(ctx context.Context, id uuid.UUID, departureTime time.Time) error
	ChangeGoogleMapsURL(ctx context.Context, id uuid.UUID, url string) error
	GetCurrentBusID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	// ChangeBus(ctx context.Context, id, currentBusID, replasingBusID uuid.UUID) error
	RegisterUpdate(ctx context.Context, update *entity.ConnectionUpdate) error
	ChangeType(ctx context.Context, id uuid.UUID, connectionType entity.ConnectionType) error
	FindConnections(ctx context.Context, request entity.FindConnectionsRequest) (dataStore.FoundConnections, error)
}

type connectionRepo struct {
	ds dataStore.Connection
}

func (r *connectionRepo) FindConnections(ctx context.Context, request entity.FindConnectionsRequest) (dataStore.FoundConnections, error) {
	return r.ds.FindConnections(ctx, request)
}

func (r *connectionRepo) GetByID(ctx context.Context, id uuid.UUID) (entity.Connection, []uuid.UUID, error) {
	return r.ds.GetByID(ctx, id)
}

func (r *connectionRepo) GetConnections(ctx context.Context, pagination dbutil.Pagination) ([]entity.Connection, int, error, bool) {
	return r.ds.GetConnections(ctx, pagination)
}

func (r *connectionRepo) ChangeDepartureTime(ctx context.Context, id uuid.UUID, departureTime time.Time) error {
	return r.ds.ChangeDepartureTime(ctx, id, departureTime)
}

func (r *connectionRepo) ChangeGoogleMapsURL(ctx context.Context, id uuid.UUID, url string) error {
	return r.ds.ChangeGoogleMapsURL(ctx, id, url)
}

func (r *connectionRepo) GetCurrentBusID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	return r.ds.GetCurrentBusID(ctx, id)
}

// func (r *connectionRepo) ChangeBus(ctx context.Context, id, currentBusID, replasingBusID uuid.UUID) error {
// 	return r.ds.ChangeBus(ctx, id, currentBusID, replasingBusID)
// }

func (r *connectionRepo) RegisterUpdate(ctx context.Context, update *entity.ConnectionUpdate) error {
	return r.ds.RegisterUpdate(ctx, update)
}

func (r *connectionRepo) ChangeType(ctx context.Context, id uuid.UUID, connectionType entity.ConnectionType) error {
	return r.ds.ChangeType(ctx, id, connectionType)
}

// Constructor
func NewConnectionRepo(db *gorm.DB) Connection {
	return &connectionRepo{dataStore.NewConnection(db)}
}
