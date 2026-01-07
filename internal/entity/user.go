package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"strings"
	"time"

	"github.com/nazarkurii/marshrutka_api/config"
	"github.com/nazarkurii/marshrutka_api/pkg/auth"
	rfc7807 "github.com/nazarkurii/marshrutka_api/pkg/problem"
	"github.com/nazarkurii/marshrutka_api/pkg/security"

	"github.com/asaskevich/govalidator"
	"github.com/d3code/uuid"

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
		return "", errors.New("Role is a nil interface")
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
	if dateOfBirth.Before(time.Now().Add(-18*time.Hour*24*365 + 4)) {
		dateOfBirth.Add(time.Hour*24*365 + 4)
	}
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

func createTestData(db gorm.DB) {

}

func TestDrivers() []User {
	now := time.Now()
	return []User{
		{
			ID:          uuid.New(),
			FirstName:   "John",
			LastName:    "Doe",
			DateOfBirth: time.Date(1990, 3, 15, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000001",
			Email:       "john.doe1@example.com",
			Password:    "hashedpassword1",
			ImageUrl:    "https://example.com/images/john1.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Emily",
			LastName:    "Smith",
			DateOfBirth: time.Date(1988, 7, 22, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000002",
			Email:       "emily.smith@example.com",
			Password:    "hashedpassword2",
			ImageUrl:    "https://example.com/images/emily.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Michael",
			LastName:    "Brown",
			DateOfBirth: time.Date(1992, 1, 5, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000003",
			Email:       "michael.brown@example.com",
			Password:    "hashedpassword3",
			ImageUrl:    "https://example.com/images/michael.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Sophia",
			LastName:    "Johnson",
			DateOfBirth: time.Date(1995, 11, 3, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000004",
			Email:       "sophia.johnson@example.com",
			Password:    "hashedpassword4",
			ImageUrl:    "https://example.com/images/sophia.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "James",
			LastName:    "Williams",
			DateOfBirth: time.Date(1985, 4, 17, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000005",
			Email:       "james.williams@example.com",
			Password:    "hashedpassword5",
			ImageUrl:    "https://example.com/images/james.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Olivia",
			LastName:    "Davis",
			DateOfBirth: time.Date(1991, 9, 12, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000006",
			Email:       "olivia.davis@example.com",
			Password:    "hashedpassword6",
			ImageUrl:    "https://example.com/images/olivia.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Daniel",
			LastName:    "Miller",
			DateOfBirth: time.Date(1987, 2, 9, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000007",
			Email:       "daniel.miller@example.com",
			Password:    "hashedpassword7",
			ImageUrl:    "https://example.com/images/daniel.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Ava",
			LastName:    "Martinez",
			DateOfBirth: time.Date(1993, 6, 18, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000008",
			Email:       "ava.martinez@example.com",
			Password:    "hashedpassword8",
			ImageUrl:    "https://example.com/images/ava.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "William",
			LastName:    "Garcia",
			DateOfBirth: time.Date(1996, 8, 30, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000009",
			Email:       "william.garcia@example.com",
			Password:    "hashedpassword9",
			ImageUrl:    "https://example.com/images/william.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Mia",
			LastName:    "Rodriguez",
			DateOfBirth: time.Date(1994, 12, 21, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000010",
			Email:       "mia.rodriguez@example.com",
			Password:    "hashedpassword10",
			ImageUrl:    "https://example.com/images/mia.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Ethan",
			LastName:    "Clark",
			DateOfBirth: time.Date(1990, 5, 10, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000011",
			Email:       "ethan.clark@example.com",
			Password:    "hashedpassword11",
			ImageUrl:    "https://example.com/images/ethan.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Isabella",
			LastName:    "Lewis",
			DateOfBirth: time.Date(1992, 2, 28, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000012",
			Email:       "isabella.lewis@example.com",
			Password:    "hashedpassword12",
			ImageUrl:    "https://example.com/images/isabella.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Alexander",
			LastName:    "Walker",
			DateOfBirth: time.Date(1989, 8, 16, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000013",
			Email:       "alexander.walker@example.com",
			Password:    "hashedpassword13",
			ImageUrl:    "https://example.com/images/alexander.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Charlotte",
			LastName:    "Hall",
			DateOfBirth: time.Date(1994, 3, 7, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000014",
			Email:       "charlotte.hall@example.com",
			Password:    "hashedpassword14",
			ImageUrl:    "https://example.com/images/charlotte.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Benjamin",
			LastName:    "Allen",
			DateOfBirth: time.Date(1987, 12, 1, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000015",
			Email:       "benjamin.allen@example.com",
			Password:    "hashedpassword15",
			ImageUrl:    "https://example.com/images/benjamin.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Amelia",
			LastName:    "Young",
			DateOfBirth: time.Date(1993, 9, 19, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000016",
			Email:       "amelia.young@example.com",
			Password:    "hashedpassword16",
			ImageUrl:    "https://example.com/images/amelia.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Henry",
			LastName:    "Hernandez",
			DateOfBirth: time.Date(1986, 7, 23, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000017",
			Email:       "henry.hernandez@example.com",
			Password:    "hashedpassword17",
			ImageUrl:    "https://example.com/images/henry.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Grace",
			LastName:    "King",
			DateOfBirth: time.Date(1995, 11, 8, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000018",
			Email:       "grace.king@example.com",
			Password:    "hashedpassword18",
			ImageUrl:    "https://example.com/images/grace.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Samuel",
			LastName:    "Wright",
			DateOfBirth: time.Date(1991, 1, 14, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000019",
			Email:       "samuel.wright@example.com",
			Password:    "hashedpassword19",
			ImageUrl:    "https://example.com/images/samuel.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			FirstName:   "Chloe",
			LastName:    "Lopez",
			DateOfBirth: time.Date(1988, 6, 25, 0, 0, 0, 0, time.UTC),
			PhoneNumber: "+15550000020",
			Email:       "chloe.lopez@example.com",
			Password:    "hashedpassword20",
			ImageUrl:    "https://example.com/images/chloe.jpg",
			Role:        userRole{auth.Driver},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}
