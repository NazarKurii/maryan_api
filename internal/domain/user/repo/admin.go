package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"

	"gorm.io/gorm"
)

type AdminRepo interface {
	UserRepo
	Users(ctx context.Context, pagination dbutil.CondtionPagination) ([]entity.User, hypermedia.Links, error)
	NewUser(ctx context.Context, user *entity.User) error
}

type adminRepo struct {
	UserRepo
	store dataStore.AdminDataStore
}

func (ar *adminRepo) Users(ctx context.Context, pagination dbutil.CondtionPagination) ([]entity.User, hypermedia.Links, error) {
	return ar.store.Users(ctx, pagination)
}

func (ar *adminRepo) NewUser(ctx context.Context, user *entity.User) error {
	return ar.store.NewUser(ctx, user)
}

func (ar *adminRepo) SetEmployeeAvailability(ctx context.Context, user *entity.User) error {
	return ar.store.NewUser(ctx, user)
}

// Constructor function
func NewAdminRepo(db *gorm.DB) AdminRepo {
	return &adminRepo{
		UserRepo: NewUserRepo(db),
		store:    dataStore.NewAdmin(db),
	}
}
