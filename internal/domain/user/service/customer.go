package service

import (
	"context"
	"errors"
	"maryan_api/config"
	"maryan_api/internal/domain/user/repo"
	"maryan_api/internal/entity"
	google "maryan_api/internal/infrastructure/clients/googleoauth"
	"maryan_api/internal/infrastructure/clients/verification"
	"maryan_api/pkg/auth"
	"maryan_api/pkg/images"
	rfc7807 "maryan_api/pkg/problem"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
)

type CustomerService interface {
	UserService

	//----------Not authentificated------------------
	Register(u entity.RegistrantionUser, image *multipart.FileHeader, emailAccessToken, numberAccessToken string, ctx context.Context) (string, error)

	VerifyEmail(email string, ctx context.Context) (string, bool, error)
	VerifyEmailCode(code, token string, ctx context.Context) (string, error)

	VerifyNumber(number string, ctx context.Context) (string, error)
	VerifyNumberCode(number, token string, ctx context.Context) (string, error)

	GoogleOAUTH(code string, ctx context.Context) (string, bool, error)

	//------------Authentificated--------------------
	Delete(id uuid.UUID, ctx context.Context) error
}

type customerServiceImpl struct {
	UserService
	repo   repo.CustomerRepo
	client *http.Client
}

func (cs *customerServiceImpl) VerifyEmailToken(token string, email string) error {
	claims, err := auth.VerifyAccessToken(token, config.EmailAccessTokenSecretKey(), []auth.ClaimValidation{{
		"email",
		true,
		auth.ClaimString,
	}})

	if err != nil {
		return errors.New("Invalid email access token.")
	}

	tokenEmail := claims[0].(string)
	if email != tokenEmail {
		return errors.New("Invalid email access token. Sent email does not mach the one in the token")
	}

	return nil

}

func (cs *customerServiceImpl) VerifyNumberToken(token string, number string) error {
	claims, err := auth.VerifyAccessToken(token, config.EmailAccessTokenSecretKey(), []auth.ClaimValidation{{
		"number",
		true,
		auth.ClaimString,
	}})

	if err != nil {
		return errors.New("Invalid number access token.")
	}

	tokenNumber := claims[0].(string)
	if number != tokenNumber {
		return errors.New("Invalid number access token. Sent number does not mach the one in the token")
	}

	return nil

}

func (cs *customerServiceImpl) Register(ru entity.RegistrantionUser, image *multipart.FileHeader, emailAccessToken, numberAccessToken string, ctx context.Context) (string, error) {
	u, invalidParams := ru.ToNewUser(cs.Role())

	err := cs.VerifyEmailToken(emailAccessToken, u.Email)
	if err != nil {
		invalidParams.SetInvalidParam("EmailToken", err.Error())
	}

	err = cs.VerifyNumberToken(numberAccessToken, u.PhoneNumber)
	if err != nil {
		invalidParams.SetInvalidParam("EmailToken", err.Error())
	}

	if invalidParams != nil {
		return "", rfc7807.BadRequest(
			"user-credentials-validation",
			"user Credentials Error",
			"Could not save the users due to invalid credentials.",
			invalidParams...,
		)
	}

	u.ID = uuid.New()

	if image != nil {

		err := images.Save("../../../static/imgs/"+u.ID.String(), image)
		if err != nil {
			return "", rfc7807.Internal("image-saving-error", err.Error())
		}
		u.ImageUrl = config.APIURL() + "/imgs/" + u.ID.String()
	} else {
		u.ImageUrl = config.APIURL() + "/imgs/guest-female.png"
	}

	u.Role.Val = auth.Customer
	err = cs.repo.Create(&u, ctx)
	if err != nil {
		return "", err
	}

	token, err := u.Role.Val.GenerateToken(u.Email, u.ID)
	return token, err
}

func (us *customerServiceImpl) Delete(id uuid.UUID, ctx context.Context) error {
	return us.repo.Delete(id, ctx)
}

func (cs *customerServiceImpl) VerifyEmail(email string, ctx context.Context) (string, bool, error) {
	if !govalidator.IsEmail(email) {
		return "", false, rfc7807.BadRequest(
			"invalid-email",
			"Invalid Email Error",
			"Provided email contains forbidden characters or is not an email at all.",
		)
	}

	exists, err := cs.repo.EmailExists(email, ctx)
	if err != nil || exists {
		return "", true, err
	}

	verificationCode, err := verification.VerifyEmail(email)
	if err != nil {
		rfc7807.BadGateway("email-verification-service", "Email Verification Error", err.Error())
	}

	sessionID, err := cs.repo.StartEmailVerification(entity.NewEmailVerificationSession(verificationCode, email), ctx)
	if err != nil {
		return "", false, err
	}

	token, err := auth.GenerateAccessToken(config.EmailAccessTokenSecretKey(), jwt.MapClaims{"email": email, "id": sessionID.String()})

	return token, false, err
}

func (cs *customerServiceImpl) VerifyEmailCode(code, token string, ctx context.Context) (string, error) {
	err := entity.ValidateVerificationCode(code)
	if err != nil {
		return "", err
	}

	claims, err := auth.VerifyAccessToken(token, config.EmailAccessTokenSecretKey(), []auth.ClaimValidation{
		{"email", true, auth.ClaimString},
		{"id", true, auth.ClaimUUID},
	})

	email := claims[0].(string)
	sessionID := claims[0].(uuid.UUID)

	if err != nil {
		return "", rfc7807.Unauthorized("email-code-verification-token", "Unauthorized", "Unauthorized")
	}

	session, err := cs.repo.EmailVerificationSession(sessionID, ctx)
	if err != nil {
		return "", err
	}

	if session.Expires.Before(time.Now()) {
		return "", rfc7807.New(http.StatusGone, "expired-session", "Expired Session Error", "The sesion has expired and no longer can be used to verify th email.")
	}

	if email != session.Code {
		return "", rfc7807.BadRequest("incorect-email-verification-token", "Incorect Email Verification Token Error", "Provided token does not match the previously sent one.")
	}

	if code != session.Code {
		return "", rfc7807.BadRequest("incorect-email-verification-code", "Incorect Email Verification Code Error", "Provided code does not match the sent one.")
	}

	err = cs.repo.CompleteEmailVerification(sessionID, ctx)
	if err != nil {
		return "", err
	}

	return auth.GenerateAccessToken(config.EmailAccessTokenSecretKey(), jwt.MapClaims{"email": email})
}

func (cs *customerServiceImpl) VerifyNumber(number string, ctx context.Context) (string, error) {
	num, err := phonenumbers.Parse(number, "UA")
	if err != nil {
		return "", rfc7807.BadRequest("invalid-phone-number", "Phone Number Error", err.Error())
	}

	if !phonenumbers.IsValidNumber(num) {
		return "", rfc7807.BadRequest("invalid-phone-number", "Phone Number Error", "Provided phone number is invalid.")
	}

	numberE164 := phonenumbers.Format(num, phonenumbers.E164)
	verificationCode, err := verification.VerifyNumber(numberE164)
	if err != nil {
		rfc7807.BadGateway("phone-number-verification", "Phone Number Verification Error", err.Error())
	}

	sessionID, err := cs.repo.StartNumberVerification(entity.NewNumberVerificationSession(verificationCode, numberE164), ctx)
	if err != nil {
		return "", err
	}

	return auth.GenerateAccessToken(config.NumberAccessTokenSecretKey(), jwt.MapClaims{"number": phonenumbers.Format(num, phonenumbers.E164), "id": sessionID})
}

func (cs *customerServiceImpl) VerifyNumberCode(code, token string, ctx context.Context) (string, error) {
	err := entity.ValidateVerificationCode(code)
	if err != nil {
		return "", err
	}

	claims, err := auth.VerifyAccessToken(token, config.NumberAccessTokenSecretKey(), []auth.ClaimValidation{
		{"number", true, auth.ClaimString},
		{"id", true, auth.ClaimUUID},
	})

	number := claims[0].(string)
	sessionID := claims[0].(uuid.UUID)

	if err != nil {
		return "", rfc7807.Unauthorized("email-code-verification-token", "Unauthorized", "Unauthorized")
	}

	session, err := cs.repo.NumberVerificationSession(sessionID, ctx)
	if err != nil {
		return "", err
	}

	if number != session.Number {
		return "", rfc7807.BadRequest("incorect-number-verification-token", "Incorect Number Verification Token Error", "Provided token does not match the sent one")
	}

	if session.Expires.Before(time.Now()) {
		return "", rfc7807.New(http.StatusGone, "expired-session", "Expired Session Error", "The sesion has expired and no longer can be used to verify th email")
	}

	if code != session.Code {
		return "", rfc7807.BadRequest("incorect-number-verification-code", "Incorect Number Verification Code Error", "Provided code does not match the sent one")
	}

	err = cs.repo.CompleteNumberVerification(sessionID, ctx)
	if err != nil {
		return "", err
	}

	return auth.GenerateAccessToken(config.NumberAccessTokenSecretKey(), jwt.MapClaims{"number": number})
}

func (cs *customerServiceImpl) GoogleOAUTH(code string, ctx context.Context) (string, bool, error) {
	credentials, err := google.GetCredentialsByCode(code, ctx, cs.client)
	if err != nil {
		return "", false, err
	}

	id, exists, err := cs.repo.UserExists(credentials.Email, ctx)
	if err != nil {
		return "", false, err
	}

	if exists {
		token, err := cs.Role().GenerateToken(credentials.Email, id)

		return token, true, err
	}

	var user = entity.NewForGoogleOAUTH(credentials.Email, credentials.Name, credentials.SurName)

	err = cs.repo.Create(&user, ctx)
	if err != nil {
		return "", false, err
	}

	token, err := cs.Role().GenerateToken(user.Email, user.ID)
	return token, false, err
}

func NewCustomerServiceImpl(repo repo.CustomerRepo, client *http.Client) CustomerService {
	return &customerServiceImpl{newUserService(auth.Customer, repo), repo, client}
}
