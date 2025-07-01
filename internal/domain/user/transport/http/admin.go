package transport

import (
	"maryan_api/internal/domain/user/service"
	"maryan_api/internal/entity"
	ginutil "maryan_api/pkg/ginutils"
	rfc7807 "maryan_api/pkg/problem"
	"maryan_api/pkg/security"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type adminHandler struct {
	userHandler
	service service.AdminService
}

func (ah *adminHandler) users(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	users, urls, err := ah.service.Users(entity.UsersPaginationStr{
		ctx.Param("page"),
		ctx.Param("size"),
		ctx.Param("role"),
		ctx.Param("order-by"),
		ctx.Param("order-way")},
		ctxWithTimeout)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Users []entity.User `json:"users"`
	}{ginutil.Response{
		"Users have successfuly been retrieved.",
		urls,
	},
		users})
}

func (ah *adminHandler) hashPassword(c *gin.Context) {
	var password struct {
		Val string `json:"password"`
	}

	err := c.ShouldBindJSON(&password)
	if err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("invalid-password", "Invalid Passrord Error", err.Error()))
		return
	}

	hashedPassword, err := security.HashPassword(password.Val)
	if err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("invalid-password", "Invalid Passrord Error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, struct {
		ginutil.Response
		Password string `json:"password"`
	}{
		ginutil.Response{Message: "The password has successfuly been hashed"},
		hashedPassword,
	})
}

// Declaration function
func newAdminHandler(service service.AdminService) adminHandler {
	return adminHandler{userHandler: newUserHandler(service), service: service}
}
