package repo

import (
	"context"

	"github.com/nazarkurii/marshrutka_api/internal/entity"
	dataStore "github.com/nazarkurii/marshrutka_api/internal/infrastructure/persistence"
	"github.com/nazarkurii/marshrutka_api/pkg/dbutil"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
)

type Parcel interface {
	GetConnectionsByMonth(ctx context.Context, fromCountry, toCountry uuid.UUID, monthNumber, year int) ([]entity.Connection, error)
	CreateParcelStops(ctx context.Context, paymentSessionID string) error
	Create(ctx context.Context, parcel *entity.Parcel) error
	GetParcels(ctx context.Context, pagination dbutil.Pagination) ([]entity.Parcel, []entity.Connection, int, error, bool)
	DeleteParcels(ctx context.Context, paymentSessionID string) error
	RemoveParcelStops(ctx context.Context, paymentSessionID string) error
	PaymentSucceeded(ctx context.Context, paymentSessionID string) error
	GetConnectionByID(ctx context.Context, id uuid.UUID) (entity.Connection, error)
}

type parcelRepo struct {
	parcel     dataStore.Parsel
	connection dataStore.Connection
}

func (r *parcelRepo) GetConnectionByID(ctx context.Context, id uuid.UUID) (entity.Connection, error) {
	connection, _, err := r.connection.GetByID(ctx, id, 0)
	return connection, err
}

func (r *parcelRepo) GetConnectionsByMonth(ctx context.Context, fromCountry, toCountry uuid.UUID, monthNumber, year int) ([]entity.Connection, error) {
	return r.parcel.GetConnectionsByMonth(ctx, fromCountry, toCountry, monthNumber, year)
}

func (r *parcelRepo) Create(ctx context.Context, parcel *entity.Parcel) error {
	return r.parcel.Create(ctx, parcel)
}

func (r *parcelRepo) CreateParcelStops(ctx context.Context, paymentSessionID string) error {
	return r.parcel.CreateParcelStops(ctx, paymentSessionID)
}

func (r *parcelRepo) GetParcels(ctx context.Context, pagination dbutil.Pagination) ([]entity.Parcel, []entity.Connection, int, error, bool) {
	return r.parcel.GetParcels(ctx, pagination)
}

func (r *parcelRepo) DeleteParcels(ctx context.Context, paymentSessionID string) error {
	return r.parcel.DeleteParcels(ctx, paymentSessionID)
}

func (r *parcelRepo) PaymentSucceeded(ctx context.Context, paymentSessionID string) error {
	return r.parcel.PaymentSucceeded(ctx, paymentSessionID)
}

func (r *parcelRepo) RemoveParcelStops(ctx context.Context, paymentSessionID string) error {
	return r.parcel.RemoveParcelStops(ctx, paymentSessionID)
}

func NewParcelRepo(db *gorm.DB) Parcel {
	return &parcelRepo{
		dataStore.NewParsel(db), dataStore.NewConnection(db),
	}
}
