package user

import (
	"database/sql/driver"
	"fmt"
	"maryan_api/pkg/auth"
	"strings"

	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gorm.io/gorm"
)

func Test_dateOfBirthUnmarshalJSON(t *testing.T) {
	var dob dateOfBirth
	if err := dob.UnmarshalJSON([]byte("2006.09.16")); err != nil {
		t.Error(err.Error())
	}

	if err := dob.UnmarshalJSON([]byte("2006.9.16")); err == nil {
		t.Error("expected error")
	}

	if err := dob.UnmarshalJSON([]byte("2006.16.16")); err == nil {
		t.Error("expected error")
	}
}

func Test_user_formatPhoneNumber(t *testing.T) {
	testtestData := []struct {
		name           string
		testedNumber   string
		expectedResult string
		errorMessage   string
	}{
		{
			name:           "Valid E.164 UA number",
			testedNumber:   "+380501234567",
			expectedResult: "+380501234567",
			errorMessage:   "",
		},
		{
			name:           "Valid local UA number",
			testedNumber:   "0501234567",
			expectedResult: "+380501234567",
			errorMessage:   "",
		},
		{
			name:           "Invalid UA number",
			testedNumber:   "000",
			expectedResult: "",
			errorMessage:   "invalid phone number",
		},
		{
			name:           "Number with an invalid character",
			testedNumber:   "++38050123",
			expectedResult: "",
			errorMessage:   "invalid phone number",
		},
		{
			name:           "Number with an invalid character",
			testedNumber:   "+38n050123",
			expectedResult: "",
			errorMessage:   "the phone number supplied is not a number",
		},
	}

	var user User
	for _, tc := range testtestData {
		t.Run(tc.name, func(t *testing.T) {
			user.PhoneNumber = tc.testedNumber
			if err := user.fomratPhoneNumber(); err != nil {
				if tc.errorMessage == "" {
					t.Errorf("Expected error to be nil, got '%s'", err.Error())
				} else if err.Error() != tc.errorMessage {
					t.Errorf("Expected error to be '%s', got '%s", tc.errorMessage, err.Error())
				}
			} else if tc.errorMessage != "" {
				t.Errorf("Expected err to be '%s', got nil", tc.errorMessage)
			}
		})
	}
}

func Test_userBeforeCreate(t *testing.T) {
	testPassword := "testPassword123"
	user := User{Password: testPassword}

	if user.BeforeCreate(&gorm.DB{}); user.Password == testPassword {
		t.Error("The password has not been hashed")
	}

	if len(user.Password) != 60 {
		t.Errorf("The password length expected to be 60, got %v", len(user.Password))
	}

}

func Test_userRoleValue(t *testing.T) {
	testData := []struct {
		name         string
		userRole     userRole
		errorMessage string
		expected     driver.Value
	}{
		{
			"Valid Role",
			userRole{auth.Customer},
			"",
			"Customer",
		},
		{
			"Nil Role",
			userRole{nil},
			"Role is a nil interface",
			nil,
		},
	}

	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.userRole.Value()
			if tc.errorMessage != "" {
				if err == nil {
					t.Errorf("Expected error to be '%s', got nil", tc.errorMessage)
				} else if err.Error() != tc.errorMessage {
					t.Errorf("Expected error to be '%s', got '%s'", tc.errorMessage, err.Error())
				}

			} else if err != nil {
				t.Errorf("Expected error to be nil, got '%s'", err.Error())
			}

			if result != tc.expected {
				t.Errorf("Expected result to be '%v', got '%v'", tc.expected, result)
			}
		})
	}
}

func Test_userRoleScan(t *testing.T) {
	testData := []struct {
		name       string
		values     []any
		expected   []userRole
		errMessage func(value any) string
	}{{
		"Valid Value",
		[]any{"Customer", "Admin", "Driver", "Support Employee"},
		[]userRole{{auth.Customer}, {auth.Admin}, {auth.Driver}, {auth.SupportEmployee}},
		func(value any) string {
			return ""
		},
	},
		{
			"Invalid Value Type",
			[]any{1, true, []int{}, make(chan int)},
			[]userRole{{}, {}, {}, {}},
			func(value any) string {
				return fmt.Sprintf("UserRole: cannot scan type %T into string", value)
			},
		},
		{
			"Invalid Role Name",
			[]any{"InvalidRole"},
			[]userRole{{}},
			func(value any) string {
				return fmt.Sprintf("unknown role: %s", value.(string))
			},
		},
	}

	for _, testCases := range testData {
		length := len(testCases.expected)
		for i := 0; i < length; i++ {
			var userRole userRole
			value := testCases.values[i]
			resultErr := userRole.Scan(value)
			expectedError := testCases.errMessage(value)
			expected := testCases.expected[i]

			t.Run(testCases.name, func(t *testing.T) {

				if resultErr != nil {
					if expectedError == "" {
						t.Errorf("Expected error to be nil, got '%s'", resultErr.Error())
					} else if resultErr.Error() != expectedError {
						t.Errorf("Expected error to be '%s', got '%s'", expectedError, resultErr.Error())
					}
				} else if expectedError != "" {
					t.Errorf("Expected error to be '%s', got nil", expectedError)
				}

				if diff := cmp.Diff(userRole, expected); diff != "" {
					t.Error(diff)
				}

			})
		}
	}

}

func TestUserValidate(t *testing.T) {
	u := User{
		FirstName:   "",
		LastName:    "",
		DateOfBirth: dateOfBirth{time.Now().AddDate(-126, 0, 0)}, // too old
		Email:       "invalid-email",
		Password:    "abc",
	}

	params, isValid := u.validate()

	if isValid {
		t.Fatalf("expected validation to fail, but got success")
	}

	expected := map[string][]string{
		"firstName":   {"Cannot be blank."},
		"lastName":    {"Cannot be blank."},
		"dateOfBirth": {"Has to be greater or equal to 18."},
		"email":       {"Contains invalid characters or is not an email."},
		"password": {
			"Has to be at least 6 characters long.",
			"Has to contain at least one uppercase letter.",
			"Has to contain at least one number.",
		},
	}

	for _, param := range params {
		expectedReasons, ok := expected[param.Name]
		if !ok {
			t.Errorf("unexpected validation field: %s", param.Name)
			continue
		}
		found := false
		for _, reason := range expectedReasons {
			if strings.TrimSpace(reason) == strings.TrimSpace(param.Reason) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unexpected reason for field %s: %s", param.Name, param.Reason)
		}
	}

	u = User{
		FirstName:   "N",
		LastName:    "L",
		DateOfBirth: dateOfBirth{time.Now().AddDate(-124, 0, 0)},
		Email:       "email@email.email",
		Password:    "Passw0rd",
	}

	params, isValid = u.validate()

	if !isValid {
		t.Error("Expected user to be valid, got valid = false")
	}

	if params != nil {
		t.Errorf("Expected params to be nil slice, got '%v'", params)
	}
}
