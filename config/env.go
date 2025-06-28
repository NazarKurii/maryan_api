package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type API struct {
	Key string
	URL string
}

func LoadConfig(path string) {
	err := godotenv.Load(path)
	if err != nil {
		log.Fatal("Could not load env: ", err.Error())
	}

}

func APIURL() string {
	url := os.Getenv("API_URL")
	if url == "" {
		panic("Could not load customer secret key")
	}
	return url
}

func GuestCustomerSecretKey() []byte {
	key := os.Getenv("GUEST_CUSTOMER_SECRET_KEY")
	if key == "" {
		panic("Could not load customer secret key")
	}
	return []byte(key)
}

func CustomerSecretKey() []byte {
	key := os.Getenv("CUSTOMER_SECRET_KEY")
	if key == "" {
		panic("Could not load customer secret key")
	}
	return []byte(key)
}

func AdminSecretKey() []byte {
	key := os.Getenv("ADMIN_SECRET_KEY")
	if key == "" {
		panic("Could not load customer secret key")
	}
	return []byte(key)
}

func DriverSecretKey() []byte {
	key := os.Getenv("DRIVER_SECRET_KEY")
	if key == "" {
		panic("Could not load customer secret key")
	}
	return []byte(key)
}
func SupportEmployeeSecretKey() []byte {
	key := os.Getenv("SUPPORT_EMPLOYEE_SECRET_KEY")
	if key == "" {
		panic("Could not load customer secret key")
	}
	return []byte(key)
}

func EmailAccessTokenSecretKey() []byte {
	key := os.Getenv("EMAIL_ACCESS_TOKEN_SECRET_KEY")
	if key == "" {
		panic("Could not load customer secret key")
	}
	return []byte(key)
}

func NumberAccessTokenSecretKey() []byte {
	key := os.Getenv("NUMBER_ACCESS_TOKEN_SECRET_KEY")
	if key == "" {
		panic("Could not load customer secret key")
	}
	return []byte(key)
}
