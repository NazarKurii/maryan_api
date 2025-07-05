package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/pagination"
	rfc7807 "maryan_api/pkg/problem"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Adress interface {
	Create(ctx context.Context, a *entity.Adress) error
	Update(ctx context.Context, a *entity.Adress) error
	ForseDelete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Status(ctx context.Context, id uuid.UUID) (exists bool, usedByTicket bool, err error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Adress, error)
	GetAdresses(ctx context.Context, cfg pagination.CfgCondition) ([]entity.Adress, int, error)
}

type adressMySQL struct {
	db *gorm.DB
}

func (ams *adressMySQL) Create(ctx context.Context, adress *entity.Adress) error {
	return dbutil.PossibleCreateError(
		ams.db.WithContext(ctx).Create(adress),
		"invalid-adress-data",
	)
}

func (ams *adressMySQL) Update(ctx context.Context, adress *entity.Adress) error {
	return dbutil.PossibleRawsAffectedError(
		ams.db.WithContext(ctx).Save(adress),
		"invalid-adress-data",
	)
}

func (ams *adressMySQL) ForseDelete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		ams.db.WithContext(ctx).Unscoped().Delete(&entity.Adress{ID: id}),
		"invalid-adress-data",
	)
}

func (ams *adressMySQL) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		ams.db.WithContext(ctx).Delete(&entity.Adress{ID: id}),
		"invalid-adress-data",
	)
}

func (ams *adressMySQL) Status(ctx context.Context, id uuid.UUID) (bool, bool, error) {
	var exists bool
	var usedByTicket bool

	if err := ams.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM adresses WHERE id = ?)", id).
		Scan(&exists).Error; err != nil {
		return false, false, rfc7807.DB(err.Error())
	}

	if err := ams.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM tickets WHERE from_adress_id = ? OR to_adress_id = ?)", id).
		Scan(&usedByTicket).Error; err != nil {
		return false, false, rfc7807.DB(err.Error())
	}

	return exists, usedByTicket, nil
}

func (ams *adressMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Adress, error) {
	adress := entity.Adress{ID: id}
	return adress, dbutil.PossibleFirstError(
		ams.db.WithContext(ctx).First(&adress),
		"non-existing-adress",
	)
}

func (ams *adressMySQL) GetAdresses(ctx context.Context, cfg pagination.CfgCondition) ([]entity.Adress, int, error) {
	return dbutil.PaginationWithCondition[entity.Adress](
		ctx,
		ams.db,
		cfg,
	)
}

func NewAdress(db *gorm.DB) Adress {
	return &adressMySQL{db: db}
}
