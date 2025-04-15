package security

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(hashedPassword), err
}

func VerifyPassword(password, hasedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hasedPassword), []byte(password)) == nil
}
