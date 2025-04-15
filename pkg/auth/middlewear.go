package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maryan_api/pkg/log"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthorizeLog(secretKey []byte, db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger, doLog := createLog(db, c)
		defer doLog()
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			problem := rfc7807.Unauthorized("unauthorized", "Unauthorized", `Missing "Authorization" header`, logger.GetID())
			logger.SetProblem(problem)
			c.AbortWithStatusJSON(http.StatusUnauthorized, problem)
			return
		}

		id, email, err := VerifyToken(token, secretKey)

		if err != nil {
			err := rfc7807.Unauthorized("unauthorized", "Unauthorized", err.Error(), logger.GetID())
			logger.SetProblem(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, err)
			return
		}

		c.Set("log", logger)
		c.Set("userID", id)
		c.Set("email", email)

		c.Next()
	}
}

func createLog(db *gorm.DB, c *gin.Context) (log.Logger, func()) {

	queryParams, _ := json.Marshal(c.Request.URL.Query())
	headers, _ := json.Marshal(c.Request.Header)
	body, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	logRecord := log.New(c.ClientIP(), c.FullPath(), queryParams, headers, body, c.Request.Method)

	return &logRecord, func() {
		if err := logRecord.Do(db); err != nil {
			fmt.Printf(
				`\n
						...............LOGGING ERROR...................
						%s
						...............................................
						\n
					`, err.Error())
		}
	}

}

func Log(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger, doLog := createLog(db, c)
		defer doLog()

		c.Set("log", logger)
		c.Next()
	}
}
