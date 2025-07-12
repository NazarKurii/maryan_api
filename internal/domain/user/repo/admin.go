package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/pkg/dbutil"
	"time"

	"gorm.io/gorm"
)

type AdminRepo interface {
	UserRepo
	Users(ctx context.Context, pagination dbutil.Pagination) ([]entity.User, int, error)
	NewUser(ctx context.Context, user *entity.User) error
	SetEmployeeAvailability(ctx context.Context, schedule []entity.EmployeeAvailability) error
	GetAvailableUsers(ctx context.Context, dates []time.Time, p dbutil.Pagination) ([]entity.User, int, error)
}

type adminRepo struct {
	UserRepo
	store dataStore.AdminDataStore
}

func (ar *adminRepo) Users(ctx context.Context, pagination dbutil.Pagination) ([]entity.User, int, error) {
	return ar.store.Users(ctx, pagination)
}

func (ar *adminRepo) NewUser(ctx context.Context, user *entity.User) error {
	return ar.store.NewUser(ctx, user)
}

func (ar *adminRepo) SetEmployeeAvailability(ctx context.Context, schedule []entity.EmployeeAvailability) error {
	return ar.store.SetEmployeeAvailability(ctx, schedule)
}

func (ar *adminRepo) GetAvailableUsers(ctx context.Context, dates []time.Time, p dbutil.Pagination) ([]entity.User, int, error) {
	return ar.store.GetAvailableUsers(ctx, dates, p)
}

// Constructor function
func NewAdminRepo(db *gorm.DB) AdminRepo {
	return &adminRepo{
		UserRepo: NewUserRepo(db),
		store:    dataStore.NewAdmin(db),
	}
}
