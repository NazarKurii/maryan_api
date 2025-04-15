package auth

import (
	"fmt"
	rfc7807 "maryan_api/pkg/problem"
)

func problem(role string, err error) rfc7807.Problem {
	return rfc7807.Internal("JWT Token Generation Error", fmt.Sprintf("Could not generate %s token: %s", role, err.Error()))
}
