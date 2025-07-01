package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"

	"gorm.io/gorm"
)

type AdminRepo interface {
	UserRepo
	Users(pageNumber, pageSize int, orderBy string, roles []string, ctx context.Context) ([]entity.User, int, error)
}

type adminRepo struct {
	UserRepo
	store dataStore.AdminDataStore
}

func (ar *adminRepo) Users(pageNumber, pageSize int, orderBy string, roles []string, ctx context.Context) ([]entity.User, int, error) {
	return ar.store.Users(pageNumber, pageSize, orderBy, roles, ctx)

}

//Declaration function

func NewAdminRepo(db *gorm.DB) AdminRepo {
	return &adminRepo{
		newUserRepo(db),
		dataStore.NewAdminDataStore(db),
	}
}
