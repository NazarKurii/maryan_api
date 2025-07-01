package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	rfc7807 "maryan_api/pkg/problem"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// customerRepo embeds userRepo and defines additional Customer functionality.
type CustomerDataStore interface {
	UserDataStore

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
type customerMySQL struct {
	userMySQL
}

func (cr *customerMySQL) Create(u *entity.User, ctx context.Context) error {
	return dbutil.PossibleCreateError(cr.db.WithContext(ctx).Create(u), "user-credentials-validation")
}

func (cr *customerMySQL) Delete(id uuid.UUID, ctx context.Context) error {
	return dbutil.PossibleRawsAffectedError(
		cr.db.WithContext(ctx).Delete(&entity.User{ID: id}),
		"non-existing-user",
	)
}

func (cr *customerMySQL) EmailExists(email string, ctx context.Context) (bool, error) {
	var exists bool
	result := cr.db.WithContext(ctx).Raw("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?) ", email).Scan(&exists)
	if result.Error != nil {
		return false, rfc7807.DB(result.Error.Error())
	}

	return exists, nil
}

func (cr *customerMySQL) UserExists(email string, ctx context.Context) (uuid.UUID, bool, error) {
	var user entity.User
	err := cr.db.WithContext(ctx).Select("id").Where("email = ?", email).Find(&user).Error
	if err != nil {
		return uuid.Nil, false, rfc7807.DB(err.Error())
	}

	return user.ID, user.ID != uuid.Nil, nil
}

func (cr *customerMySQL) StartEmailVerification(session entity.EmailVerificationSession, ctx context.Context) (uuid.UUID, error) {

	return session.ID, dbutil.PossibleCreateError(cr.db.WithContext(ctx).Create(session), "invalid-email-verification-session-data")
}

func (cr *customerMySQL) EmailVerificationSession(sessionID uuid.UUID, ctx context.Context) (entity.EmailVerificationSession, error) {
	var session = entity.EmailVerificationSession{ID: sessionID}
	return session, dbutil.PossibleFirstError(cr.db.WithContext(ctx).First(&session), "non-existing-email-verification-session")
}

func (cr *customerMySQL) CompleteEmailVerification(sessionID uuid.UUID, ctx context.Context) error {
	return dbutil.PossibleDeleteError(
		cr.db.WithContext(ctx).Delete(&entity.EmailVerificationSession{ID: sessionID}),
		"non-existing-email-verification-session")

}

func (cr *customerMySQL) StartNumberVerification(session entity.NumberVerificationSession, ctx context.Context) (uuid.UUID, error) {

	return session.ID, dbutil.PossibleCreateError(cr.db.WithContext(ctx).Create(session), "invalid-number-verification-session-data")
}

func (cr *customerMySQL) NumberVerificationSession(sessionID uuid.UUID, ctx context.Context) (entity.NumberVerificationSession, error) {
	var session = entity.NumberVerificationSession{ID: sessionID}
	return session, dbutil.PossibleFirstError(cr.db.WithContext(ctx).First(&session), "non-existing-number-verification-session")
}

func (cr *customerMySQL) CompleteNumberVerification(sessionID uuid.UUID, ctx context.Context) error {
	return dbutil.PossibleDeleteError(
		cr.db.WithContext(ctx).Delete(&entity.NumberVerificationSession{ID: sessionID}),
		"non-existing-number-verification-session")
}

// Declaration function
func NewCustomerDataStore(db *gorm.DB) CustomerDataStore {
	return &customerMySQL{userMySQL{db}}
}
