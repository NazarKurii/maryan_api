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
)

type adminHandler struct {
	userHandler
	service service.AdminService
}

func (ah *adminHandler) users(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	users, urls, err := ah.service.Users(ctxWithTimeout, dbutil.PaginationStr{
		"admin/users",
		ctx.Param("page"),
		ctx.Param("size"),
		ctx.Param("order_by"),
		ctx.Param("order_way"),
	},
		ctx.Param("role"),
	)
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

func (ah *adminHandler) NewUser(role auth.Role) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var user entity.RegistrantionUser

		if err := ctx.ShouldBindJSON(&user); err != nil {
			ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("user-parsing", "Body Parsing Error", err.Error()))
			return
		}

		image, err := ctx.FormFile("image")
		if err != nil {
			if err.Error() != "no multipart boundary param in Content-Type" {
				ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("image-forming-error", "Image Froming Error", err.Error()))
			}
		}

		type Headers struct {
			EmailToken  string `header:"X-Email-Access-Token" binding:"required"`
			NumberToken string `header:"X-Number-Access-Token" binding:"required"`
		}

		var headers Headers
		if err := ctx.ShouldBindHeader(&headers); err != nil {
			ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("headers-parsing-error", "Headers Error", err.Error()))
			return
		}

		ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
		defer cancel()

		token, err := ah.service.NewUser(ctxWithTimeout, user, image, role)
		if err != nil {
			ginutil.ServiceErrorAbort(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, struct {
			ginutil.Response
			Token string `json:"token"`
		}{

			ginutil.Response{
				"The user has successfuly been saved.",
				[]hypermedia.Link{deleteUserLink},
			},
			token,
		})
	}
}

// Declaration function
func newAdminHandler(service service.AdminService) adminHandler {
	return adminHandler{userHandler: newUserHandler(service), service: service}
}
