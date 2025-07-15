package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Ticket interface {
	Create(ctx context.Context, ticket *entity.Ticket) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Ticket, error)
	GetTickets(ctx context.Context, pagination dbutil.Pagination) ([]entity.Ticket, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ChangeConnection(ctx context.Context, id, connectionID uuid.UUID) error
	ChangePassenger(ctx context.Context, id, passengerID uuid.UUID) error
	Complete(ctx context.Context, id uuid.UUID) error
}

type ticketMySQL struct {
	db *gorm.DB
}

func (ds ticketMySQL) Create(ctx context.Context, ticket *entity.Ticket) error {
	return dbutil.PossibleCreateError(ds.db.WithContext(ctx).Create(ticket), "ticket-data")
}

func (ds ticketMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Ticket, error) {
	var ticket = entity.Ticket{ID: id}
	return ticket, dbutil.PossibleFirstError(ds.db.WithContext(ctx).Preload(clause.Associations).First(&ticket), "non-existing-ticket")
}

func (ds ticketMySQL) GetTickets(ctx context.Context, pagination dbutil.Pagination) ([]entity.Ticket, int, error) {
	return dbutil.Paginate[entity.Ticket](ctx, ds.db, pagination, clause.Associations)
}

func (ds ticketMySQL) Delete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Delete(&entity.Ticket{}, id), "non-existing-ticket")
}

func (ds ticketMySQL) ChangeConnection(ctx context.Context, id, connectionID uuid.UUID) error {
	return dbutil.PossibleForeignKeyError(ds.db.WithContext(ctx).Where("id = ?", id).Update("connection_id", connectionID), "non-existing-ticket", "non-existing-connection", "invalid-id")
}

func (ds ticketMySQL) ChangePassenger(ctx context.Context, id uuid.UUID, passengerID uuid.UUID) error {
	return dbutil.PossibleForeignKeyError(ds.db.WithContext(ctx).Where("id = ?", id).Update("passenger_id", passengerID), "non-existing-ticket", "non-existing-passenger", "invalid-id")
}

func (ds ticketMySQL) Complete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Where("id = ?", id).Update("completed_at", time.Now().UTC()), "non-existing-ticket")
}

func NewTicket(db *gorm.DB) Ticket {
	return &ticketMySQL{db}
}
