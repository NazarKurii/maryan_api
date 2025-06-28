package user

import (
	"maryan_api/pkg/dbutil"
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userRepo defines basic login methods and the getByID method, that available for all users.
type userRepo interface {
	//----------Basic User Manipulations ---------
	getByID(id uuid.UUID) (User, error) //OK
	login(email string) (uuid.UUID, string, error)
	userExists(email string) (uuid.UUID, bool, error)
	emailExists(email string) (bool, error)

	create(user *User) error
	delete(id uuid.UUID) error

	//-------------Aditional Methods--------------
	database() *gorm.DB
	repo() *userRepoMySQL
}

// customerRepo embeds userRepo and defines additional Customer functionality.
type customerRepo interface {
	userRepo

	startEmailVerification(code, email string) (uuid.UUID, error)
	emailVerificationSession(sessionID uuid.UUID) (EmailVerificationSession, error)
	completeEmailVerification(sessionID uuid.UUID, email string) (uuid.UUID, error)
	verifiedEmail(sessionID uuid.UUID) (EmailVerifiedSession, error)
	useVerifiedEmail(sessionID uuid.UUID) error

	startNumberVerification(code, number string) (uuid.UUID, error)
	numberVerificationSession(sessionID uuid.UUID) (NumberVerificationSession, error)
	completeNumberVerification(sessionID uuid.UUID, number string) (uuid.UUID, error)
	verifiedNumber(sessionID uuid.UUID) (NumberVerifiedSession, error)
	useVerifiedNumber(sessionID uuid.UUID) error
}

// MYSQL IMPLEMENTATION
type userRepoMySQL struct {
	db *gorm.DB
}

func (mscr *userRepoMySQL) repo() *userRepoMySQL {
	return mscr
}

func (mscr *userRepoMySQL) database() *gorm.DB {
	return mscr.db
}

func (mscr *userRepoMySQL) getByID(id uuid.UUID) (User, error) {
	var user = User{ID: id}

	return user, dbutil.PossibleFirstError(
		mscr.db.First(&user),
		"non-existing-user",
	)
}

func (mscr *userRepoMySQL) login(email string) (uuid.UUID, string, error) {
	var user User

	return user.ID, user.Password, dbutil.PossibleFirstError(
		mscr.db.Select("id", "password").Where("email = ?", email).First(&user),
		"non-existing-user",
	)
}

func (mscr *userRepoMySQL) userExists(email string, id uuid.UUID) (bool, error) {
	var exists bool
	err := mscr.db.Raw("SELECT EXISTS(SELECT 1 FROM users WHERe email = ? AND id = ?)", email, id).Scan(&exists).Error
	if err != nil {
		return false, rfc7807.DB(err.Error())
	}

	return exists, nil
}

type customerRepoMySQL struct {
	userRepoMySQL
}

func (mscr *customerRepoMySQL) create(u *User) error {
	return dbutil.PossibleCreateError(mscr.db.Create(u), "user-credentials-validation")
}

func (mscr *customerRepoMySQL) delete(id uuid.UUID) error {
	return dbutil.PossibleRawsAffectedError(
		mscr.db.Delete(&User{ID: id}),
		"non-existing-user",
	)
}

func (mscr *customerRepoMySQL) emailExists(email string) (bool, error) {
	var exists bool
	result := mscr.db.Raw("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?) ", email).Scan(&exists)
	if result.Error != nil {
		return false, rfc7807.DB(result.Error.Error())
	}

	return exists, nil
}

func (mscr *customerRepoMySQL) userExists(email string) (uuid.UUID, bool, error) {
	var user User
	err := mscr.db.Select("id").Where("email = ?", email).Find(&user).Error
	if err != nil {
		return uuid.Nil, false, rfc7807.DB(err.Error())
	}

	return user.ID, user.ID != uuid.Nil, nil
}

func (mscr *customerRepoMySQL) startEmailVerification(code, email string) (uuid.UUID, error) {
	var session = EmailVerificationSession{
		uuid.New(),
		code,
		email,
		time.Now().Add(time.Minute + 10),
	}

	return session.ID, dbutil.PossibleCreateError(mscr.db.Create(session), "invalid-email-verification-session-data")
}

func (mscr *customerRepoMySQL) emailVerificationSession(sessionID uuid.UUID) (EmailVerificationSession, error) {
	var session = EmailVerificationSession{ID: sessionID}
	return session, dbutil.PossibleFirstError(mscr.db.First(&session), "non-existing-email-verification-session")
}

func (mscr *customerRepoMySQL) completeEmailVerification(sessionID uuid.UUID, email string) (uuid.UUID, error) {
	if err := dbutil.PossibleDeleteError(
		mscr.db.Where("email = ?", email).Delete(&EmailVerificationSession{ID: sessionID}),
		"non-existing-email-verification-session"); err != nil {
		return uuid.Nil, err
	}

	sessionID = uuid.New()
	return sessionID, dbutil.PossibleCreateError(
		mscr.db.Create(&EmailVerifiedSession{sessionID, email, time.Now().Add(time.Minute * 10)}),
		"invalid-verified-email-session-data",
	)
}

func (mscr *customerRepoMySQL) verifiedEmail(sessionID uuid.UUID) (EmailVerifiedSession, error) {
	var session = EmailVerifiedSession{ID: sessionID}
	return session, dbutil.PossibleFirstError(mscr.db.First(&session), "non-existing-verified-email-session")

}

func (mscr *customerRepoMySQL) useVerifiedEmail(sessionID uuid.UUID) error {
	return dbutil.PossibleDeleteError(
		mscr.db.Delete(&EmailVerifiedSession{ID: sessionID}),
		"non-existing-number-verification-session",
	)
}

func (mscr *customerRepoMySQL) startNumberVerification(code, number string) (uuid.UUID, error) {
	var session = NumberVerificationSession{
		uuid.New(),
		code,
		number,
		time.Now().Add(time.Minute + 10),
	}

	return session.ID, dbutil.PossibleCreateError(mscr.db.Create(session), "invalid-number-verification-session-data")
}

func (mscr *customerRepoMySQL) numberVerificationSession(sessionID uuid.UUID) (NumberVerificationSession, error) {
	var session = NumberVerificationSession{ID: sessionID}
	return session, dbutil.PossibleFirstError(mscr.db.First(&session), "non-existing-number-verification-session")
}

func (mscr *customerRepoMySQL) completeNumberVerification(sessionID uuid.UUID, number string) (uuid.UUID, error) {
	if err := dbutil.PossibleDeleteError(
		mscr.db.Where("number = ?", number).Delete(&NumberVerificationSession{ID: sessionID}),
		"non-existing-number-verification-session"); err != nil {
		return uuid.Nil, err
	}

	sessionID = uuid.New()
	return sessionID, dbutil.PossibleCreateError(
		mscr.db.Create(&NumberVerifiedSession{sessionID, number, time.Now().Add(time.Minute * 10)}),
		"invalid-verified-number-session-data",
	)
}

func (mscr *customerRepoMySQL) verifiedNumber(sessionID uuid.UUID) (NumberVerifiedSession, error) {
	var session = NumberVerifiedSession{ID: sessionID}
	return session, dbutil.PossibleFirstError(mscr.db.First(&session), "non-existing-verified-number-session")

}

func (mscr *customerRepoMySQL) useVerifiedNumber(sessionID uuid.UUID) error {
	return dbutil.PossibleDeleteError(
		mscr.db.Delete(&NumberVerifiedSession{ID: sessionID}),
		"non-existing-number-verification-session",
	)

}

//Declaration functions

func newCustomerRepoMySQL(db *gorm.DB) customerRepo {
	return &customerRepoMySQL{userRepoMySQL{db}}
}
