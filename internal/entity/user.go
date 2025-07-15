package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"maryan_api/config"
	"maryan_api/pkg/auth"
	rfc7807 "maryan_api/pkg/problem"
	"maryan_api/pkg/security"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"

	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
	"gorm.io/gorm"
)

// USER
type User struct {
	ID          uuid.UUID      `gorm:"type:binary(16);primaryKey"                                              json:"id"`
	FirstName   string         `gorm:"type:varchar(50);not null"                                         json:"firstName"`
	LastName    string         `gorm:"type:varchar(50);not null"                                         json:"lastName"`
	DateOfBirth time.Time      `gorm:"type:DATE;not null"                                                json:"dateOfBirth"`
	PhoneNumber string         `gorm:"type:varchar(15)"                                                  json:"phoneNumber"`
	Email       string         `gorm:"type:varchar(255);not null;unique; index"                          json:"email"`
	Password    string         `gorm:"type:varchar(255);not null"                                        json:"password"`
	ImageUrl    string         `gorm:"type:varchar(255);not null"                                        json:"imageUrl"`
	Role        userRole       `gorm:"type:enum('Customer','Admin','Driver','Support');not null"         json:"-"`
	CreatedAt   time.Time      `gorm:"not null"                                                          json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"not null"                                                          json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `                                                                         json:"deletedAt"`
}

func (u *User) AfterFind(tx *gorm.DB) (err error) {

	return
}

type UserPersonalInfo struct {
	FirstName   string
	LastName    string
	DateOfBirth string
}

func (u UserPersonalInfo) Validate() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams

	if len(u.FirstName) < 1 {
		params.SetInvalidParam("firstName", "Cannot be blank.")
	}

	if len(u.LastName) < 1 {
		params.SetInvalidParam("lastName", "Cannot be blank.")
	}

	return params
}

type UserContactInfo struct {
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"email"`
}

func (u UserContactInfo) Prepare() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams

	if !govalidator.IsEmail(u.Email) {
		params.SetInvalidParam("email", "Contains invalid characters or is not an email.")
	}

	phoneNumber, err := fomratPhoneNumber(u.PhoneNumber)
	if err != nil {
		params.SetInvalidParam("phoneNumber", err.Error())
	}

	u.PhoneNumber = phoneNumber

	return params
}

// USER -> ROLE
type userRole struct {
	Val auth.Role
}

func (ur userRole) MarshalJSON() ([]byte, error) {
	return json.Marshal((ur.Val.Name()))
}

func (ur userRole) Value() (driver.Value, error) {
	if ur.Val == nil {
		return nil, errors.New("Role is a nil interface")
	}
	return ur.Val.Name(), nil

}
func (ur *userRole) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		var err error
		ur.Val, err = auth.DefineRole(v)
		return err
	case []byte:
		str := string(v)
		var err error
		ur.Val, err = auth.DefineRole(str)
		return err
	default:
		return fmt.Errorf("UserRole: cannot scan type %T into string", value)
	}
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Password != "" {
		u.Password, err = security.HashPassword(u.Password)
	}
	return

}

type userSex string

const (
	maleSex   userSex = "Male"
	femaleSex userSex = "Female"
)

func sexImage(sex string) (string, error) {
	switch sex {
	case string(maleSex):
		return config.APIURL() + "/imgs/guest-male.png", nil
	case string(femaleSex):
		return config.APIURL() + "/imgs/guest-female.png", nil
	default:
		return "", rfc7807.BadRequest("incorect-sex", "Sex Error", "Provided sex is not valid", rfc7807.InvalidParam{"sex", fmt.Sprintf("Can only be 'male' or 'female', got '%s'.", sex)})
	}

}

// ************************************* //
// USER HELPING METHODS FOR THE SERVICE //
// ************************************* //
func fomratPhoneNumber(phoneNumber string) (string, error) {
	pn, err := phonenumbers.Parse(phoneNumber, "UA")
	if err != nil {
		return "", err
	}

	if !phonenumbers.IsValidNumber(pn) {
		return "", errors.New("invalid phone number")
	}

	return phonenumbers.Format(pn, phonenumbers.E164), nil
}

func (u *User) Validate() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams

	if len(u.FirstName) < 1 {
		params.SetInvalidParam("firstName", "Cannot be blank.")
	}

	if len(u.LastName) < 1 {
		params.SetInvalidParam("lastName", "Cannot be blank.")
	}

	if u.DateOfBirth.Before(time.Now().AddDate(-125, 0, 0)) {
		params.SetInvalidParam("dateOfBirth", "Has to be younger than 125 years old.")
	} else if u.DateOfBirth.After(time.Now().AddDate(-18, 0, 0)) {
		params.SetInvalidParam("dateOfBirth", "Has to be at least 18 years old.")
	}

	if !govalidator.IsEmail(u.Email) {
		params.SetInvalidParam("email", "Contains invalid characters or is not an email.")
	}

	if len(u.Password) < 6 {
		params.SetInvalidParam("password", "Has to be at least 6 characters long.")
	}

	if !govalidator.HasUpperCase(u.Password) {
		params.SetInvalidParam("password", "Has to contain at least one uppercase letter.")
	}

	if !strings.ContainsAny(u.Password, "0123456789") {
		params.SetInvalidParam("password", "Has to contain at least one number.")
	}

	return params
}

func (u *User) PrepareNew() rfc7807.InvalidParams {
	invalidParams := u.Validate()

	phoneNumber, err := fomratPhoneNumber(u.PhoneNumber)
	if err != nil {
		invalidParams.SetInvalidParam("phoneNumber", err.Error())
	}
	u.PhoneNumber = phoneNumber

	if invalidParams == nil {
		u.ID = uuid.New()
	}

	return invalidParams
}

func (u *User) PrepareNewEmployee(firstWorkingDay time.Time) ([]EmployeeAvailability, rfc7807.InvalidParams) {

	params := u.PrepareNew()

	now := time.Now()

	if firstWorkingDay.Before(now) {
		params.SetInvalidParam("starts", "Is in past.")
		return nil, params
	}

	datesAmount := int(firstWorkingDay.Sub(now).Hours() / 24)
	var employeeAvailabilityDates = make([]EmployeeAvailability, datesAmount)

	for i := 0; i < datesAmount; i++ {
		employeeAvailabilityDates[i] = EmployeeAvailability{
			u.ID,
			EmployeeAvailabilityStatusOther,
			now.Add(time.Duration(i) * 24 * time.Hour),
			"Before first working day",
		}
	}

	return employeeAvailabilityDates, nil
}

type EmployeeAvailability struct {
	UserID uuid.UUID                  `gorm:"type:binary(16); not null"                                                  json:"id"`
	Status employeeAvailabilityStatus `gorm:"type:enum('Unavailable','Sick','Other','Busy'); not null"             json:"status"`
	Date   time.Time                  `gorm:"type:datetime(3);not null;default:CURRENT_TIMESTAMP(3)" json:"date" json:"date"`

	Comment string `gorm:"type:varchar(500)"                                                    json:"comment;omitempty"`
}

type employeeAvailabilityStatus string

const (
	EmployeeAvailabilityStatusUnavailable employeeAvailabilityStatus = "Unavailable"
	EmployeeAvailabilityStatusSick        employeeAvailabilityStatus = "Sick"
	EmployeeAvailabilityStatusOther       employeeAvailabilityStatus = "Other"
	EmployeeAvailabilityStatusBusy        employeeAvailabilityStatus = "Busy"
)

type EmployeeAvailabilityJSON struct {
	UserID  string                     `json:"id"`
	Status  employeeAvailabilityStatus `json:"status"`
	Date    time.Time                  `json:"date"`
	Comment string                     `json:"comment;omitempty"`
}

func (status employeeAvailabilityStatus) IsValid() bool {
	switch status {
	case EmployeeAvailabilityStatusUnavailable,
		EmployeeAvailabilityStatusSick,
		EmployeeAvailabilityStatusOther,
		EmployeeAvailabilityStatusBusy:
		return true
	default:
		return false
	}
}

//----------------- Migrations ----------------------

func MigrateUser(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&EmployeeAvailability{},
	)
}

// -----------------HyperMedia------------------------

// ----------------strcuct-manipulations------------------
type UserSimplified struct {
	ID          uuid.UUID `json:"id" form:"id"`
	FirstName   string    `json:"firstName" form:"firstName"`
	LastName    string    `json:"lastName" form:"lastName"`
	DateOfBirth string    `json:"dateOfBirth" form:"dateOfBirth"`
	PhoneNumber string    `json:"phoneNumber" form:"phoneNumber"`
	Email       string    `json:"email" form:"email"`
	ImageUrl    string    `json:"imageUrl" form:"imageUrl"`
	Sex         string    `json:"sex" form:"sex"`
}

func (user User) Simplify() UserSimplified {
	return UserSimplified{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DateOfBirth: user.DateOfBirth.Format("2006-01-02"),
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
		ImageUrl:    user.ImageUrl,
	}
}

func SimplifyUsers(users []User) []UserSimplified {
	var simplifiedUsers = make([]UserSimplified, len(users))
	for i, user := range users {
		simplifiedUsers[i] = user.Simplify()
	}
	return simplifiedUsers
}

type RegistrantionUser struct {
	UserSimplified
	Password string `json:"password" form:"password"`
}

type RegistrantionEmployee struct {
	RegistrantionUser
	Starts string `json:"starts" form:"starts"`
}

func (ru RegistrantionUser) ToUser(role auth.Role) User {
	var params rfc7807.InvalidParams

	dateOfBirth, err := time.Parse("2006-01-02", ru.DateOfBirth)
	if err != nil {
		params.SetInvalidParam("dateOfBirth", err.Error())
	}

	return User{
		ID:          ru.ID,
		FirstName:   ru.FirstName,
		LastName:    ru.LastName,
		DateOfBirth: dateOfBirth,
		PhoneNumber: ru.PhoneNumber,
		Email:       ru.Email,
		ImageUrl:    ru.ImageUrl,
		Role:        userRole{role},
		Password:    ru.Password,
	}
}

func (re RegistrantionEmployee) ToUser(role auth.Role) (User, time.Time, rfc7807.InvalidParams) {
	var params rfc7807.InvalidParams

	dateOfBirth, err := time.Parse("2006-01-02", re.DateOfBirth)
	if err != nil {
		params.SetInvalidParam("dateOfBirth", err.Error())
	}

	starts, err := time.Parse("2006-01-02", re.Starts)
	if err != nil {
		params.SetInvalidParam("starts", err.Error())
	}

	return User{
		ID:          re.ID,
		FirstName:   re.FirstName,
		LastName:    re.LastName,
		DateOfBirth: dateOfBirth,
		PhoneNumber: re.PhoneNumber,
		Email:       re.Email,
		ImageUrl:    re.ImageUrl,
		Role:        userRole{role},
		Password:    re.Password,
	}, starts, params
}

func NewForGoogleOAUTH(email, firstName, lastName string, dateOfBirth time.Time) User {
	return User{
		ID:          uuid.New(),
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		DateOfBirth: dateOfBirth,
		ImageUrl:    config.APIURL() + "/imgs/guest-female.png",
		Role:        userRole{auth.Customer},
	}
}

func ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	return nil
}
