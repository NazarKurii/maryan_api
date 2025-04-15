package user

import (
	"errors"
	"fmt"
	"maryan_api/config"
	"maryan_api/pkg/auth"
	rfc7807 "maryan_api/pkg/problem"
	"maryan_api/pkg/security"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepoStub struct {
	databaseFunc func() *gorm.DB
	repoFunc     func() *userRepoMySQL
	getByIDFunc  func(id uuid.UUID) (User, error)
	loginFunc    func(email string) (uuid.UUID, string, error)
	loginJWTFunc func(email string, id uuid.UUID) (bool, error)
}

func (s userRepoStub) database() *gorm.DB {
	return s.databaseFunc()
}

func (s userRepoStub) repo() *userRepoMySQL {
	return s.repoFunc()
}

func (s userRepoStub) getByID(id uuid.UUID) (User, error) {
	return s.getByIDFunc(id)
}

func (s userRepoStub) login(email string) (uuid.UUID, string, error) {
	return s.loginFunc(email)
}

func (s userRepoStub) loginJWT(email string, id uuid.UUID) (bool, error) {
	return s.loginJWTFunc(email, id)
}

func initServiceTest() {
	config.LoadConfig("../../../.env")
}

func Test_userServiceImpl_login(t *testing.T) {
	testData := []struct {
		name           string
		email          string
		password       string
		resultTestFunc func(t *testing.T, result, email string, id uuid.UUID, roleSecretKey []byte)
		expectedError  error
		repoLogin      func(email string) (uuid.UUID, string, error)
	}{
		{
			"Invalid Email",
			"invalid.email.go",
			"",
			func(t *testing.T, result, email string, id uuid.UUID, roleSecretKey []byte) {
				if result != "" {
					t.Errorf("Expected result to be an empty string, got '%s'", result)
				}
			},
			rfc7807.BadRequest(
				"email-invalid",
				"Invald Email",
				"Provided email contains forbidden characters or is not an email at all"),
			nil,
		},
		{
			"Repo Func Error",
			"valid@email.go",
			"",
			func(t *testing.T, result, email string, id uuid.UUID, roleSecretKey []byte) {
				if result != "" {
					t.Errorf("Expected result to be an empty string, got '%s'", result)
				}
			},
			errors.New("Repo func error"),
			func(email string) (uuid.UUID, string, error) {
				return uuid.Nil, "", errors.New("Repo func error")
			},
		},
		{
			"Invalid Password",
			"valid@email.go",
			"incorrectPassword",
			func(t *testing.T, result, email string, id uuid.UUID, roleSecretKey []byte) {
				if result != "" {
					t.Errorf("Expected result to be an empty string, got '%s'", result)
				}
			},
			rfc7807.Unauthorized(
				"invalid-password",
				"Invalid Password",
				"Invalid password for user assosiated with provided email",
			),
			func(email string) (uuid.UUID, string, error) {
				return uuid.Nil, "correctPasswordhashed", nil
			},
		},
		{
			"Success case",
			"valid@email.go",
			"correctPasswordhashed",
			func(t *testing.T, result, email string, id uuid.UUID, roleSecretKey []byte) {
				tokenID, tokenEmail, err := auth.VerifyToken(result, roleSecretKey)
				if err != nil {
					t.Fatalf("Expected token to be valid insted got an error during verification: %s", err.Error())
				}

				if tokenID.String() != id.String() {
					t.Errorf("Expected id to be '%s', got '%s'", id, tokenID)
				}

				if email != tokenEmail {
					t.Errorf("Expected email to be '%s', got '%s'", email, tokenEmail)
				}
			},
			nil,
			func(email string) (uuid.UUID, string, error) {
				hashedPassword, _ := security.HashPassword("correctPasswordhashed")
				return uuid.Nil, hashedPassword, nil
			},
		},
	}

	testDataRoles := []auth.Role{
		auth.Admin,
		auth.Customer,
		auth.Driver,
		auth.SupportEmployee,
	}

	for _, role := range testDataRoles {
		for _, testCase := range testData {
			t.Run(fmt.Sprintf("%s: %s", role.Role(), testCase.name), func(t *testing.T) {
				userService := userServiceImpl{userRepoStub{loginFunc: testCase.repoLogin}, role}
				result, err := userService.login(testCase.email, testCase.password)

				if err != nil {
					if testCase.expectedError == nil {
						t.Errorf("Expecter error to be nil, got '%s'", err.Error())
					} else if err.Error() != testCase.expectedError.Error() {
						t.Errorf("Expected error to be '%s', got '%s'", testCase.expectedError.Error(), err.Error())
					}
				} else if testCase.expectedError != nil {
					t.Errorf("Expecter error not to be '%s', got nil", err.Error())
				}

				var id uuid.UUID
				if testCase.repoLogin != nil {
					id, _, _ = testCase.repoLogin(testCase.email)
				}

				testCase.resultTestFunc(t, result, testCase.email, id, userService.secretKey())
			})
		}

	}
}
