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
	rfc7807 "maryan_api/pkg/problem"
	"maryan_api/pkg/security"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/d3code/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nyaruka/phonenumbers"
)

type CustomerService interface {
	UserService

	//----------Not authenticated------------------
	Register(ctx context.Context, u entity.RegistrantionUser, image *multipart.FileHeader, saveImageFunc func(file *multipart.FileHeader, dst string) error, emailAccessToken string) (string, error)

	VerifyEmailIfExists(ctx context.Context, email string) (string, bool, error)
	VerifyEmailCode(ctx context.Context, code, token string) (string, error)
	VerifyEmailPasswordChange(ctx context.Context, email string) (string, error)
	VerifyEmailCodePasswordChange(ctx context.Context, code, token string) (string, error)

	VerifyNumber(ctx context.Context, number string) (string, error)
	VerifyNumberCode(ctx context.Context, code, token string) (string, error)

	GoogleOAUTH(ctx context.Context, code string) (string, bool, error)
	ChangePassword(ctx context.Context, newPassword, email, emailAccessToken string) error
	//------------Authenticated--------------------
	Delete(ctx context.Context, id uuid.UUID) error
	UpdatePersonalInfo(ctx context.Context, userBasicInfo entity.UserPersonalInfo, id uuid.UUID) error
	VerifyEmailCustomerUpdate(ctx context.Context, email string) (string, error)
	VerifyCustomerUpdateCode(ctx context.Context, code, token string) (string, error)
	UpdateContactInfo(ctx context.Context, id uuid.UUID, user entity.UserContactInfo, emailAccessToken string) error
}

type customerServiceImpl struct {
	UserService
	repo   repo.CustomerRepo
	client *http.Client
}

func (cs *customerServiceImpl) verifyEmailToken(token string, email string, secretKey []byte) error {
	claims, err := auth.VerifyAccessToken(token, secretKey, []auth.ClaimValidation{{
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

func (cs *customerServiceImpl) verifyEmailTokenSoft(token string, secretKey []byte) error {
	_, err := auth.VerifyAccessToken(token, secretKey, []auth.ClaimValidation{{
		"email",
		true,
		auth.ClaimString,
	}})

	if err != nil {
		return errors.New("invalid email access token")
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

func (cs *customerServiceImpl) Register(ctx context.Context, ru entity.RegistrantionUser, image *multipart.FileHeader, saveImageFunc func(file *multipart.FileHeader, dst string) error, emailAccessToken string) (string, error) {
	u := ru.ToUser(cs.Role())
	invalidParams := u.PrepareNew()

	err := cs.verifyEmailToken(emailAccessToken, u.Email, config.EmailAccessTokenSecretKey())
	if err != nil {
		invalidParams.SetInvalidParam("EmailToken", err.Error())
	}

	// err = cs.VerifyNumberToken(numberAccessToken, u.PhoneNumber)
	// if err != nil {
	// 	invalidParams.SetInvalidParam("NumberToken", err.Error())
	// }

	if invalidParams != nil {
		return "", rfc7807.BadRequest(
			"user-credentials-validation",
			"user Credentials Error",
			"Could not save the user due to invalid credentials.",
			invalidParams...,
		)
	}

	if image != nil {
		imageName := u.ID.String() + ".jpg"
		filePath := filepath.Join("../../static", "imgs", imageName)
		err := saveImageFunc(image, filePath)
		if err != nil {
			return "", rfc7807.Internal("image-saving-error", err.Error())
		}
		u.ImageUrl = config.APIURL() + "/imgs/" + u.ID.String() + ".jpg"
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

func (cs *customerServiceImpl) ChangePassword(ctx context.Context, newPassword, email, emailAccessToken string) error {
	err := cs.verifyEmailToken(emailAccessToken, email, config.EmailChangePasswordAccessTokenSecretKey())
	if err != nil {
		return rfc7807.Unauthorized("email-validation-acccess-token", "Email Access Token Error", "Invalid token.")
	}

	err = entity.ValidatePassword(newPassword)
	if err != nil {
		return rfc7807.BadRequest("invalid-password", "Invalid Password Error", "Invalid password. Has to be at least 6 characters long, contain at least one uppercase letter and one digit.")
	}

	hashedPassword, err := security.HashPassword(newPassword)
	if err != nil {
		return rfc7807.Internal("password-hashing", err.Error())
	}

	return cs.repo.ChangePassword(ctx, hashedPassword, email)

}

func (cs *customerServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return cs.repo.Delete(ctx, id)
}

func (cs *customerServiceImpl) VerifyEmailIfExists(ctx context.Context, email string) (string, bool, error) {
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

	token, err := auth.GenerateAccessToken(config.EmailCodeVerificationTokenSecretKey(), jwt.MapClaims{"email": email, "id": sessionID.String()})

	return token, false, err
}

func (cs *customerServiceImpl) VerifyEmailPasswordChange(ctx context.Context, email string) (string, error) {
	return cs.verifyEmail(ctx, email, config.EmailChangePasswordCodeVerificationTokenSecretKey())
}

func (cs *customerServiceImpl) VerifyEmailCustomerUpdate(ctx context.Context, email string) (string, error) {
	return cs.verifyEmail(ctx, email, config.EmailCodeSecretKeyCustomerUpdate())
}

func (cs *customerServiceImpl) verifyEmail(ctx context.Context, email string, secretKey []byte) (string, error) {

	if !govalidator.IsEmail(email) {
		return "", rfc7807.BadRequest(
			"invalid-email",
			"Invalid Email Error",
			"Provided email contains forbidden characters or is not a valid email.",
		)
	}

	_, exists, err := cs.repo.EmailExists(ctx, email)
	if err != nil {
		return "", err
	} else if !exists {
		return "", rfc7807.BadRequest(
			"non-existing-email",
			"Non-existing Email Errro",
			"THere is no user assosiated with provided email.",
		)
	}

	verificationCode, err := verification.VerifyEmail(email)
	if err != nil {
		return "", rfc7807.BadGateway("email-verification-service", "Email Verification Error", err.Error())
	}

	sessionID, err := cs.repo.StartEmailVerification(ctx, valueobject.NewEmailVerificationSession(verificationCode, email))
	if err != nil {
		return "", err
	}

	token, err := auth.GenerateAccessToken(secretKey, jwt.MapClaims{"email": email, "id": sessionID.String()})

	return token, err

}

func (cs *customerServiceImpl) verifyEmailCodeWithSecretKey(codeAccessKey, emailAcessKey []byte) func(ctx context.Context, code, token string) (string, error) {
	return func(ctx context.Context, code, token string) (string, error) {

		err := valueobject.ValidateVerificationCode(code)
		if err != nil {
			return "", err
		}

		claims, err := auth.VerifyAccessToken(token, codeAccessKey, []auth.ClaimValidation{
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

		return auth.GenerateAccessToken(emailAcessKey, jwt.MapClaims{"email": email})
	}
}

func (cs *customerServiceImpl) VerifyEmailCode(ctx context.Context, code, token string) (string, error) {
	return cs.verifyEmailCodeWithSecretKey(config.EmailCodeVerificationTokenSecretKey(), config.EmailAccessTokenSecretKey())(ctx, code, token)
}

func (cs *customerServiceImpl) VerifyEmailCodePasswordChange(ctx context.Context, code, token string) (string, error) {
	return cs.verifyEmailCodeWithSecretKey(config.EmailChangePasswordCodeVerificationTokenSecretKey(), config.EmailChangePasswordAccessTokenSecretKey())(ctx, code, token)
}

func (cs *customerServiceImpl) VerifyCustomerUpdateCode(ctx context.Context, code, token string) (string, error) {
	return cs.verifyEmailCodeWithSecretKey(config.EmailCodeSecretKeyCustomerUpdate(), config.SecretKeyCustomerUpdate())(ctx, code, token)
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

	user := entity.NewForGoogleOAUTH(credentials.Email, credentials.FirstName, credentials.LastName, credentials.DateOfBirth)

	err = cs.repo.Create(ctx, &user)
	if err != nil {
		return "", false, err
	}

	token, err := cs.Role().GenerateToken(user.Email, user.ID)
	return token, false, err
}

func (cs *customerServiceImpl) UpdatePersonalInfo(ctx context.Context, user entity.UserPersonalInfo, id uuid.UUID) error {
	params := user.Validate()

	dateOfBirth, err := time.Parse("2006-01-02", user.DateOfBirth)
	if err != nil {
		params.SetInvalidParam("dateOfBirth", err.Error())
	}

	if dateOfBirth.Before(time.Now().AddDate(-125, 0, 0)) {
		params.SetInvalidParam("dateOfBirth", "Has to be younger than 125 years old.")
	} else if dateOfBirth.After(time.Now().AddDate(-18, 0, 0)) {
		params.SetInvalidParam("dateOfBirth", "Has to be at least 18 years old.")
	}

	if params != nil {
		return rfc7807.BadRequest("invalid-user-basic-info", "Invalid User Basic Info Error", "Provided data is not valid", params...)
	}

	return cs.repo.UpdatePersonalInfo(ctx, user.FirstName, user.LastName, dateOfBirth, id)
}

func (cs *customerServiceImpl) UpdateContactInfo(ctx context.Context, id uuid.UUID, user entity.UserContactInfo, emailAccessToken string) error {
	err := cs.verifyEmailTokenSoft(emailAccessToken, config.SecretKeyCustomerUpdate())
	if err != nil {
		return rfc7807.Unauthorized("customer-update-acccess-token", "Customer Update Token Error", err.Error())
	}

	params := user.Prepare()
	if params != nil {
		return rfc7807.BadRequest("invalid-user-contact-info", "Invalid User Contact Info Error", "Provided data is not valid", params...)
	}

	return cs.repo.UpdateContactInfo(ctx, user.Email, user.PhoneNumber, id)

}

func NewCustomerServiceImpl(repo repo.CustomerRepo, client *http.Client) CustomerService {
	return &customerServiceImpl{
		UserService: NewUserService(auth.Customer, repo),
		repo:        repo,
		client:      client,
	}
}
