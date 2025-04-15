package google

import rfc7807 "maryan_api/pkg/problem"

func badGateway(err error) rfc7807.Problem {
	return rfc7807.BadGateway("user-credentials", "User Credentials Error", err.Error())
}

func invalidCode(err error) rfc7807.Problem {
	return rfc7807.BadRequest("invalid-code", "Invalid Code Error", err.Error())
}
