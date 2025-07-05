package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Adress interface {
	Create(ctx context.Context, p *entity.Adress) error
	Update(ctx context.Context, p *entity.Adress) error
	ForseDelete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Status(ctx context.Context, id uuid.UUID) (exists bool, usedByTicket bool, err error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Adress, error)
	GetAdresses(ctx context.Context, cfg pagination.CfgCondition) ([]entity.Adress, int, error)
}

type adressMySQL struct {
	ds dataStore.Adress
}

func (a *adressMySQL) Create(ctx context.Context, adress *entity.Adress) error {
	return a.ds.Create(ctx, adress)
}

func (a *adressMySQL) Update(ctx context.Context, adress *entity.Adress) error {
	return a.ds.Update(ctx, adress)
}

func (a *adressMySQL) ForseDelete(ctx context.Context, id uuid.UUID) error {
	return a.ds.ForseDelete(ctx, id)
}

func (a *adressMySQL) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return a.ds.SoftDelete(ctx, id)
}

func (a *adressMySQL) Status(ctx context.Context, id uuid.UUID) (bool, bool, error) {
	return a.ds.Status(ctx, id)
}

func (a *adressMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.Adress, error) {
	return a.ds.GetByID(ctx, id)
}

func (a *adressMySQL) GetAdresses(ctx context.Context, cfg pagination.CfgCondition) ([]entity.Adress, int, error) {
	return a.ds.GetAdresses(ctx, cfg)
}

func NewPassengerRepo(db *gorm.DB) Adress {
	return &adressMySQL{dataStore.NewAdress(db)}
}
