package user

import (
	"bytes"
	"encoding/json"
	"maryan_api/pkg/log"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type userServiceStub struct {
	loginFunc     func(email, password string) (string, error)
	loginJWTFunc  func(id uuid.UUID, email string) (string, error)
	usersDataFunc func(id uuid.UUID) (User, error)
	serviceFunc   func() userService
	secretKeyFunc func() []byte
}

func (s userServiceStub) login(email, password string) (string, error) {
	return s.loginFunc(email, password)
}

func (s userServiceStub) loginJWT(id uuid.UUID, email string) (string, error) {
	return s.loginJWTFunc(id, email)
}

func (s userServiceStub) usersData(id uuid.UUID) (User, error) {
	return s.usersDataFunc(id)
}

func (s userServiceStub) userService() userService {
	return s.serviceFunc()
}

func (s userServiceStub) secretKey() []byte {
	return s.secretKeyFunc()
}

func newGinTestHandler(register func(group *gin.RouterGroup)) *httptest.Server {
	server := gin.Default()
	group := server.Group("/test")
	group.Use(func(ctx *gin.Context) {
		ctx.Set("userID", uuid.New())
		ctx.Set("log", log.Mock())
	})

	register(group)

	return httptest.NewServer(server)
}
func Test_userHandler_login(t *testing.T) {
	testData := []struct {
		name                   string
		requestBody            map[string]any
		expectedBody           map[string]any
		expectedResponseStatus int
		loginServiceFunc       func(email, password string) (string, error)
	}{
		{"Missing Email In Request Body",
			map[string]any{"password": "password"},
			rfc7807.BadRequest(
				"login-creadentials-parsing",
				"Credentials Parsing Error",
				"Key: 'Email' Error:Field validation for 'Email' failed on the 'required' tag",
			).Map(),
			400,
			func(email, password string) (string, error) {
				return "", nil
			},
		},
		{"Missing Password In Request Body",
			map[string]any{"email": "email@test.go"},
			rfc7807.BadRequest(
				"login-creadentials-parsing",
				"Credentials Parsing Error",
				"Key: 'Password' Error:Field validation for 'Password' failed on the 'required' tag",
			).Map(),
			400,
			func(email, password string) (string, error) {
				return "", nil
			},
		},

		{"Service Error",
			map[string]any{"password": "password", "email": "email@test.go"},
			rfc7807.BadRequest("error-type",
				"Error Title",
				"Error Detail").Map(),
			400,
			func(email, password string) (string, error) {
				return "", rfc7807.BadRequest("error-type",
					"Error Title",
					"Error Detail")
			},
		},

		{"Successful Request",
			map[string]any{"password": "password", "email": "email@test.go"},
			map[string]any{"token": "token"},
			200,
			func(email, password string) (string, error) {
				return "token", nil
			},
		},
	}

	for _, testCase := range testData {
		t.Run(testCase.name, func(t *testing.T) {
			ts := newGinTestHandler(func(group *gin.RouterGroup) {
				group.POST("/login", userHandler{userServiceStub{loginFunc: testCase.loginServiceFunc}}.login)
			})

			reqBodyJSON, _ := json.Marshal(testCase.requestBody)
			resp, err := http.Post(ts.URL+"/test/login", "application/json", bytes.NewReader(reqBodyJSON))
			if err != nil {
				t.Fatalf("Expected response, got error '%s'", err.Error())
			}

			if resp.StatusCode != testCase.expectedResponseStatus {
				t.Errorf("Expected response status code to be '%v', got '%v'", testCase.expectedResponseStatus, resp.StatusCode)
			}

			defer resp.Body.Close()

			var respBody map[string]any
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			if err != nil {
				t.Fatalf("No error during parsing response body has to occure, got '%s'", err.Error())
			}

			if resp.StatusCode >= 400 {
				if diff, err := rfc7807.CmpMaps(respBody, testCase.expectedBody); err == nil {
					if diff != "" {
						t.Error(diff)
					}
				} else {
					t.Fatal(err.Error())
				}
			}

		})
	}
}
