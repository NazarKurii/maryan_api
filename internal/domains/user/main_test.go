package user

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	initServiceTest()

	os.Exit(m.Run())
}
