package valueobject

import (
	"fmt"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"
	"regexp"
	"time"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
)

type EmailVerificationSession struct {
	ID      uuid.UUID `gorm:"type:binary(16);primaryKey"         json:"id"`
	Code    string    `gorm:"type:char(6);not null" json:"code"`
	Email   string    `gorm:"type:varchar(255);not null" json:"email"`
	Expires time.Time `gorm:"not null" json:"expires"`
}

type NumberVerificationSession struct {
	ID      uuid.UUID `gorm:"type:binary(16);primaryKey"         json:"id"`
	Code    string    `gorm:"type:char(6);not null" json:"code"`
	Number  string    `gorm:"type:varchar(15);not null" json:"number"`
	Expires time.Time `gorm:"not null" json:"expires"`
}

func ValidateVerificationCode(code string) error {
	var errMessage string

	if length := len(code); length != 6 {
		errMessage = fmt.Sprintf("Invalid cpde length. Want 6, got '%d'. ", length)
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(string(code)) {
		errMessage += fmt.Sprintf("The code has to only contain digits, got '%s'.", code)
	}

	if errMessage != "" {
		return rfc7807.New(http.StatusUnprocessableEntity, "invalid-verificaiton-code-format", "Code Forman Error", errMessage)
	}

	return nil
}

func NewEmailVerificationSession(code, email string) EmailVerificationSession {
	return EmailVerificationSession{uuid.New(), code, email, time.Now().Add(time.Minute * 10)}
}

func NewNumberVerificationSession(code, number string) NumberVerificationSession {
	return NumberVerificationSession{uuid.New(), code, number, time.Now().Add(time.Minute * 10)}
}

func MigrateVerifications(db *gorm.DB) error {
	return db.AutoMigrate(
		&NumberVerificationSession{},
		&EmailVerificationSession{},
	)
}
