package dataStore

import (
	"context"

	"github.com/nazarkurii/marshrutka_api/internal/entity"
	"github.com/nazarkurii/marshrutka_api/pkg/dbutil"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Ticket interface {
	Create(ctx context.Context, ticket *entity.Ticket) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Ticket, error)
	GetTickets(ctx context.Context, pagination dbutil.Pagination) ([]entity.Ticket, []entity.Connection, int, error, bool)
	// Delete(ctx context.Context, id uuid.UUID) error
	// ChangeConnection(ctx context.Context, id, connectionID uuid.UUID) error
	// ChangePassenger(ctx context.Context, id, passengerID uuid.UUID) error
	// Complete(ctx context.Context, id uuid.UUID) error
	DeleteTickets(ctx context.Context, paymentSessionID string) error
	CreatePassengerStops(ctx context.Context, paymentSessionID string) error
	RemovePassengerStops(ctx context.Context, paymentSessionID string) error
	PaymentSucceeded(ctx context.Context, paymentSessionID string) error
}

type ticketMySQL struct {
	db *gorm.DB
}

func (ds *ticketMySQL) PaymentSucceeded(ctx context.Context, paymentSessionID string) error {
	return dbutil.PossibleRawsAffectedError(ds.db.Table("ticket_payments").Where("session_id = ?", paymentSessionID).Update("succeeded", true))
}

func (ds *ticketMySQL) CreatePassengerStops(ctx context.Context, paymentSessionID string) error {
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
			ID: id,
			TicketID: uuid.NullUUID{
				ticket.ID,
				true,
			},
			ConnectionID: ticket.ConnectionID,
			LocationType: entity.PickUpStopType,
			Type:         entity.PassengerStopType,
			Updates: []entity.StopUpdate{
				{
					StopID: id,
					Status: entity.ConfirmedStopStatus,
				},
			},
		})

		id = uuid.New()
		stops = append(stops, &entity.Stop{
			ID: id,
			TicketID: uuid.NullUUID{
				ticket.ID,
				true,
			},
			ConnectionID: ticket.ConnectionID,
			LocationType: entity.DropOffStopType,
			Type:         entity.PassengerStopType,
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

func (ds *ticketMySQL) RemovePassengerStops(ctx context.Context, paymentSessionID string) error {
	err := dbutil.PossibleRawsAffectedError(ds.db.Unscoped().Table("stop_updates").Where("stop_id IN (SELECT id FROM stops WHERE ticket_id IN (SELECT ticket_id FROM ticket_payments WHERE session_id = ?))", paymentSessionID).Delete(&entity.Stop{}), "non-existing-data")
	if err != nil {
		return err
	}

	return dbutil.PossibleRawsAffectedError(ds.db.Unscoped().Table("stops").Where("ticket_id IN (SELECT ticket_id FROM ticket_payments WHERE session_id = ?)", paymentSessionID).Delete(&entity.Stop{}), "non-existing-data")
}

func (ds *ticketMySQL) DeleteTickets(ctx context.Context, paymentSessionID string) error {
	return ds.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var session = entity.TicketPayment{
			SessionID: paymentSessionID,
		}
		err := dbutil.PossibleRawsAffectedError(
			tx.Where("session_id = ?", paymentSessionID).First(&session),
		)
		if err != nil {
			return err
		}
		return dbutil.PossibleRawsAffectedError(tx.Unscoped().Delete(&entity.Ticket{ID: session.TicketID}))

	})
}

func (ds *ticketMySQL) Create(ctx context.Context, ticket *entity.Ticket) error {

	return dbutil.PossibleCreateError(ds.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(ticket), "ticket-data")
}

func (ds *ticketMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Ticket, error) {
	var ticket = entity.Ticket{ID: id}
	return ticket, dbutil.PossibleFirstError(ds.db.WithContext(ctx).Preload(clause.Associations).First(&ticket), "non-existing-ticket")
}

func (ds *ticketMySQL) GetTickets(ctx context.Context, pagination dbutil.Pagination) ([]entity.Ticket, []entity.Connection, int, error, bool) {
	// 	if len(pagination.Condition.Where) > 0{
	// pagination.Condition.Where += " && "
	// 	}
	// 	pagination.Condition.Where += "(SELECT )"
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

// func (ds *ticketMySQL) Delete(ctx context.Context, id uuid.UUID) error {
// 	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Delete(&entity.Ticket{}, id), "non-existing-ticket")
// }

// func (ds *ticketMySQL) ChangeConnection(ctx context.Context, id, connectionID uuid.UUID) error {
// 	return dbutil.PossibleForeignKeyError(ds.db.WithContext(ctx).Where("id = ?", id).Update("connection_id", connectionID), "non-existing-ticket", "non-existing-connection", "invalid-id")
// }

// func (ds *ticketMySQL) ChangePassenger(ctx context.Context, id uuid.UUID, passengerID uuid.UUID) error {
// 	return dbutil.PossibleForeignKeyError(ds.db.WithContext(ctx).Where("id = ?", id).Update("passenger_id", passengerID), "non-existing-ticket", "non-existing-passenger", "invalid-id")
// }

// func (ds *ticketMySQL) Complete(ctx context.Context, id uuid.UUID) error {
// 	return dbutil.PossibleRawsAffectedError(ds.db.WithContext(ctx).Table("").Where("id = ?", id).Update("completed_at", time.Now().UTC()), "non-existing-ticket")
// }

func NewTicket(db *gorm.DB) Ticket {
	return &ticketMySQL{db}
}
