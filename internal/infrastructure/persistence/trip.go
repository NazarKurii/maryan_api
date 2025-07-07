package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Trip interface {
	Create(ctx context.Context, trip *entity.Trip) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Trip, error)
	GetTrips(ctx context.Context, p dbutil.CondtionPagination) ([]entity.Trip, hypermedia.Links, error)
	ChangeDriver(ctx context.Context, id, driverID uuid.UUID) error
	ChangeBus(ctx context.Context, id, busID uuid.UUID) error
	ChangeDepartureTime(ctx context.Context, id uuid.UUID, time time.Time) error
	Cancel(ctx context.Context, id uuid.UUID) error
	Start(ctx context.Context, id uuid.UUID) error
	Finish(ctx context.Context, id uuid.UUID) error
	MakeSold(ctx context.Context, id uuid.UUID) error
}

type Stop interface {
	Create(ctx context.Context, stop *entity.Stop) error
	Delete(ctx context.Context, id uuid.UUID) error
	Complete(ctx context.Context, id uuid.UUID) error
	MakeMissed(ctx context.Context, id uuid.UUID) error
}

type Ticket interface {
	Create(ctx context.Context, ticket *entity.Ticket) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Ticket, error)
	GetTickets(ctx context.Context, p dbutil.CondtionPagination) ([]entity.Ticket, hypermedia.Links, error)
	Cancel(ctx context.Context, id uuid.UUID) error
	Complete(ctx context.Context, id uuid.UUID) error
}

type Refaund interface {
	Create(ctx context.Context, refaund *entity.Refaund) error
	Complete(ctx context.Context, id uuid.UUID) error
}

type tripMySQL struct {
	db *gorm.DB
}

func (t *tripMySQL) Create(ctx context.Context, trip *entity.Trip) error {
	return dbutil.PossibleCreateError(t.db.WithContext(ctx).Create(trip), "trip-data")
}

func (t *tripMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Trip, error) {
	var trip = entity.Trip{ID: id}
	return trip, dbutil.PossibleFirstError(t.db.WithContext(ctx).
		Preload("Bus").
		Preload("Bus.Seats").
		Preload("Bus.Structure").
		Preload("Bus.Structure.Positions").
		Preload("Driver").
		Preload("Stops").
		Preload("Stops.Ticket").
		Preload("Stops.Ticket.PickUpAdress").
		Preload("Stops.Ticket.DropOffAdress").
		Preload("Stops.Ticket.Passenger").
		First(&trip), "non-existing-trip")
}

func (t *tripMySQL) GetTrips(ctx context.Context, p dbutil.CondtionPagination) ([]entity.Trip, hypermedia.Links, error) {
	return dbutil.PaginateWithCondition[entity.Trip](ctx, t.db, p,
		"Bus",
		"Bus.Seats",
		"Bus.Structure",
		"Bus.Structure.Positions",
		"Driver",
		"Stops",
		"Stops.Ticket",
		"Stops.Ticket.PickUpAdress",
		"Stops.Ticket.DropOffAdress",
		"Stops.Ticket.Passenger",
	)
}

func (t *tripMySQL) ChangeDriver(ctx context.Context, id, driverID uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(t.db.Model(&entity.Trip{}).Where("id = ?", id).Update("driver_id", driverID), "non-existing-trip")
}

func (t *tripMySQL) ChangeBus(ctx context.Context, id, busID uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(t.db.Model(&entity.Trip{}).Where("id = ?", id).Update("bus_id", busID), "non-existing-trip")
}

func (t *tripMySQL) ChangeDepartureTime(ctx context.Context, id uuid.UUID, time time.Time) error {
	return dbutil.PossibleRawsAffectedError(t.db.Model(&entity.Trip{}).Where("id = ?", id).Update("departure_time", time), "non-existing-trip")
}

func (t *tripMySQL) Cancel(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(t.db.Model(&entity.Trip{}).Where("id = ?", id).Update("canceled_at", time.Now()), "non-existing-trip")
}

func (t *tripMySQL) Start(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(t.db.Model(&entity.Trip{}).Where("id = ?", id).Update("started_at", time.Now()), "non-existing-trip")
}

func (t *tripMySQL) Finish(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(t.db.Model(&entity.Trip{}).Where("id = ?", id).Update("finished_at", time.Now()), "non-existing-trip")
}

func (t *tripMySQL) MakeSold(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(t.db.Model(&entity.Trip{}).Where("id = ?", id).Update("sold_at", time.Now()), "non-existing-trip")
}

type stopMySQL struct {
	db *gorm.DB
}

func (s *stopMySQL) Create(ctx context.Context, stop *entity.Stop) error {
	return dbutil.PossibleCreateError(s.db.WithContext(ctx).Create(stop), "stop-data")
}

func (s *stopMySQL) Delete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(s.db.WithContext(ctx).Delete(&entity.Stop{}, id), "stop-data")
}

func (s *stopMySQL) Complete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(s.db.WithContext(ctx).Model(&entity.Stop{}).Where("id = ?", id).Update("status", entity.CompletedStopStatus), "non-existing-stop")
}

func (s *stopMySQL) MakeMissed(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(s.db.WithContext(ctx).Model(&entity.Stop{}).Where("id = ?", id).Update("status", entity.MissedStopStatus), "non-existing-stop")
}

type ticketMySQL struct {
	db *gorm.DB
}

func (t *ticketMySQL) Create(ctx context.Context, ticket *entity.Ticket) error {
	return dbutil.PossibleCreateError(t.db.Create(ticket), "ticket-data")
}

func (t *ticketMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Ticket, error) {
	var ticket = entity.Ticket{ID: id}
	return ticket, dbutil.PossibleFirstError(t.db.First(&ticket), "ticket-data")
}

func (t *ticketMySQL) GetTickets(ctx context.Context, p dbutil.CondtionPagination) ([]entity.Ticket, hypermedia.Links, error) {
	return dbutil.PaginateWithCondition[entity.Ticket](ctx, t.db, p, "Passenger", "PickUpAdress", "DropOffAdress", "Payment")
}

func (t *ticketMySQL) Cancel(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(t.db.Model(&entity.Ticket{}).Where("id = ?", id).Update("canceled_at", time.Now()), "non-existing-ticket")
}

func (t *ticketMySQL) Complete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(t.db.Model(&entity.Ticket{}).Where("id = ?", id).Update("completed_at", time.Now()), "non-existing-ticket")
}

type refaundMySQL struct {
	db *gorm.DB
}

func (r *refaundMySQL) Create(ctx context.Context, refaund *entity.Refaund) error {
	return dbutil.PossibleCreateError(r.db.Create(refaund), "ticket-data")
}

func (r *refaundMySQL) Complete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(r.db.Model(&entity.Refaund{}).Where("id = ?", id).Update("completed_at", time.Now()), "non-existing-refaund")
}

func NewTicket(db *gorm.DB) Ticket {
	return &ticketMySQL{db}
}

func NewTrip(db *gorm.DB) Trip {
	return &tripMySQL{db}
}

func NewStop(db *gorm.DB) Stop {
	return &stopMySQL{db}
}

func NewRefaund(db *gorm.DB) Refaund {
	return &refaundMySQL{db}
}
