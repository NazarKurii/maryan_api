package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User defines basic user operations.
type User interface {
	GetByID(ctx context.Context, id uuid.UUID) (entity.User, error)
	Login(ctx context.Context, email string) (uuid.UUID, string, error)
	EmailExists(ctx context.Context, email string) (uuid.UUID, bool, error)
	UserExists(ctx context.Context, id uuid.UUID) (bool, error)
}

// MySQL implementation
type userMySQL struct {
	db *gorm.DB
}

func (uds userMySQL) GetByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	user := entity.User{ID: id}
	return user, dbutil.PossibleFirstError(
		uds.db.WithContext(ctx).First(&user),
		"non-existing-user",
	)
}

func (uds userMySQL) Login(ctx context.Context, email string) (uuid.UUID, string, error) {
	var user entity.User
	err := dbutil.PossibleFirstError(
		uds.db.WithContext(ctx).Select("id", "password").Where("email = ?", email).First(&user),
		"non-existing-user",
	)
	return user.ID, user.Password, err
}

func (uds userMySQL) EmailExists(ctx context.Context, email string) (uuid.UUID, bool, error) {
	var user entity.User
	err := dbutil.PossibleRawsAffectedError(
		uds.db.WithContext(ctx).Select("id").Where("email = ?", email).First(&user),
		"non-existing-user",
	)
	return user.ID, user.ID != uuid.Nil, err
}

func (uds userMySQL) UserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := dbutil.PossibleRawsAffectedError(
		uds.db.WithContext(ctx).Select("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", id).Scan(exists),
		"non-existing-user",
	)
	return exists, err
}

// Constructor
func NewUser(db *gorm.DB) User {
	return &userMySQL{db: db}
}
