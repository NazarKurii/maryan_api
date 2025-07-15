package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/dbutil"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ticket interface {
	GetConnectionByID(ctx context.Context, id uuid.UUID) (entity.Connection, []uuid.UUID, error)
	CreateAdress(ctx context.Context, a *entity.Address) error
	CreatePassenger(ctx context.Context, p *entity.Passenger) error
	SaveTickets(ctx context.Context, tickets []*entity.Ticket) error
	DeleteTickets(ctx context.Context, paymentSessionID string) error
	AddTickets(ctx context.Context, paymentSessionID string) error
	GetTickets(ctx context.Context, pagination dbutil.Pagination) ([]entity.Ticket, []entity.Connection, int, error, bool)
}

type ticketRepo struct {
	ticket     dataStore.Ticket
	adress     dataStore.Address
	passenger  dataStore.Passenger
	connection dataStore.Connection
}

func (r *ticketRepo) GetTickets(ctx context.Context, pagination dbutil.Pagination) ([]entity.Ticket, []entity.Connection, int, error, bool) {
	return r.ticket.GetTickets(ctx, pagination)
}

func (r *ticketRepo) GetConnectionByID(ctx context.Context, id uuid.UUID) (entity.Connection, []uuid.UUID, error) {
	return r.connection.GetByID(ctx, id)
}

func (r *ticketRepo) DeleteTickets(ctx context.Context, paymentSessionID string) error {
	return r.ticket.DeleteTickets(ctx, paymentSessionID)
}

func (r *ticketRepo) AddTickets(ctx context.Context, paymentSessionID string) error {
	return r.ticket.AddTickets(ctx, paymentSessionID)
}

func (r *ticketRepo) CreateAdress(ctx context.Context, a *entity.Address) error {
	return r.adress.Create(ctx, a)
}

func (r *ticketRepo) CreatePassenger(ctx context.Context, p *entity.Passenger) error {
	return r.passenger.Create(ctx, p)
}

func (r *ticketRepo) SaveTickets(ctx context.Context, tickets []*entity.Ticket) error {
	return r.ticket.Create(ctx, tickets)
}

func NewTicketRepo(db *gorm.DB) Ticket {
	return &ticketRepo{
		dataStore.NewTicket(db), dataStore.NewAddress(db), dataStore.NewPassenger(db), dataStore.NewConnection(db),
	}
}
