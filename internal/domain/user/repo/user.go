package repo

import (
	"context"

	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepo defines basic user methods with context as first param.
type UserRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (entity.User, error)
	Login(ctx context.Context, email string) (uuid.UUID, string, error)
	EmailExists(ctx context.Context, email string) (uuid.UUID, bool, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

// MYSQL implementation
type userRepo struct {
	store dataStore.User
}

func (ur *userRepo) GetByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	return ur.store.GetByID(ctx, id)
}

func (ur *userRepo) Login(ctx context.Context, email string) (uuid.UUID, string, error) {
	return ur.store.Login(ctx, email)
}

func (ur *userRepo) EmailExists(ctx context.Context, email string) (uuid.UUID, bool, error) {
	return ur.store.EmailExists(ctx, email)
}

func (ur *userRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return ur.store.Exists(ctx, id)
}
func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{dataStore.NewUser(db)}
}
