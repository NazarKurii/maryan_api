package repo

import (
	"context"

	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userRepo defines basic login methods and the getByID method, that available for all users.
type UserRepo interface {
	//----------Basic User Manipulations ---------
	GetByID(id uuid.UUID, ctx context.Context) (entity.User, error) //OK
	Login(email string, ctx context.Context) (uuid.UUID, string, error)
	UserExists(email string, ctx context.Context) (uuid.UUID, bool, error)
}

// MYSQL IMPLEMENTATION
type userRepo struct {
	store dataStore.UserDataStore
}

func (ur *userRepo) GetByID(id uuid.UUID, ctx context.Context) (entity.User, error) {
	return ur.store.GetByID(id, ctx)
}

func (ur *userRepo) Login(email string, ctx context.Context) (uuid.UUID, string, error) {
	return ur.store.Login(email, ctx)
}

func (ur *userRepo) UserExists(email string, ctx context.Context) (uuid.UUID, bool, error) {
	return ur.store.UserExists(email, ctx)
}

func newUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{dataStore.NewUserDataStore(db)}
}
