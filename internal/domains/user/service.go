package user

import (
	"context"
	"io"
	google "maryan_api/internal/infrastructure/clients/googleoauth"
	"maryan_api/internal/infrastructure/clients/verification"
	"maryan_api/pkg/auth"
	rfc7807 "maryan_api/pkg/problem"
	"maryan_api/pkg/security"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
)

// userServiceImpl defines a basic user struct for embadding.

type userService interface {
	login(email, password string) (string, error)
	loginJWT(id uuid.UUID, email string) (string, error)
	usersData(id uuid.UUID) (User, error)
	userService() userService
	secretKey() []byte
}

type customerService interface {
	userService

	save(u *User) error
	delete(id uuid.UUID) error
	verifyEmail(email string) (string, bool, error)
	verifyNumber(number string) (string, error)
	guest() (string, error)
	googleOAUTH(code string, ctx context.Context, id uuid.UUID) (string, bool, error)
}

type adminService interface {
	userService

	getUsers(pageNumber, pageSize int) ([]User, int64, error)
}

type driverService interface {
	userService
}

type supportEmployeeService interface {
	userService
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
			"Invalid Email",
			"Provided email contains forbidden characters or is not an email at all",
		)
	}

	id, passwordHashed, err := us.repo.login(email)
	if err != nil {
		return "", err
	}

	if ok := security.VerifyPassword(password, passwordHashed); !ok {
		return "", rfc7807.Unauthorized(
			"invalid-password",
			"Invalid Password",
			"Invalid password for user assosiated with provided email",
		)
	}

	token, err := us.role.GenerateToken(email, id)
	return token, err
}

func (us *userServiceImpl) loginJWT(id uuid.UUID, email string) (string, error) {
	if !govalidator.IsEmail(email) {
		return "", rfc7807.BadRequest(
			"email-invalid",
			"Invalid Email",
			"Provided email contains forbidden characters or is not an email at all",
		)
	}

	exists, err := us.repo.loginJWT(email, id)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", rfc7807.BadRequest(
			"unexistant-user",
			"Unexistant user",
			"There is no user assosiated with provided email",
		)
	}

	token, err := us.role.GenerateToken(email, id)

	return token, err
}

func (us *userServiceImpl) usersData(id uuid.UUID) (User, error) {
	u, err := us.repo.getByID(id)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

type customerServiceImpl struct {
	userServiceImpl
	repo   customerRepo
	client *http.Client
}

func (cs *customerServiceImpl) save(u *User) error {
	params, ok := u.validate()
	if err := u.fomratPhoneNumber(); err != nil {
		params = append(params, rfc7807.InvalidParam{"phoneNumber", err.Error()})
		ok = false
	}
	if !ok {
		return rfc7807.BadRequest(
			"user-credentials-validation",
			"user Credentials Error",
			"Could not save the users due to invalid credentials.",
		).SetInvalidParams(params)
	}

	u.Guest = false
	u.Role = userRole{cs.role}

	err := cs.repo.save(u)
	return err
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
	existed, err := us.repo.delete(id)

	if !existed {
		return rfc7807.BadRequest("unexistant-user", "Unexistant user Error", err.Error())
	}
	if err != nil {
		return err
	}
	return nil
}

func (cs *customerServiceImpl) verifyEmail(email string) (string, bool, error) {
	if !govalidator.IsEmail(email) {
		return "", false, rfc7807.BadRequest(
			"email-invalid",
			"Invalid Email",
			"Provided email contains forbidden characters or is not an email at all",
		)
	}

	exists, err := cs.repo.emailExists(email)
	if err != nil {
		return "", false, err
	}

	if exists {
		return "", true, nil
	}

	verificationCode, err := verification.VerifyEmail(email)
	if err != nil {
		rfc7807.BadGateway("email-verification", "Email Verification Error", err.Error())
	}

	return verificationCode, false, nil
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

	return verificationCode, nil
}

func (cs *customerServiceImpl) guest() (string, error) {
	var id = uuid.New()
	var guest = User{
		ID:          id,
		DateOfBirth: dateOfBirth{time.Now().AddDate(-25, 0, 0)},
		Email:       id.String(),
		Guest:       true,
		ImageUrl:    "https://example.com/default-guest-avatar.png",
		Role:        userRole{auth.Customer},
	}

	err := cs.repo.save(&guest)
	if err != nil {
		return "", err
	}

	token, err := cs.role.GenerateToken(guest.Email, guest.ID)

	return token, err
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

	err = cs.repo.save(&user)
	if err != nil {
		return "", false, err
	}

	token, err := cs.role.GenerateToken(user.Email, user.ID)
	return token, false, err
}

type adminServiceImpl struct {
	userServiceImpl
	repo adminRepo
}

func (us *adminServiceImpl) getUsers(pageNumber, pageSize int) ([]User, int64, error) {
	params, add, isNil := rfc7807.StartSettingInvalidParams()

	if pageNumber < 1 {
		add("pageNumber", "Has to be greater than 0")
	}
	if pageSize < 1 {
		add("pageSize", "Has to be greater than 0")
	}

	if !isNil() {
		return nil, 0, rfc7807.BadRequest(
			"users-order-data",
			"user Order Data Error",
			"Could not retrieve users due to invalid users order data.",
		).SetInvalidParams(*params)
	}

	users, pages, err := us.repo.getUsers(pageNumber, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return users, pages, nil
}

type driverServiceImpl struct {
	userServiceImpl
}

type supportEmployeeServiceImpl struct {
	userServiceImpl
}

//Declaration functions

func newCustomerServiceImpl(repo customerRepo, client *http.Client) customerService {
	return &customerServiceImpl{userServiceImpl{repo.repo(), auth.Customer}, repo, client}
}

func newAdminServiceImpl(repo adminRepo) adminService {
	return &adminServiceImpl{userServiceImpl{repo.repo(), auth.Admin}, repo}
}

func newDriverServiceImpl(repo driverRepo) driverService {
	return &driverServiceImpl{userServiceImpl{repo.repo(), auth.Driver}}
}

func newSupportEmployeeServiceImpl(repo supportEmployeeRepo) supportEmployeeService {
	return &supportEmployeeServiceImpl{userServiceImpl{repo.repo(), auth.SupportEmployee}}
}
