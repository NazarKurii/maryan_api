package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"time"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Ticket interface {
	Create(ctx context.Context, ticket []*entity.Ticket) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Ticket, error)
	GetTickets(ctx context.Context, pagination dbutil.Pagination) ([]entity.Ticket, []entity.Connection, int, error, bool)
	Delete(ctx context.Context, id uuid.UUID) error
	ChangeConnection(ctx context.Context, id, connectionID uuid.UUID) error
	ChangePassenger(ctx context.Context, id, passengerID uuid.UUID) error
	Complete(ctx context.Context, id uuid.UUID) error
	DeleteTickets(ctx context.Context, paymentSessionID string) error
	AddTickets(ctx context.Context, paymentSessionID string) error
}

type ticketMySQL struct {
	db *gorm.DB
}

func (ds *ticketMySQL) AddTickets(ctx context.Context, paymentSessionID string) error {
	var tickets []entity.Ticket
	err := dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).
		Where("id IN  (SELECT ticket_id FROM ticket_payments WHERE session_id = ?)", paymentSessionID).
		Find(&tickets), "non-existing-session ")
	if err != nil {
		return err
	}

	var stops = make([]*entity.Stop, 0, len(tickets)*2)

	for _, ticket := range tickets {
		id := uuid.New()
		stops = append(stops, &entity.Stop{
			ID:           id,
			TicketID:     ticket.ID,
			ConnectionID: ticket.ConnectionID,
			Type:         entity.PickUpStopType,
			Updates: []entity.StopUpdate{
				{
					StopID: id,
					Status: entity.ConfirmedStopStatus,
				},
			},
		})

		id = uuid.New()
		stops = append(stops, &entity.Stop{
			ID:           id,
			TicketID:     ticket.ID,
			ConnectionID: ticket.ConnectionID,
			Type:         entity.DropOffStopType,
			Updates: []entity.StopUpdate{
				{
					StopID: id,
					Status: entity.ConfirmedStopStatus,
				},
			},
		})
	}

	return dbutil.PossibleCreateError(ds.db.WithContext(ctx).Create(stops), "non-existing-connection")
}

func (ds *ticketMySQL) DeleteTickets(ctx context.Context, paymentSessionID string) error {
	var ticketIDs []uuid.UUID
	err := dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Table("tickets").
		Select("id").Where("id IN  (SELECT ticket_id FROM ticket_payments WHERE session_id = ?)", paymentSessionID).
		Find(&ticketIDs), "non-existing-session")
	if err != nil {
		return err
	}

	var passengerIDs []uuid.UUID
	err = dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Table("tickets").
		Select("passenger_id").Where("id IN  (?)", ticketIDs).
		Find(&passengerIDs), "non-existing-session")
	if err != nil {
		return err
	}

	var adressIDs []uuid.UUID
	err = dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Raw(`
	 SELECT pick_up_adress_id FROM tickets WHERE id IN  (?) UNION SELECT drop_off_adress_id FROM tickets WHERE id IN  (?)
	`, ticketIDs, ticketIDs).
		Scan(&adressIDs), "non-existing-session")
	if err != nil {
		return err
	}

	err = dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).
		Where("ticket_id IN  (?)", ticketIDs).Unscoped().Delete(&entity.TicketPayment{}), "non-existing-session")
	if err != nil {
		return err
	}

	err = dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Where("id IN  (?)", ticketIDs).Unscoped().Delete(&entity.Ticket{}), "non-existing-session")
	if err != nil {
		return err
	}

	err = dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).
		Where("id IN (?)", passengerIDs).
		Unscoped().Delete(&entity.Passenger{}), "non-existing-session")
	if err != nil {
		return err
	}

	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).
		Where("id IN (?)", adressIDs).
		Unscoped().Delete(&entity.Address{}), "non-existing-session")
}

func (ds *ticketMySQL) Create(ctx context.Context, tickets []*entity.Ticket) error {

	return dbutil.PossibleCreateError(ds.db.WithContext(ctx).Create(tickets), "ticket-data")
}

func (ds *ticketMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Ticket, error) {
	var ticket = entity.Ticket{ID: id}
	return ticket, dbutil.PossibleFirstError(ds.db.WithContext(ctx).Preload(clause.Associations).First(&ticket), "non-existing-ticket")
}

func (ds *ticketMySQL) GetTickets(ctx context.Context, pagination dbutil.Pagination) ([]entity.Ticket, []entity.Connection, int, error, bool) {
	tickets, total, err, empty := dbutil.Paginate[entity.Ticket](ctx, ds.db, pagination, clause.Associations)
	if err != nil && empty {
		return nil, nil, 0, err, true
	}

	var connectionIDs = make([]uuid.UUID, len(tickets))
	for i, ticket := range tickets {
		connectionIDs[i] = ticket.ConnectionID
	}

	var connections []entity.Connection

	return tickets, connections, total, dbutil.PossibleDbError(
		ds.db.WithContext(ctx).
			Preload(clause.Associations).
			Where(
				"id IN (?)", connectionIDs,
			).
			Group("id").
			Find(&connections),
	), false
}

func (ds *ticketMySQL) Delete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Delete(&entity.Ticket{}, id), "non-existing-ticket")
}

func (ds *ticketMySQL) ChangeConnection(ctx context.Context, id, connectionID uuid.UUID) error {
	return dbutil.PossibleForeignKeyError(ds.db.WithContext(ctx).Where("id = ?", id).Update("connection_id", connectionID), "non-existing-ticket", "non-existing-connection", "invalid-id")
}

func (ds *ticketMySQL) ChangePassenger(ctx context.Context, id uuid.UUID, passengerID uuid.UUID) error {
	return dbutil.PossibleForeignKeyError(ds.db.WithContext(ctx).Where("id = ?", id).Update("passenger_id", passengerID), "non-existing-ticket", "non-existing-passenger", "invalid-id")
}

func (ds *ticketMySQL) Complete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Where("id = ?", id).Update("completed_at", time.Now().UTC()), "non-existing-ticket")
}

func NewTicket(db *gorm.DB) Ticket {
	return &ticketMySQL{db}
}
