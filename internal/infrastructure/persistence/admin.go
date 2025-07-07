package dataStore

import (
	"context"

	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"

	"gorm.io/gorm"
)

type AdminDataStore interface {
	User
	Users(ctx context.Context, p dbutil.CondtionPagination) ([]entity.User, hypermedia.Links, error)
	NewUser(ctx context.Context, user *entity.User) error
}

type adminMySQL struct {
	userMySQL
}

func (ads *adminMySQL) Users(ctx context.Context, p dbutil.CondtionPagination) ([]entity.User, hypermedia.Links, error) {
	return dbutil.PaginateWithCondition[entity.User](ctx, ads.db, p)
}

func (ads *adminMySQL) NewUser(ctx context.Context, user *entity.User) error {
	return dbutil.PossibleCreateError(ads.db.WithContext(ctx).Create(user), "user-credentials-validation")
}

// Declaration function
func NewAdmin(db *gorm.DB) AdminDataStore {
	return &adminMySQL{userMySQL{db}}
}
