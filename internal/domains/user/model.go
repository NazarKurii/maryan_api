package user

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

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
	Guest       bool         `gorm:"not null" json:"guest"`
	Password    string       `gorm:"type:varchar(255);not null" json:"password"  binding:"required"`
	ImageUrl    string       `gorm:"type:varchar(255);not null" json:"imageUrl"`
	Role        userRole     `gorm:"type:enum('Customer','Admin','Driver','Support Employee');not null" json:"role"`
	CreatedAt   time.Time    `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time    `gorm:"not null" json:"updatedAt"`
	DeletedAt   sql.NullTime `gorm:"index" json:"deletedAt"`
}

// USER -> DATEOFBIRTH
type dateOfBirth struct {
	time.Time
}

func (dob *dateOfBirth) UnmarshalJSON(b []byte) error {
	// b is a JSON string literal, e.g. `"1990-05-15"`
	s := strings.Trim(string(b), `"`) // remove quotes

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
	Role auth.Role
}

func (ur userRole) MarshalJSON() ([]byte, error) {
	return json.Marshal((ur.Role.Role()))
}

func (ur userRole) Value() (driver.Value, error) {
	if ur.Role == nil {
		return nil, errors.New("Role is a nil interface")
	}
	return ur.Role.Role(), nil

}
func (ur *userRole) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		var err error
		ur.Role, err = auth.DefineRole(v)
		return err
	case []byte:
		str := string(v)
		var err error
		ur.Role, err = auth.DefineRole(str)
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

func (u *User) validate() ([]rfc7807.InvalidParam, bool) {
	params, add, isNil := rfc7807.StartSettingInvalidParams()

	if len(u.FirstName) < 1 {
		add("firstName", "Cannot be blank.")
	}

	if len(u.LastName) < 1 {
		add("lastName", "Cannot be blank.")
	}

	if u.DateOfBirth.Before(time.Now().AddDate(-125, 0, 0)) {
		add("dateOfBirth", "Has to be greater or equal to 18.")
	}

	if !govalidator.IsEmail(u.Email) {
		add("email", "Contains invalid characters or is not an email.")
	}

	if len(u.Password) < 6 {
		add("password", "Has to be at least 6 characters long.")
	}

	if !govalidator.HasUpperCase(u.Password) {
		add("password", "Has to contain at least one uppercase letter.")
	}

	if !strings.ContainsAny(u.Password, "0123456789") {
		add("password", "Has to contain at least one number.")
	}

	return *params, isNil()
}
