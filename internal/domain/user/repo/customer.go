package repo

import (
	"context"
	"maryan_api/internal/entity"
	dataStore "maryan_api/internal/infrastructure/persistence"
	objectvalue "maryan_api/internal/valueobject"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomerRepo embeds UserRepo and adds Customer-specific methods.
type CustomerRepo interface {
	UserRepo

	Create(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	EmailExists(ctx context.Context, email string) (bool, error)

	StartEmailVerification(ctx context.Context, session objectvalue.EmailVerificationSession) (uuid.UUID, error)
	EmailVerificationSession(ctx context.Context, sessionID uuid.UUID) (objectvalue.EmailVerificationSession, error)
	CompleteEmailVerification(ctx context.Context, sessionID uuid.UUID) error

	StartNumberVerification(ctx context.Context, session objectvalue.NumberVerificationSession) (uuid.UUID, error)
	NumberVerificationSession(ctx context.Context, sessionID uuid.UUID) (objectvalue.NumberVerificationSession, error)
	CompleteNumberVerification(ctx context.Context, sessionID uuid.UUID) error
}

// MySQL implementation of CustomerRepo
type customerRepo struct {
	UserRepo
	store dataStore.Customer
}

func (cr *customerRepo) Create(ctx context.Context, u *entity.User) error {
	return cr.store.Create(ctx, u)
}

func (cr *customerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return cr.store.Delete(ctx, id)
}

func (cr *customerRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	return cr.store.EmailExists(ctx, email)
}

func (cr *customerRepo) UserExists(ctx context.Context, email string) (uuid.UUID, bool, error) {
	return cr.store.UserExists(ctx, email)
}

func (cr *customerRepo) StartEmailVerification(ctx context.Context, session objectvalue.EmailVerificationSession) (uuid.UUID, error) {
	return cr.store.StartEmailVerification(ctx, session)
}

func (cr *customerRepo) EmailVerificationSession(ctx context.Context, sessionID uuid.UUID) (objectvalue.EmailVerificationSession, error) {
	return cr.store.EmailVerificationSession(ctx, sessionID)
}

func (cr *customerRepo) CompleteEmailVerification(ctx context.Context, sessionID uuid.UUID) error {
	return cr.store.CompleteEmailVerification(ctx, sessionID)
}

func (cr *customerRepo) StartNumberVerification(ctx context.Context, session objectvalue.NumberVerificationSession) (uuid.UUID, error) {
	return cr.store.StartNumberVerification(ctx, session)
}

func (cr *customerRepo) NumberVerificationSession(ctx context.Context, sessionID uuid.UUID) (objectvalue.NumberVerificationSession, error) {
	return cr.store.NumberVerificationSession(ctx, sessionID)
}

func (cr *customerRepo) CompleteNumberVerification(ctx context.Context, sessionID uuid.UUID) error {
	return cr.store.CompleteNumberVerification(ctx, sessionID)
}

// Constructor function
func NewCustomerRepo(db *gorm.DB) CustomerRepo {
	return &customerRepo{
		UserRepo: NewUserRepo(db),
		store:    dataStore.NewCustomer(db),
	}
}
