package auth

import (
	"maryan_api/pkg/log"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authorize(secretKey []byte) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger := c.MustGet("logger").(log.Logger)
		token := c.Request.Header.Get("Authorization")

		if token == "" {
			problem := rfc7807.Unauthorized("unauthorized", "Unauthorized", `Missing "Authorization" header`)
			logger.SetProblem(problem)
			c.AbortWithStatusJSON(http.StatusUnauthorized, problem)
			return
		}

		id, email, role, err := verifyUserToken(token, secretKey)

		if err != nil {
			err := rfc7807.Unauthorized("unauthorized", "Unauthorized", err.Error())
			logger.SetProblem(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, err)
			return
		}

		c.Set("userID", id)
		c.Set("email", email)
		c.Set("role", role)

		c.Next()
	}
}
