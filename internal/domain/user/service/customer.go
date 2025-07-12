package service

import (
	"context"
	"errors"
	"maryan_api/config"
	"maryan_api/internal/domain/user/repo"
	"maryan_api/internal/entity"
	google "maryan_api/internal/infrastructure/clients/google/oauth_2.0"
	"maryan_api/internal/infrastructure/clients/verification"
	"maryan_api/internal/valueobject"
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

	//----------Not authenticated------------------
	Register(ctx context.Context, u entity.RegistrantionUser, image *multipart.FileHeader, emailAccessToken, numberAccessToken string) (string, error)

	VerifyEmail(ctx context.Context, email string) (string, bool, error)
	VerifyEmailCode(ctx context.Context, code, token string) (string, error)

	VerifyNumber(ctx context.Context, number string) (string, error)
	VerifyNumberCode(ctx context.Context, code, token string) (string, error)

	GoogleOAUTH(ctx context.Context, code string) (string, bool, error)

	//------------Authenticated--------------------
	Delete(ctx context.Context, id uuid.UUID) error
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
		return errors.New("invalid email access token")
	}

	tokenEmail := claims[0].(string)
	if email != tokenEmail {
		return errors.New("email in token does not match provided email")
	}

	return nil
}

func (cs *customerServiceImpl) VerifyNumberToken(token string, number string) error {
	claims, err := auth.VerifyAccessToken(token, config.NumberAccessTokenSecretKey(), []auth.ClaimValidation{{
		"number",
		true,
		auth.ClaimString,
	}})

	if err != nil {
		return errors.New("invalid number access token")
	}

	tokenNumber := claims[0].(string)
	if number != tokenNumber {
		return errors.New("number in token does not match provided number")
	}

	return nil
}

func (cs *customerServiceImpl) Register(ctx context.Context, ru entity.RegistrantionUser, image *multipart.FileHeader, emailAccessToken, numberAccessToken string) (string, error) {
	u := ru.ToUser(cs.Role())
	invalidParams := u.PrepareNew()

	err := cs.VerifyEmailToken(emailAccessToken, u.Email)
	if err != nil {
		invalidParams.SetInvalidParam("EmailToken", err.Error())
	}

	err = cs.VerifyNumberToken(numberAccessToken, u.PhoneNumber)
	if err != nil {
		invalidParams.SetInvalidParam("NumberToken", err.Error())
	}

	if invalidParams != nil {
		return "", rfc7807.BadRequest(
			"user-credentials-validation",
			"user Credentials Error",
			"Could not save the user due to invalid credentials.",
			invalidParams...,
		)
	}

	if image != nil {
		err := images.Save("../../../../static/imgs/"+u.ID.String(), image)
		if err != nil {
			return "", rfc7807.Internal("image-saving-error", err.Error())
		}
		u.ImageUrl = config.APIURL() + "/imgs/" + u.ID.String()
	} else {
		u.ImageUrl = config.APIURL() + "/imgs/guest-female.png"
	}

	u.Role.Val = auth.Customer
	err = cs.repo.Create(ctx, &u)
	if err != nil {
		return "", err
	}

	token, err := u.Role.Val.GenerateToken(u.Email, u.ID)
	return token, err
}

func (cs *customerServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return cs.repo.Delete(ctx, id)
}

func (cs *customerServiceImpl) VerifyEmail(ctx context.Context, email string) (string, bool, error) {
	if !govalidator.IsEmail(email) {
		return "", false, rfc7807.BadRequest(
			"invalid-email",
			"Invalid Email Error",
			"Provided email contains forbidden characters or is not a valid email.",
		)
	}

	_, exists, err := cs.repo.EmailExists(ctx, email)
	if err != nil || exists {
		return "", true, err
	}

	verificationCode, err := verification.VerifyEmail(email)
	if err != nil {
		return "", false, rfc7807.BadGateway("email-verification-service", "Email Verification Error", err.Error())
	}

	sessionID, err := cs.repo.StartEmailVerification(ctx, valueobject.NewEmailVerificationSession(verificationCode, email))
	if err != nil {
		return "", false, err
	}

	token, err := auth.GenerateAccessToken(config.EmailAccessTokenSecretKey(), jwt.MapClaims{"email": email, "id": sessionID.String()})

	return token, false, err
}

func (cs *customerServiceImpl) VerifyEmailCode(ctx context.Context, code, token string) (string, error) {
	err := valueobject.ValidateVerificationCode(code)
	if err != nil {
		return "", err
	}

	claims, err := auth.VerifyAccessToken(token, config.EmailAccessTokenSecretKey(), []auth.ClaimValidation{
		{"email", true, auth.ClaimString},
		{"id", true, auth.ClaimUUID},
	})
	if err != nil {
		return "", rfc7807.Unauthorized("email-code-verification-token", "Unauthorized", "Unauthorized")
	}

	email := claims[0].(string)
	sessionID := claims[1].(uuid.UUID)

	session, err := cs.repo.EmailVerificationSession(ctx, sessionID)
	if err != nil {
		return "", err
	}

	if session.Expires.Before(time.Now()) {
		return "", rfc7807.New(http.StatusGone, "expired-session", "Expired Session Error", "The session has expired and can no longer be used for verification.")
	}

	if email != session.Email {
		return "", rfc7807.BadRequest("incorrect-email-verification-token", "Incorrect Email Verification Token Error", "Provided token does not match the previously sent one.")
	}

	if code != session.Code {
		return "", rfc7807.BadRequest("incorrect-email-verification-code", "Incorrect Email Verification Code Error", "Provided code does not match the sent one.")
	}

	err = cs.repo.CompleteEmailVerification(ctx, sessionID)
	if err != nil {
		return "", err
	}

	return auth.GenerateAccessToken(config.EmailAccessTokenSecretKey(), jwt.MapClaims{"email": email})
}

func (cs *customerServiceImpl) VerifyNumber(ctx context.Context, number string) (string, error) {
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
		return "", rfc7807.BadGateway("phone-number-verification", "Phone Number Verification Error", err.Error())
	}

	sessionID, err := cs.repo.StartNumberVerification(ctx, valueobject.NewNumberVerificationSession(verificationCode, numberE164))
	if err != nil {
		return "", err
	}

	return auth.GenerateAccessToken(config.NumberAccessTokenSecretKey(), jwt.MapClaims{"number": numberE164, "id": sessionID.String()})
}

func (cs *customerServiceImpl) VerifyNumberCode(ctx context.Context, code, token string) (string, error) {
	err := valueobject.ValidateVerificationCode(code)
	if err != nil {
		return "", err
	}

	claims, err := auth.VerifyAccessToken(token, config.NumberAccessTokenSecretKey(), []auth.ClaimValidation{
		{"number", true, auth.ClaimString},
		{"id", true, auth.ClaimUUID},
	})
	if err != nil {
		return "", rfc7807.Unauthorized("number-code-verification-token", "Unauthorized", "Unauthorized")
	}

	number := claims[0].(string)
	sessionID := claims[1].(uuid.UUID)

	session, err := cs.repo.NumberVerificationSession(ctx, sessionID)
	if err != nil {
		return "", err
	}

	if number != session.Number {
		return "", rfc7807.BadRequest("incorrect-number-verification-token", "Incorrect Number Verification Token Error", "Provided token does not match the sent one")
	}

	if session.Expires.Before(time.Now()) {
		return "", rfc7807.New(http.StatusGone, "expired-session", "Expired Session Error", "The session has expired and can no longer be used for verification")
	}

	if code != session.Code {
		return "", rfc7807.BadRequest("incorrect-number-verification-code", "Incorrect Number Verification Code Error", "Provided code does not match the sent one")
	}

	err = cs.repo.CompleteNumberVerification(ctx, sessionID)
	if err != nil {
		return "", err
	}

	return auth.GenerateAccessToken(config.NumberAccessTokenSecretKey(), jwt.MapClaims{"number": number})
}

func (cs *customerServiceImpl) GoogleOAUTH(ctx context.Context, code string) (string, bool, error) {
	credentials, err := google.GetCredentialsByCode(code, ctx, cs.client)
	if err != nil {
		return "", false, err
	}

	id, exists, err := cs.repo.EmailExists(ctx, credentials.Email)
	if err != nil {
		return "", false, err
	}

	if exists {
		token, err := cs.Role().GenerateToken(credentials.Email, id)
		return token, true, err
	}

	user := entity.NewForGoogleOAUTH(credentials.Email, credentials.Name, credentials.SurName)
	err = cs.repo.Create(ctx, &user)
	if err != nil {
		return "", false, err
	}

	token, err := cs.Role().GenerateToken(user.Email, user.ID)
	return token, false, err
}

func NewCustomerServiceImpl(repo repo.CustomerRepo, client *http.Client) CustomerService {
	return &customerServiceImpl{
		UserService: NewUserService(auth.Customer, repo),
		repo:        repo,
		client:      client,
	}
}
