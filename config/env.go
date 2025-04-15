package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig loads the .env file
func LoadConfig(path string) {
	if err := godotenv.Load(path); err != nil {
		log.Fatal("Could not load env: ", err.Error())
	}
}

// mustGetEnv returns the environment variable value or panics if not set
func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("COULD NOT GET " + key)
	}
	return val
}

// mustGetEnvBytes returns the environment variable as []byte or panics if not set
func mustGetEnvBytes(key string) []byte {
	return []byte(mustGetEnv(key))
}

// Your getters:
func APIURL() string {
	return mustGetEnv("API_URL")
}

func GuestCustomerSecretKey() []byte {
	return mustGetEnvBytes("GUEST_CUSTOMER_SECRET_KEY")
}

func CustomerSecretKey() []byte {
	return mustGetEnvBytes("CUSTOMER_SECRET_KEY")
}

func EmailCodeVerificationTokenSecretKey() []byte {
	return mustGetEnvBytes("EMAIL_CODE_VERIFICATION_TOKEN_SECRET_KEY")
}

func EmailAccessTokenSecretKey() []byte {
	return mustGetEnvBytes("EMAIL_ACCESS_TOKEN_SECRET_KEY")
}

func EmailChangePasswordCodeVerificationTokenSecretKey() []byte {
	return mustGetEnvBytes("EMAIL_CHANGE_PASSWORD_CODE_VERIFICATION_TOKEN_SECRET_KEY")
}

func EmailChangePasswordAccessTokenSecretKey() []byte {
	return mustGetEnvBytes("EMAIL_CHANGE_PASSWORD_ACCESS_TOKEN_SECRET_KEY")
}

func EmailCodeSecretKeyCustomerUpdate() []byte {
	return mustGetEnvBytes("EMAIL_CODE_SECRET_KEY_CUSTOMER_UPDATE")
}

func SecretKeyCustomerUpdate() []byte {
	return mustGetEnvBytes("SECRET_KEY_CUSTOMER_UPDATE")
}

func AdminSecretKey() []byte {
	return mustGetEnvBytes("ADMIN_SECRET_KEY")
}

func DriverSecretKey() []byte {
	return mustGetEnvBytes("DRIVER_SECRET_KEY")
}

func SupportEmployeeSecretKey() []byte {
	return mustGetEnvBytes("SUPPORT_EMPLOYEE_SECRET_KEY")
}

func NumberAccessTokenSecretKey() []byte {
	return mustGetEnvBytes("NUMBER_ACCESS_TOKEN_SECRET_KEY")
}

func StripSekretKey() string {
	return mustGetEnv("STRIPE_SECRET_KEY")
}

type db struct {
	User     string
	Password string
	Host     string
	Name     string
	Port     string
}

func DB() db {
	return db{
		User:     mustGetEnv("DB_USER"),
		Password: mustGetEnv("DB_PASSWORD"),
		Host:     mustGetEnv("DB_HOST"),
		Name:     mustGetEnv("DB_NAME"),
		Port:     mustGetEnv("DB_PORT"),
	}
}

func FrontendURL() string {
	return mustGetEnv("FRONTEND_URL")
}
