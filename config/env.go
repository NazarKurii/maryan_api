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
	return "http://localhost:8080"
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
