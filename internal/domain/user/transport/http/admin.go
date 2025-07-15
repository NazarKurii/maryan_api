package http

import (
	"maryan_api/internal/domain/user/service"
	"maryan_api/internal/entity"
	"maryan_api/pkg/auth"
	"maryan_api/pkg/dbutil"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"

	rfc7807 "maryan_api/pkg/problem"
	"maryan_api/pkg/security"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

type adminHandler struct {
	userHandler
	service service.AdminService
}

func (ah *adminHandler) getUsers(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	users, urls, err := ah.service.GetUsers(ctxWithTimeout, dbutil.PaginationStr{
		"/admin/users",
		ctx.DefaultQuery("page", "1"),
		ctx.DefaultQuery("size", "8"),
		ctx.DefaultQuery("order_by", "first_name"),
		ctx.DefaultQuery("order_way", "asc"),
		ctx.DefaultQuery("search", ""),
	}, ctx.DefaultQuery("roles", "admin,driver,support"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Users []entity.UserSimplified `json:"users"`
	}{ginutil.Response{
		"Users have successfuly been retrieved.",
		urls,
	},
		users})
}

func (ah *adminHandler) getUser(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	user, err := ah.service.GetUserByID(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		entity.User `json:"user"`
	}{
		ginutil.Response{
			"The user has successfuly been found.",
			hypermedia.Links{
				deleteUserLink,
			},
		},
		user,
	})
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

func (ah *adminHandler) newEmployee(role auth.Role) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var user entity.RegistrantionEmployee

		if err := ctx.ShouldBindWith(&user, binding.FormMultipart); err != nil {
			ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("user-parsing", "Body Parsing Error", err.Error()))
			return
		}

		image, err := ctx.FormFile("image")
		if err != nil {
			if err.Error() != "no multipart boundary param in Content-Type" {
				ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("image-forming-error", "Image Froming Error", err.Error()))
			}
		}

		ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
		defer cancel()

		err = ah.service.NewEmployee(ctxWithTimeout, user, image, ctx.SaveUploadedFile, role)
		if err != nil {
			ginutil.ServiceErrorAbort(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, ginutil.Response{
			"The employee has successfuly been saved.",
			hypermedia.Links{
				deleteUserLink,
			},
		})
	}
}

func (ah *adminHandler) setEmployeeAvailability(ctx *gin.Context) {
	var request struct {
		Schedule []entity.EmployeeAvailability `json:"schedule"`
	}

	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("body-parsing", "Body Parsing Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	err = ah.service.SetEmployeeAvailability(ctxWithTimeout, request.Schedule)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, ginutil.Response{
		"The users's schedule has successfuly been updated.",
		hypermedia.Links{
			getUsersLink,
		},
	})
}

func (ah *adminHandler) get(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	user, err := ah.service.GetByID(ctxWithTimeout, ctx.MustGet("userID").(uuid.UUID))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		entity.User `json:"user"`
	}{
		ginutil.Response{
			"The user has successfuly been found.",
			hypermedia.Links{},
		},
		user,
	})
}

func (ah *adminHandler) getAvailableEmployees(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	users, urls, err := ah.service.GetAvailableEmployees(ctxWithTimeout, dbutil.PaginationStr{
		"admin/available-employees",
		ctx.DefaultQuery("page", "1"),
		ctx.DefaultQuery("size", "20"),
		ctx.DefaultQuery("order_by", "id"),
		ctx.DefaultQuery("order_way", "asc"),
		ctx.DefaultQuery("search", ""),
	}, ctx.DefaultQuery("roles", "admin+driver+support"), ctx.DefaultQuery("from", ""), ctx.DefaultQuery("to", ""))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Users []entity.UserSimplified `json:"users"`
	}{ginutil.Response{
		"Users have successfuly been retrieved.",
		urls,
	},
		users})
}

func (ah *adminHandler) getFreeDrivers(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	users, urls, err := ah.service.GetFreeDrivers(ctxWithTimeout, dbutil.PaginationStr{
		"/admin/free-drivers",
		ctx.DefaultQuery("page", "1"),
		ctx.DefaultQuery("size", "20"),
		ctx.DefaultQuery("order_by", "id"),
		ctx.DefaultQuery("order_way", "asc"),
		ctx.DefaultQuery("search", ""),
	})
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Users []entity.UserSimplified `json:"users"`
	}{ginutil.Response{
		"Users have successfuly been retrieved.",
		urls,
	},
		users})
}

// Declaration function
func newAdminHandler(service service.AdminService) adminHandler {
	return adminHandler{userHandler: newUserHandler(service), service: service}
}
