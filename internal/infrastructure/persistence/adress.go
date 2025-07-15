package dataStore

import (
	"context"
	"fmt"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	rfc7807 "maryan_api/pkg/problem"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Address interface {
	Create(ctx context.Context, a *entity.Address) error
	Update(ctx context.Context, a *entity.Address) error
	ForseDelete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Status(ctx context.Context, id uuid.UUID) (exists bool, usedByTicket bool, err error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Address, error)
	GetAddresses(ctx context.Context, p dbutil.Pagination) ([]entity.Address, int, error, bool)
}

type addressMySQL struct {
	db *gorm.DB
}

func (ams *addressMySQL) Create(ctx context.Context, address *entity.Address) error {
	return dbutil.PossibleCreateError(
		ams.db.WithContext(ctx).Create(address),
		"invalid-address-data",
	)
}

func (ams *addressMySQL) Update(ctx context.Context, address *entity.Address) error {
	fmt.Println(*address)
	return dbutil.PossibleRawsAffectedError(
		ams.db.WithContext(ctx).Updates(*address),
		"invalid-address-data",
	)
}

func (ams *addressMySQL) ForseDelete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		ams.db.WithContext(ctx).Unscoped().Delete(&entity.Address{ID: id}),
		"invalid-address-data",
	)
}

func (ams *addressMySQL) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		ams.db.WithContext(ctx).Delete(&entity.Address{ID: id}),
		"invalid-address-data",
	)
}

func (ams *addressMySQL) Status(ctx context.Context, id uuid.UUID) (bool, bool, error) {
	var exists bool
	var usedByTicket bool

	if err := ams.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM addresses WHERE id = ?)", id).
		Scan(&exists).Error; err != nil {
		return false, false, rfc7807.DB(err.Error())
	}

	if err := ams.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM tickets WHERE pick_up_address_id = ? OR drop_off_address_id = ?)", id, id).
		Scan(&usedByTicket).Error; err != nil {
		return false, false, rfc7807.DB(err.Error())
	}

	return exists, usedByTicket, nil
}

func (ams *addressMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Address, error) {
	address := entity.Address{ID: id}
	return address, dbutil.PossibleFirstError(
		ams.db.WithContext(ctx).First(&address),
		"non-existing-address",
	)
}

func (ams *addressMySQL) GetAddresses(ctx context.Context, p dbutil.Pagination) ([]entity.Address, int, error, bool) {
	return dbutil.Paginate[entity.Address](ctx, ams.db, p)
}

func NewAddress(db *gorm.DB) Address {
	return &addressMySQL{db: db}
}
