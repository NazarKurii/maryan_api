package repo

import (
	"context"

	"github.com/nazarkurii/marshrutka_api/internal/entity"
	dataStore "github.com/nazarkurii/marshrutka_api/internal/infrastructure/persistence"
	"github.com/nazarkurii/marshrutka_api/pkg/dbutil"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
)

type Ticket interface {
	GetConnectionByID(ctx context.Context, id uuid.UUID, passengersNumber int) (entity.Connection, []uuid.UUID, error)
	CreateAdress(ctx context.Context, a *entity.Address) error
	CreatePassenger(ctx context.Context, p *entity.Passenger) error
	SaveTicket(ctx context.Context, ticket *entity.Ticket) error
	DeleteTickets(ctx context.Context, paymentSessionID string) error
	CreatePassengerStops(ctx context.Context, paymentSessionID string) error
	PaymentSucceeded(ctx context.Context, paymentSessionID string) error
	RemovePassengerStops(ctx context.Context, paymentSessionID string) error
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

func (r *ticketRepo) GetConnectionByID(ctx context.Context, id uuid.UUID, passengersNumber int) (entity.Connection, []uuid.UUID, error) {
	return r.connection.GetByID(ctx, id, passengersNumber)
}

func (r *ticketRepo) DeleteTickets(ctx context.Context, paymentSessionID string) error {
	return r.ticket.DeleteTickets(ctx, paymentSessionID)
}

func (r *ticketRepo) PaymentSucceeded(ctx context.Context, paymentSessionID string) error {
	return r.ticket.PaymentSucceeded(ctx, paymentSessionID)
}

func (r *ticketRepo) CreatePassengerStops(ctx context.Context, paymentSessionID string) error {
	return r.ticket.CreatePassengerStops(ctx, paymentSessionID)
}

func (r *ticketRepo) RemovePassengerStops(ctx context.Context, paymentSessionID string) error {
	return r.ticket.RemovePassengerStops(ctx, paymentSessionID)
}

func (r *ticketRepo) CreateAdress(ctx context.Context, a *entity.Address) error {
	return r.adress.Create(ctx, a)
}

func (r *ticketRepo) CreatePassenger(ctx context.Context, p *entity.Passenger) error {
	return r.passenger.Create(ctx, p)
}

func (r *ticketRepo) SaveTicket(ctx context.Context, ticket *entity.Ticket) error {
	return r.ticket.Create(ctx, ticket)
}

func NewTicketRepo(db *gorm.DB) Ticket {
	return &ticketRepo{
		dataStore.NewTicket(db), dataStore.NewAddress(db), dataStore.NewPassenger(db), dataStore.NewConnection(db),
	}
}
