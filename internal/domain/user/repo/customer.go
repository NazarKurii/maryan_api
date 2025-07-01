package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// customerRepo embeds userRepo and defines additional Customer functionality.
type CustomerRepo interface {
	UserRepo

	Create(user *entity.User, ctx context.Context) error
	Delete(id uuid.UUID, ctx context.Context) error
	EmailExists(email string, ctx context.Context) (bool, error)

	StartEmailVerification(session entity.EmailVerificationSession, ctx context.Context) (uuid.UUID, error)
	EmailVerificationSession(sessionID uuid.UUID, ctx context.Context) (entity.EmailVerificationSession, error)
	CompleteEmailVerification(sessionID uuid.UUID, ctx context.Context) error

	StartNumberVerification(session entity.NumberVerificationSession, ctx context.Context) (uuid.UUID, error)
	NumberVerificationSession(sessionID uuid.UUID, ctx context.Context) (entity.NumberVerificationSession, error)
	CompleteNumberVerification(sessionID uuid.UUID, ctx context.Context) error
}

// MySQL customer repo implementation
type customerRepo struct {
	UserRepo
	store dataStore.CustomerDataStore
}

func (cr *customerRepo) Create(u *entity.User, ctx context.Context) error {
	return cr.store.Create(u, ctx)
}

func (cr *customerRepo) Delete(id uuid.UUID, ctx context.Context) error {
	return cr.store.Delete(id, ctx)
}

func (cr *customerRepo) EmailExists(email string, ctx context.Context) (bool, error) {
	return cr.store.EmailExists(email, ctx)
}

func (cr *customerRepo) UserExists(email string, ctx context.Context) (uuid.UUID, bool, error) {
	return cr.store.UserExists(email, ctx)
}

func (cr *customerRepo) StartEmailVerification(session entity.EmailVerificationSession, ctx context.Context) (uuid.UUID, error) {
	return cr.store.StartEmailVerification(session, ctx)
}

func (cr *customerRepo) EmailVerificationSession(sessionID uuid.UUID, ctx context.Context) (entity.EmailVerificationSession, error) {
	return cr.store.EmailVerificationSession(sessionID, ctx)
}

func (cr *customerRepo) CompleteEmailVerification(sessionID uuid.UUID, ctx context.Context) error {
	return cr.store.CompleteEmailVerification(sessionID, ctx)
}

func (cr *customerRepo) StartNumberVerification(session entity.NumberVerificationSession, ctx context.Context) (uuid.UUID, error) {
	return cr.store.StartNumberVerification(session, ctx)
}
func (cr *customerRepo) NumberVerificationSession(sessionID uuid.UUID, ctx context.Context) (entity.NumberVerificationSession, error) {
	return cr.store.NumberVerificationSession(sessionID, ctx)
}

func (cr *customerRepo) CompleteNumberVerification(sessionID uuid.UUID, ctx context.Context) error {
	return cr.store.CompleteNumberVerification(sessionID, ctx)
}

// Declaration function
func NewCustomerRepo(db *gorm.DB) CustomerRepo {
	return &customerRepo{
		newUserRepo(db),
		dataStore.NewCustomerDataStore(db),
	}
}
