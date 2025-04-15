package http

import (
	"maryan_api/internal/domain/user/service"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"
	"time"

	"github.com/d3code/uuid"
	"github.com/gin-gonic/gin"
)

type userHandler struct {
	service service.UserService
}

func (uh *userHandler) login(ctx *gin.Context) {
	var credentials struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ginutil.HandlerProblemAbort(
			ctx,
			rfc7807.BadRequest(
				"login-creadentials-parsing",
				"Credentials Parsing Error",
				err.Error(),
			),
		)
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	token, err := uh.service.Login(ctxWithTimeout, credentials.Email, credentials.Password)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Token string `json:"token"`
	}{
		ginutil.Response{
			"The user has been successfuly logged in.",
			hypermedia.Links{
				deleteUserLink,
				getUserLink,
			},
		},
		token,
	})
}

func (uh *userHandler) loginJWT(ctx *gin.Context) {
	id := ctx.MustGet("userID").(uuid.UUID)
	email := ctx.MustGet("email").(string)

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	token, err := uh.service.LoginJWT(ctxWithTimeout, id, email)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Token string `json:"token"`
	}{
		ginutil.Response{
			"The user has been successfuly logged in.",
			hypermedia.Links{
				deleteUserLink,
				getUserLink,
			},
		},
		token,
	})
}

// Declaration Function
func newUserHandler(service service.UserService) userHandler {
	return userHandler{service}
}
