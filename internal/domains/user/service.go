package user

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maryan_api/config"
	google "maryan_api/internal/infrastructure/clients/googleoauth"
	"maryan_api/internal/infrastructure/clients/verification"
	"maryan_api/pkg/auth"
	rfc7807 "maryan_api/pkg/problem"
	"maryan_api/pkg/security"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
)

type userService interface {
	//----------Not authentificated------------------
	login(email, password string) (string, error)
	loginJWT(id uuid.UUID, email string) (string, error)
	//------------Authentificated--------------------

	//Aditional functionality
	userService() userService
	secretKey() []byte
}

type customerService interface {
	userService

	//----------Not authentificated------------------
	register(u *User, image *multipart.FileHeader, emailAccessToken, numberAccessToken string) (string, error)
	verifyEmail(email string) (string, bool, error)
	verifyEmailCode(code, token string) (string, error)
	verifyNumber(number string) (string, error)
	verifyNumberCode(number, token string) (string, error)
	googleOAUTH(code string, ctx context.Context, id uuid.UUID) (string, bool, error)

	//------------Authentificated--------------------
	get(id uuid.UUID) (ShortUser, error)
	delete(id uuid.UUID) error
}

type userServiceImpl struct {
	repo userRepo
	role auth.Role
}

func (us *userServiceImpl) userService() userService {
	return us
}

func (us *userServiceImpl) secretKey() []byte {
	return us.role.SecretKey()
}

func (us *userServiceImpl) login(email, password string) (string, error) {
	if !govalidator.IsEmail(email) {
		return "", rfc7807.BadRequest(
			"email-invalid",
			"Invalid Email Error",
			"Provided email contains forbidden characters or is not an email at all.",
		)
	}

	id, passwordHashed, err := us.repo.login(email)
	if err != nil {
		return "", err
	}

	if ok := security.VerifyPassword(password, passwordHashed); !ok {
		return "", rfc7807.Unauthorized(
			"invalid-password",
			"Invalid Password Error",
			"Invalid password for user assosiated with the provided email.",
		)
	}

	return us.role.GenerateToken(email, id)

}

func (us *userServiceImpl) loginJWT(id uuid.UUID, email string) (string, error) {
	if !govalidator.IsEmail(email) {
		return "", rfc7807.BadRequest(
			"email-invalid",
			"Invalid Email Error",
			"Provided email contains forbidden characters or is not an email at all.",
		)
	}

	validID, exists, err := us.repo.userExists(email)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", rfc7807.BadRequest(
			"non-existing-user",
			"Non-existing User Error",
			"There is no user assosiated with the provided email.",
		)
	}

	if id != validID {
		return "", rfc7807.Unauthorized(
			"unauthorized",
			"Unauthorized",
			"Invalid token.",
		)
	}

	token, err := us.role.GenerateToken(email, id)

	return token, err
}

func (us *userServiceImpl) get(id uuid.UUID) (ShortUser, error) {
	user, err := us.repo.getByID(id)
	if err != nil {
		return ShortUser{}, err
	}

	return user.toShortUser(), nil
}

type customerServiceImpl struct {
	userServiceImpl
	repo   customerRepo
	client *http.Client
}

func (cs *customerServiceImpl) register(u *User, image *multipart.FileHeader, emailAccessToken, numberAccessToken string) (string, error) {
	params := u.validate()
	if err := u.fomratPhoneNumber(); err != nil {
		params.SetInvalidParam("phoneNumber", err.Error())
	}

	emailSessionID, err := cs.AccessEmail(u.Email, emailAccessToken)
	if err != nil {
		params.SetInvalidParam("emailAccessToken", err.Error())
	}

	numberSessionID, err := cs.AccessNumber(u.PhoneNumber, numberAccessToken)
	if err != nil {
		params.SetInvalidParam("emailAccessToken", err.Error())
	}

	if params != nil {
		return "", rfc7807.BadRequest(
			"user-credentials-validation",
			"user Credentials Error",
			"Could not save the users due to invalid credentials.",
			params...,
		)
	}

	u.ID = uuid.New()

	if image != nil {
		url := config.APIURL() + "/static/imgs/" + u.ID.String()
		err := saveUserImage(url, image)
		if err != nil {
			return "", rfc7807.Internal("image-saving-error", err.Error())
		}
		u.ImageUrl = url
	} else {
		u.ImageUrl = config.APIURL() + "/imgs/guest-female.png"
	}

	u.Role.Val = auth.Customer
	err = cs.repo.create(u)
	if err != nil {
		return "", err
	}

	err = cs.repo.useVerifiedEmail(emailSessionID)
	if err != nil {
		return "", err
	}
	err = cs.repo.useVerifiedNumber(numberSessionID)
	if err != nil {
		return "", err
	}

	token, err := u.Role.Val.GenerateToken(u.Email, u.ID)
	return token, err
}

func (cs *customerServiceImpl) AccessEmail(email, token string) (uuid.UUID, error) {

	sessionID, err := auth.VerifyAccessToken(token, config.EmailAccessTokenSecretKey())

	if err != nil {
		return uuid.Nil, err
	}

	session, err := cs.repo.verifiedEmail(sessionID)
	if err != nil {
		return uuid.Nil, err
	}

	if session.Expires.Before(time.Now()) {
		return uuid.Nil, errors.New("Provided Token has Expired.")
	}

	if session.Email != email {
		return uuid.Nil, errors.New("Email assosiated with access token differs from the provided one.")
	}

	return sessionID, nil
}

func (cs *customerServiceImpl) AccessNumber(number, token string) (uuid.UUID, error) {

	sessionID, err := auth.VerifyAccessToken(token, config.NumberAccessTokenSecretKey())

	if err != nil {
		return uuid.Nil, err
	}

	session, err := cs.repo.verifiedNumber(sessionID)
	if err != nil {
		return uuid.Nil, err
	}

	if session.Expires.Before(time.Now()) {
		return uuid.Nil, errors.New("Provided Token has Expired.")
	}

	if session.Number != number {
		return uuid.Nil, errors.New("Number assosiated with access token differs from the provided one.")
	}

	return sessionID, nil
}

func saveUserImage(path string, image *multipart.FileHeader) error {
	src, err := image.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err = os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func (us *customerServiceImpl) delete(id uuid.UUID) error {
	return us.repo.delete(id)
}

func (cs *customerServiceImpl) verifyEmail(email string) (string, bool, error) {
	if !govalidator.IsEmail(email) {
		return "", false, rfc7807.BadRequest(
			"invalid-email",
			"Invalid Email Error",
			"Provided email contains forbidden characters or is not an email at all.",
		)
	}

	exists, err := cs.repo.emailExists(email)
	if err != nil || exists {
		return "", true, err
	}

	verificationCode, err := verification.VerifyEmail(email)
	if err != nil {
		rfc7807.BadGateway("email-verification-service", "Email Verification Error", err.Error())
	}

	sessionID, err := cs.repo.startEmailVerification(verificationCode, email)
	if err != nil {
		return "", false, err
	}

	token, err := auth.GenerateAccessToken(sessionID, config.EmailAccessTokenSecretKey(), time.Minute*10)

	return token, false, err
}

func validateVerificationCode(code string) error {
	var errMessage string

	if length := len(code); length != 6 {
		errMessage = fmt.Sprintf("Invalid code length. Want 6, got '%d'. ", length)
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(code) {
		errMessage += fmt.Sprintf("The code has to only contain digits, got '%s'.", code)
	}

	if errMessage != "" {
		return rfc7807.New(http.StatusUnprocessableEntity, "invalid-verificaiton-code-format", "Code Forman Error", errMessage)
	}

	return nil
}

func (cs *customerServiceImpl) verifyEmailCode(code, token string) (string, error) {
	err := validateVerificationCode(code)
	if err != nil {
		return "", err
	}

	sessionID, err := auth.VerifyAccessToken(token, config.EmailAccessTokenSecretKey())
	if err != nil {
		return "", rfc7807.Unauthorized("email-code-verification-token", "Unauthorized", "Unauthorized")
	}

	session, err := cs.repo.emailVerificationSession(sessionID)
	if err != nil {
		return "", err
	}

	if session.Expires.Before(time.Now()) {
		return "", rfc7807.New(http.StatusGone, "expired-session", "Expired Session Error", "The sesion has expired and no longer can be used to verify th email")
	}

	if code != session.Code {
		return "", rfc7807.BadRequest("incorect-email-verification-code", "Incorect Email Verification Code Error", "Provided code does not match the sent one")
	}

	sessionID, err = cs.repo.completeEmailVerification(sessionID, session.Email)
	if err != nil {
		return "", err
	}

	return auth.GenerateAccessToken(sessionID, config.EmailAccessTokenSecretKey(), time.Minute*10)
}

func (cs *customerServiceImpl) verifyNumber(number string) (string, error) {
	num, err := phonenumbers.Parse(number, "UA")
	if err != nil {
		return "", rfc7807.BadRequest("invalid-phone-number", "Phone Number Error", err.Error())
	}

	if !phonenumbers.IsValidNumber(num) {
		return "", rfc7807.BadRequest("invalid-phone-number", "Phone Number Error", "Provided phone number is invalid.")
	}

	verificationCode, err := verification.VerifyNumber(phonenumbers.Format(num, phonenumbers.E164))
	if err != nil {
		rfc7807.BadGateway("phone-number-verification", "Phone Number Verification Error", err.Error())
	}

	sessionID, err := cs.repo.startNumberVerification(verificationCode, phonenumbers.Format(num, phonenumbers.E164))
	if err != nil {
		return "", err
	}

	return auth.GenerateAccessToken(sessionID, config.NumberAccessTokenSecretKey(), time.Minute*10)
}

func (cs *customerServiceImpl) verifyNumberCode(code, token string) (string, error) {
	err := validateVerificationCode(code)
	if err != nil {
		return "", err
	}

	sessionID, err := auth.VerifyAccessToken(token, config.NumberAccessTokenSecretKey())
	if err != nil {
		return "", rfc7807.Unauthorized("email-code-verification-token", "Unauthorized", "Unauthorized")
	}

	session, err := cs.repo.numberVerificationSession(sessionID)
	if err != nil {
		return "", err
	}

	if session.Expires.Before(time.Now()) {
		return "", rfc7807.New(http.StatusGone, "expired-session", "Expired Session Error", "The sesion has expired and no longer can be used to verify th email")
	}

	if code != session.Code {
		return "", rfc7807.BadRequest("incorect-number-verification-code", "Incorect Number Verification Code Error", "Provided code does not match the sent one")
	}

	sessionID, err = cs.repo.completeNumberVerification(sessionID, session.Number)
	if err != nil {
		return "", err
	}

	return auth.GenerateAccessToken(sessionID, config.NumberAccessTokenSecretKey(), time.Minute*10)
}

func (cs *customerServiceImpl) googleOAUTH(code string, ctx context.Context, id uuid.UUID) (string, bool, error) {
	credentials, err := google.GetCredentialsByCode(code, ctx, cs.client)
	if err != nil {
		return "", false, err
	}

	id, exists, err := cs.repo.userExists(credentials.Email)
	if err != nil {
		return "", false, err
	}

	if exists {
		token, err := cs.role.GenerateToken(credentials.Email, id)

		return token, true, err
	}

	var user = User{
		ID:          id,
		Email:       credentials.Email,
		FirstName:   credentials.Name,
		LastName:    credentials.SurName,
		DateOfBirth: dateOfBirth{time.Now().AddDate(-25, 0, 0)},
		ImageUrl:    "https://example.com/default-guest-avatar.png",
		Role:        userRole{auth.Customer},
	}

	err = cs.repo.create(&user)
	if err != nil {
		return "", false, err
	}

	token, err := cs.role.GenerateToken(user.Email, user.ID)
	return token, false, err
}

//Declaration functions

func newCustomerServiceImpl(repo customerRepo, client *http.Client) customerService {
	return &customerServiceImpl{userServiceImpl{repo, auth.Customer}, repo, client}
}
