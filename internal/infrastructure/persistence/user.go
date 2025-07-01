package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userMySQL defines basic login methods and the getByID method, that available for all users.
type UserDataStore interface {
	//----------Basic User Manipulations ---------
	GetByID(id uuid.UUID, ctx context.Context) (entity.User, error) //OK
	Login(email string, ctx context.Context) (uuid.UUID, string, error)
	UserExists(email string, ctx context.Context) (uuid.UUID, bool, error)
}

// MYSQL IMPLEMENTATION
type userMySQL struct {
	db *gorm.DB
}

func (ur *userMySQL) GetByID(id uuid.UUID, ctx context.Context) (entity.User, error) {
	var user = entity.User{ID: id}

	return user, dbutil.PossibleFirstError(
		ur.db.WithContext(ctx).First(&user),
		"non-existing-user",
	)
}

func (ur *userMySQL) Login(email string, ctx context.Context) (uuid.UUID, string, error) {
	var user entity.User

	return user.ID, user.Password, dbutil.PossibleFirstError(
		ur.db.WithContext(ctx).Select("id", "password").Where("email = ?", email).First(&user),
		"non-existing-user",
	)
}

func (ur *userMySQL) UserExists(email string, ctx context.Context) (uuid.UUID, bool, error) {
	var user entity.User

	return user.ID, user.ID != uuid.Nil, dbutil.PossibleRawsAffectedError(
		ur.db.WithContext(ctx).Select("id").Where("email = ?", email).First(&user),
		"non-existing-user",
	)
}

func NewUserDataStore(db *gorm.DB) UserDataStore {
	return &userMySQL{db}
}
