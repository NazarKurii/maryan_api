package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	rfc7807 "maryan_api/pkg/problem"
	"math"

	"gorm.io/gorm"
)

type AdminDataStore interface {
	UserDataStore
	Users(pageNumber, pageSize int, orderBy string, roles []string, ctx context.Context) ([]entity.User, int, error)
}

type adminMySQL struct {
	userMySQL
}

func (ar *adminMySQL) Users(pageNumber, pageSize int, orderBy string, roles []string, ctx context.Context) ([]entity.User, int, error) {
	pageNumber--
	var users []entity.User

	err := dbutil.PossibleRawsAffectedError(ar.db.WithContext(ctx).Limit(pageSize).Offset(pageNumber*pageSize).Find(&users), "non-existing-page")
	if err != nil {
		return nil, 0, err
	}

	var totalUsers int64

	err = ar.db.Model(&entity.User{}).Count(&totalUsers).Error
	if err != nil || totalUsers == 0 {
		return nil, 0, rfc7807.DB("Could not count users.")
	}

	return users, int(math.Ceil(float64(totalUsers) / float64(pageSize))), nil

}

//Declaration function

func NewAdminDataStore(db *gorm.DB) AdminDataStore {
	return &adminMySQL{userMySQL{db}}
}
