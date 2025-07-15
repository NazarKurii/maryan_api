package dataStore

import (
	"context"
	"maryan_api/internal/entity"
	objectvalue "maryan_api/internal/valueobject"
	"maryan_api/pkg/dbutil"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// customerRepo embeds userRepo and defines additional Customer functionality.
type Customer interface {
	User

	Create(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	StartEmailVerification(ctx context.Context, session objectvalue.EmailVerificationSession) (uuid.UUID, error)
	EmailVerificationSession(ctx context.Context, sessionID uuid.UUID) (objectvalue.EmailVerificationSession, error)
	CompleteEmailVerification(ctx context.Context, sessionID uuid.UUID) error

	StartNumberVerification(ctx context.Context, session objectvalue.NumberVerificationSession) (uuid.UUID, error)
	NumberVerificationSession(ctx context.Context, sessionID uuid.UUID) (objectvalue.NumberVerificationSession, error)
	CompleteNumberVerification(ctx context.Context, sessionID uuid.UUID) error
	ChangePassword(ctx context.Context, newPassword string, email string) error

	UpdatePersonalInfo(ctx context.Context, firstName, lastName string, dateOfBirth time.Time, id uuid.UUID) error
	UpdateContantInfo(ctx context.Context, email, number string, id uuid.UUID) error
}

// MySQL customer repo implementation
type customerMySQL struct {
	userMySQL
}

func (cds *customerMySQL) Create(ctx context.Context, u *entity.User) error {
	return dbutil.PossibleCreateError(cds.db.WithContext(ctx).Create(u), "user-credentials-validation")
}

func (cds *customerMySQL) Delete(ctx context.Context, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		cds.db.WithContext(ctx).Delete(&entity.User{ID: id}),
		"non-existing-user",
	)
}

func (cds *customerMySQL) StartEmailVerification(ctx context.Context, session objectvalue.EmailVerificationSession) (uuid.UUID, error) {
	return session.ID, dbutil.PossibleCreateError(cds.db.WithContext(ctx).Create(session), "invalid-email-verification-session-data")
}

func (cds *customerMySQL) EmailVerificationSession(ctx context.Context, sessionID uuid.UUID) (objectvalue.EmailVerificationSession, error) {
	var session = objectvalue.EmailVerificationSession{ID: sessionID}
	return session, dbutil.PossibleFirstError(cds.db.WithContext(ctx).First(&session), "non-existing-email-verification-session")
}

func (cds *customerMySQL) CompleteEmailVerification(ctx context.Context, sessionID uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		cds.db.WithContext(ctx).Delete(&objectvalue.EmailVerificationSession{ID: sessionID}),
		"non-existing-email-verification-session")
}

func (cds *customerMySQL) StartNumberVerification(ctx context.Context, session objectvalue.NumberVerificationSession) (uuid.UUID, error) {
	return session.ID, dbutil.PossibleCreateError(cds.db.WithContext(ctx).Create(session), "invalid-number-verification-session-data")
}

func (cds *customerMySQL) NumberVerificationSession(ctx context.Context, sessionID uuid.UUID) (objectvalue.NumberVerificationSession, error) {
	var session = objectvalue.NumberVerificationSession{ID: sessionID}
	return session, dbutil.PossibleFirstError(cds.db.WithContext(ctx).First(&session), "non-existing-number-verification-session")
}

func (cds *customerMySQL) CompleteNumberVerification(ctx context.Context, sessionID uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		cds.db.WithContext(ctx).Delete(&objectvalue.NumberVerificationSession{ID: sessionID}),
		"non-existing-number-verification-session")
}

func (cds *customerMySQL) ChangePassword(ctx context.Context, newPassword string, email string) error {
	return dbutil.PossibleRawsAffectedError(cds.db.Table("users").Where("email = ?", email).Update("password", newPassword), "non-existing-user")
}

func (cds *customerMySQL) UpdatePersonalInfo(ctx context.Context, firstName, lastName string, dateOfBirth time.Time, id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(cds.db.Table("users").Where("id = ?", id).Updates(&entity.User{ID: id, FirstName: firstName, LastName: lastName, DateOfBirth: dateOfBirth}), "non-existing-user")
}

func (cds *customerMySQL) UpdateContantInfo(ctx context.Context, email, phoneNumber string, id uuid.UUID) error {
	return dbutil.ErrDuplicatedKey(cds.db.Table("users").Where("id = ?", id).Updates(&entity.User{
		ID:          id,
		Email:       email,
		PhoneNumber: phoneNumber,
	}), "used-email", "non-existing-user")
}

// Declaration function
func NewCustomer(db *gorm.DB) Customer {
	return &customerMySQL{userMySQL{db}}
}
