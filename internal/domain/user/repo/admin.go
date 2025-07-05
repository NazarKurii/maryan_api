package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/pagination"

	"gorm.io/gorm"
)

type AdminRepo interface {
	UserRepo
	Users(ctx context.Context, cfg pagination.CfgCondition) ([]entity.User, int, error)
	NewUser(ctx context.Context, user *entity.User) error
}

type adminRepo struct {
	UserRepo
	store dataStore.AdminDataStore
}

func (ar *adminRepo) Users(ctx context.Context, cfg pagination.CfgCondition) ([]entity.User, int, error) {
	return ar.store.Users(ctx, cfg)
}

func (ar *adminRepo) NewUser(ctx context.Context, user *entity.User) error {
	return ar.store.NewUser(ctx, user)
}

// Constructor function
func NewAdminRepo(db *gorm.DB) AdminRepo {
	return &adminRepo{
		UserRepo: NewUserRepo(db),
		store:    dataStore.NewAdmin(db),
	}
}
