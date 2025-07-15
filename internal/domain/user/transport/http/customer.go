package http

import (
	"context"
	"encoding/json"
	"fmt"
	"maryan_api/config"
	"maryan_api/internal/domain/user/service"
	"maryan_api/internal/entity"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
)

type customerHandler struct {
	userHandler
	service service.CustomerService
}

func (ch *customerHandler) verifyEmailIfExists(ctx *gin.Context) {
	var email struct {
		Val string `json:"email"`
	}

	if err := ctx.ShouldBindJSON(&email); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("email-parsing", "Body Parsing Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	token, exists, err := ch.service.VerifyEmailIfExists(ctxWithTimeout, email.Val)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	resp := struct {
		ginutil.Response
		Exists bool `json:"exists"`
	}{
		Response: ginutil.Response{
			Links: hypermedia.Links{
				verifyNumberLink,
				registerUserLink,
			},
		},
	}

	if !exists {
		resp.Message = "The code has successfully been sent."
		resp.Links.Add(
			hypermedia.Link{
				"verifyEmailCode",
				hypermedia.LinkData{
					Href:   "/verify-email-code/" + token,
					Method: http.MethodPost,
				},
			})
	} else {
		resp.Message = "Email already exists."
		resp.Exists = true
	}

	ctx.JSON(http.StatusOK, resp)
}

func (ch *customerHandler) verifyEmailChangePassword(ctx *gin.Context) {
	ch.verifyEmail(ctx, ch.service.VerifyEmailPasswordChange)
}

func (ch *customerHandler) verifyEmailCustomerUpdate(ctx *gin.Context) {

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	token, err := ch.service.VerifyEmailCustomerUpdate(ctxWithTimeout, ctx.MustGet("email").(string))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, ginutil.Response{
		Message: "The code has successfully been sent.",

		Links: hypermedia.Links{
			{
				"verifyEmailCode",
				hypermedia.LinkData{
					Href:   "/verify-update-code/" + token,
					Method: http.MethodPost,
				},
			},
		},
	})
}

func (ch *customerHandler) verifyEmail(ctx *gin.Context, serviceFunc func(context.Context, string) (string, error)) {

	var email struct {
		Val string `json:"email"`
	}

	if err := ctx.ShouldBindJSON(&email); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("email-parsing", "Body Parsing Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	token, err := serviceFunc(ctxWithTimeout, email.Val)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, ginutil.Response{
		Message: "The code has successfully been sent.",

		Links: hypermedia.Links{
			{
				"verifyEmailCode",
				hypermedia.LinkData{
					Href:   "/change-password/verify-email-code/" + token,
					Method: http.MethodPost,
				},
			},
		},
	})

}

func (ch *customerHandler) verifyEmailCodeHandler(ctx *gin.Context, serviceFunc func(ctx context.Context, code string, token string) (string, error)) {

	var code struct {
		Val string `json:"code"`
	}

	if err := ctx.ShouldBindJSON(&code); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("code-parsing", "Body Parsing Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	token, err := serviceFunc(ctxWithTimeout, code.Val, ctx.Param("token"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Token string `json:"token"`
	}{
		ginutil.Response{
			Message: "The code has successfully been verified",
			Links: hypermedia.Links{
				verifyNumberLink,
				registerUserLink,
			},
		},
		token,
	})

}

func (ch *customerHandler) verifyEmailCode(ctx *gin.Context) {
	ch.verifyEmailCodeHandler(ctx, ch.service.VerifyEmailCode)
}

func (ch *customerHandler) verifyEmailCodePasswordChanging(ctx *gin.Context) {
	ch.verifyEmailCodeHandler(ctx, ch.service.VerifyEmailCodePasswordChange)
}

func (ch *customerHandler) VerifyCustomerUpdateCode(ctx *gin.Context) {
	ch.verifyEmailCodeHandler(ctx, ch.service.VerifyCustomerUpdateCode)
}

func (ch *customerHandler) verifyNumber(ctx *gin.Context) {
	var number struct {
		Val string `json:"phoneNumber"`
	}

	if err := ctx.ShouldBindJSON(&number); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("number-parsing", "Body Parsing Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	token, err := ch.service.VerifyNumber(ctxWithTimeout, number.Val)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, ginutil.Response{
		Message: "The code has successfully been sent.",
		Links: hypermedia.Links{
			verifyEmailLink,
			registerUserLink,
			hypermedia.Link{

				"verifyNumberCode", hypermedia.LinkData{
					Href:   config.APIURL() + "/customer/verify-number-code/" + token,
					Method: http.MethodPost,
				},
			},
		},
	})
}

func (ch *customerHandler) verifyNumberCode(ctx *gin.Context) {
	var code struct {
		Val string `json:"code"`
	}

	if err := ctx.ShouldBindJSON(&code); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("code-parsing", "Body Parsing Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	token, err := ch.service.VerifyNumberCode(ctxWithTimeout, code.Val, ctx.Param("token"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Token string `json:"token"`
	}{
		ginutil.Response{
			Message: "The code has successfully been verified",
			Links: hypermedia.Links{
				verifyEmailLink,
				registerUserLink,
			},
		},
		token,
	})
}

func (ch *customerHandler) googleOAUTH(ctx *gin.Context) {
	var request struct {
		Code string `json:"code"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("google-code-parsing", "Body Parsing Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	token, isNew, err := ch.service.GoogleOAUTH(ctxWithTimeout, request.Code)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Token string `json:"token"`
		IsNew bool   `json:"isNew"`
	}{
		ginutil.Response{
			Message: "User has been logged in successfully.",
			Links: hypermedia.Links{
				deleteUserLink,
			},
		},
		token,
		isNew,
	})
}

func (ch *customerHandler) register(ctx *gin.Context) {
	userJSON := ctx.PostForm("user")
	fmt.Println(userJSON)
	var user entity.RegistrantionUser
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("user-parsing", "Body Parsing Error", err.Error()))
		return
	}

	image, err := ctx.FormFile("image")
	// if err != nil && err.Error() != "no multipart boundary param in Content-Type" {
	// 	ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("image-forming-error", "Image Forming Error", err.Error()))
	// 	return
	// }

	type Headers struct {
		EmailToken string `header:"X-Email-Access-Token"`
	}

	var headers Headers
	if err := ctx.ShouldBindHeader(&headers); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("headers-parsing-error", "Headers Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	token, err := ch.service.Register(ctxWithTimeout, user, image, ctx.SaveUploadedFile, headers.EmailToken)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Token string `json:"token"`
	}{
		ginutil.Response{
			Message: "The user has successfully been saved.",
			Links: hypermedia.Links{
				deleteUserLink,
			},
		},
		token,
	})
}

func (ch *customerHandler) changePassword(ctx *gin.Context) {

	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest(
			"request-parsing",
			"Request Parsing Error",
			err.Error(),
		))
	}

	type Headers struct {
		EmailToken string `header:"X-Email-Access-Token"`
	}

	var headers Headers
	if err := ctx.ShouldBindHeader(&headers); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("headers-parsing-error", "Headers Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	err = ch.service.ChangePassword(ctxWithTimeout, request.Password, request.Email, headers.EmailToken)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, ginutil.Response{
		"The password has successfully been changed.",
		hypermedia.Links{
			loginLink,
		},
	})
}

func (ch *customerHandler) delete(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	err := ch.service.Delete(ctxWithTimeout, ctx.MustGet("userID").(uuid.UUID))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, ginutil.Response{
		Message: "The user has successfully been deleted.",
		Links: hypermedia.Links{
			registerUserLink,
			verifyEmailLink,
			verifyNumberLink,
		},
	})
}

func (uh *customerHandler) get(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	user, err := uh.service.GetByID(ctxWithTimeout, ctx.MustGet("userID").(uuid.UUID))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		entity.User `json:"user"`
	}{
		ginutil.Response{
			Message: "The user has successfully been found.",
			Links: hypermedia.Links{
				deleteUserLink,
			},
		},
		user,
	})
}

func (uh *customerHandler) updatePersonalInfo(ctx *gin.Context) {
	var user entity.UserPersonalInfo

	err := ctx.ShouldBindJSON(&user)
	if err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.JSON(err.Error()))
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	err = uh.service.UpdatePersonalInfo(ctxWithTimeout, user, ctx.MustGet("userID").(uuid.UUID))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK,
		ginutil.Response{
			Message: "The user has successfully been updated.",
			Links: hypermedia.Links{
				deleteUserLink,
			},
		},
	)
}

func (uh *customerHandler) updateContactInfo(ctx *gin.Context) {
	var user entity.UserContactInfo

	err := ctx.ShouldBindJSON(&user)
	if err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.JSON(err.Error()))
	}

	type Headers struct {
		CustomerUpdateToken string `header:"X-Customer-Update-Token"`
	}

	var headers Headers
	if err := ctx.ShouldBindHeader(&headers); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("headers-parsing-error", "Headers Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	err = uh.service.UpdateContactInfo(ctxWithTimeout, ctx.MustGet("userID").(uuid.UUID), user, headers.CustomerUpdateToken)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK,
		ginutil.Response{
			Message: "The user has successfully been updated.",
			Links: hypermedia.Links{
				deleteUserLink,
			},
		},
	)
}

func newcustomerHandler(service service.CustomerService) customerHandler {
	return customerHandler{userHandler: newUserHandler(service), service: service}
}
