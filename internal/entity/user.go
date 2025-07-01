package entity

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

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
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;" json:"id" `
	FirstName   string       `gorm:"type:varchar(50);not null" json:"firstName"  binding:"required"`
	LastName    string       `gorm:"type:varchar(50);not null" json:"lastName"  binding:"required"`
	DateOfBirth dateOfBirth  `gorm:"not null" json:"dateOfBirth"  binding:"required"`
	PhoneNumber string       `gorm:"type:varchar(15);not null" json:"phoneNumber"  binding:"required"`
	Email       string       `gorm:"type:varchar(255);not null;unique" json:"email"  binding:"required"`
	Password    string       `gorm:"type:varchar(255);not null" json:"password"  binding:"required"`
	ImageUrl    string       `gorm:"type:varchar(255);not null" json:"imageUrl"`
	Sex         userSex      `gorm:"type:enum('Female','Male');not null" json:"sex"`
	Role        userRole     `gorm:"type:enum('Customer','Admin','Driver','Support');not null" json:"role"`
	CreatedAt   time.Time    `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time    `gorm:"not null" json:"updatedAt"`
	DeletedAt   sql.NullTime `gorm:"index" json:"deletedAt"`
}

// USER -> DATEOFBIRTH
type dateOfBirth struct {
	time.Time
}

func (dob *dateOfBirth) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	// parse using the correct layout "2006-01-02"
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	dob.Time = t
	return nil
}

func (dob *dateOfBirth) Scan(value interface{}) error {
	if value == nil {
		*dob = dateOfBirth{Time: time.Time{}}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*dob = dateOfBirth{Time: v}
	case []byte:
		t, err := time.Parse("2006-01-02", string(v))
		if err != nil {
			return err
		}
		*dob = dateOfBirth{Time: t}
	case string:
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return err
		}
		*dob = dateOfBirth{Time: t}
	default:
		return fmt.Errorf("expected time.Time or string for dateOfBirth scan, got %T", value)
	}
	return nil
}

func (dob dateOfBirth) Value() (driver.Value, error) {
	if dob.IsZero() {
		return nil, nil
	}
	return dob.Time.Format("2006-01-02"), nil
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
func (u *User) fomratPhoneNumber() error {
	pn, err := phonenumbers.Parse(u.PhoneNumber, "UA")
	if err != nil {
		return err
	}

	if !phonenumbers.IsValidNumber(pn) {
		return errors.New("invalid phone number")
	}
	u.PhoneNumber = phonenumbers.Format(pn, phonenumbers.E164)
	return nil
}

func (u *User) validate() rfc7807.InvalidParams {
	var params rfc7807.InvalidParams

	if len(u.FirstName) < 1 {
		params.SetInvalidParam("firstName", "Cannot be blank.")
	}

	if len(u.LastName) < 1 {
		params.SetInvalidParam("lastName", "Cannot be blank.")
	}

	if u.Sex != maleSex && u.Sex != femaleSex {
		params.SetInvalidParam("sex", "Can only be 'Male' or 'Female'")
	}

	if u.DateOfBirth.Before(time.Now().AddDate(-125, 0, 0)) {
		params.SetInvalidParam("dateOfBirth", "Has to be greater or equal to 18.")
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

//----------------- Migrations ----------------------

func MigrateUser(db *gorm.DB) {
	db.AutoMigrate(
		&User{},
	)
}

// -----------------HyperMedia------------------------

// ----------------strcuct-manipulations------------------
type UserSimplified struct {
	ID          uuid.UUID `json:"id"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	DateOfBirth string    `json:"dateOfBirth"`
	PhoneNumber string    `json:"phoneNumber"`
	Email       string    `json:"email"`
	ImageUrl    string    `json:"imageUrl"`
	Sex         userSex   `json:"sex"`
}

func (user User) ToSimplified() UserSimplified {
	return UserSimplified{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DateOfBirth: user.DateOfBirth.String(),
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
		ImageUrl:    user.ImageUrl,
	}
}

func (su UserSimplified) ToUser(role auth.Role) (User, error) {
	dob, err := time.Parse(time.RFC3339, su.DateOfBirth)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:          su.ID,
		FirstName:   su.FirstName,
		LastName:    su.LastName,
		DateOfBirth: dateOfBirth{dob},
		PhoneNumber: su.PhoneNumber,
		Email:       su.Email,
		ImageUrl:    su.ImageUrl,
		Role:        userRole{role},
	}, nil
}

type RegistrantionUser struct {
	UserSimplified
	Password string `json:"password"`
}

func (ru RegistrantionUser) ToNewUser(role auth.Role) (User, rfc7807.InvalidParams) {
	user, _ := ru.UserSimplified.ToUser(role)
	user.Password = ru.Password
	user.ID = uuid.New()
	params := user.validate()

	return user, params
}

func NewForGoogleOAUTH(email, name, surname string) User {
	return User{
		ID:          uuid.New(),
		Email:       email,
		FirstName:   name,
		LastName:    surname,
		DateOfBirth: dateOfBirth{time.Now().AddDate(-25, 0, 0)},
		ImageUrl:    "https://example.com/default-guest-avatar.png",
		Role:        userRole{auth.Customer},
	}
}

type UsersPaginationStr struct {
	PageNumber string
	PageSize   string
	OrderBy    string
	OrderWay   string
	Roles      string
}

type UserPagination struct {
	PageNumber int
	PageSize   int
	Orderby    string
	Roles      []string
}

func (upstr UsersPaginationStr) Parse() (UserPagination, error) {
	var err error
	var params rfc7807.InvalidParams
	stringToInt := func(s string, name string, destination *int) {
		*destination, err = strconv.Atoi(s)
		if err != nil {
			if errors.Is(err, strconv.ErrSyntax) {
				params.SetInvalidParam(name, err.Error())
			} else {

			}
		} else if *destination < 1 {
			params.SetInvalidParam(name, "Must be equal or greater than 1.")
		}
	}

	var userPagination UserPagination

	stringToInt("pageNumber", upstr.PageNumber, &userPagination.PageNumber)
	stringToInt("pageSize", upstr.PageSize, &userPagination.PageSize)

	switch upstr.OrderBy {
	case "name", "role", "age", "registrationDate":
		userPagination.Orderby = upstr.OrderBy
	default:
		params.SetInvalidParam("orderBy", "non-existing orderBy value.")
	}

	switch upstr.OrderWay {
	case "DESC", "ASC":
		upstr.OrderBy += " " + upstr.OrderWay
	default:
		params.SetInvalidParam("orderWay", "non-existing orderWay value.")
	}

	roles := strings.Split(upstr.Roles, "+")
	for _, role := range roles {
		switch role {
		case "Admin", "Customer", "Support", "Driver":
			userPagination.Roles = append(userPagination.Roles, role)
		default:
			params.SetInvalidParam("roles", "non-existing role value.")
		}

	}

	if params != nil {
		return userPagination, rfc7807.BadRequest("invalid-users-pagination-data", "Invalid Users Pagination Data Error", "Provided pagination data is not valid.", params...)
	}
	return userPagination, nil
}
