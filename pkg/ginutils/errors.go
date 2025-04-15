package ginutil

import (
	"fmt"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ServiceErrorAbort(c *gin.Context, err error) {
	logger := RequestLog(c)
	problem, ok := rfc7807.Is(err)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, rfc7807.Internal("Could not convert error into rfc7807 representaion", fmt.Sprintf("Error message: %s", err.Error()), logger.GetID()))
		logger.SetError(err, http.StatusInternalServerError)
	} else {
		problem.SetInstance(logger.GetID())
		c.AbortWithStatusJSON(problem.Status, problem)
		logger.SetProblem(problem)
	}

}

func HandlerProblemAbort(c *gin.Context, problem rfc7807.Problem) {
	logger := RequestLog(c)
	problem.SetInstance(logger.GetID())
	c.AbortWithStatusJSON(problem.Status, problem)
	logger.SetProblem(problem)
}
