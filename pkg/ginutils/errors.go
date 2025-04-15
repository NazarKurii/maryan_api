package ginutil

import (
	"fmt"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ServiceErrorAbort(ctx *gin.Context, err error) {
	logger := getLogger(ctx)
	problem, ok := rfc7807.Is(err)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, rfc7807.Internal("Could not convert error into rfc7807 representaion", fmt.Sprintf("Error message: %s", err.Error())))
		logger.SetError(err, http.StatusInternalServerError)
	} else {
		ctx.AbortWithStatusJSON(problem.Status, problem)
		logger.SetProblem(problem)
	}

}

func HandlerProblemAbort(ctx *gin.Context, problem rfc7807.Problem) {
	logger := getLogger(ctx)
	ctx.AbortWithStatusJSON(problem.Status, problem)
	logger.SetProblem(problem)
}
