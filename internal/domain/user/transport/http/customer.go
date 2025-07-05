package http

import (
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

func (ch *customerHandler) verifyEmail(ctx *gin.Context) {
	var email struct {
		Val string `json:"email"`
	}

	if err := ctx.ShouldBindJSON(&email); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("email-parsing", "Body Parsing Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	token, exists, err := ch.service.VerifyEmail(ctxWithTimeout, email.Val)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	var resp = struct {
		ginutil.Response
		Exists bool
	}{
		Response: ginutil.Response{
			Links: hypermedia.Links{
				verifyNumberLink,
				registerUserLink,
			},
		},
	}

	if !exists {
		resp.Message = "The code has successfuly been sent."
		resp.Links.Add("verifyEmailCode", config.APIURL()+"/customer/verify-email-code/"+token, http.MethodPost)
	} else {
		resp.Message = "Email already exists."
		resp.Exists = true
		resp.Links.AddLink(verifyEmailLink)
	}

	ctx.JSON(http.StatusOK, resp)
}

func (ch *customerHandler) verifyEmailCode(ctx *gin.Context) {
	var code struct {
		Val string `json:"code"`
	}

	if err := ctx.ShouldBindJSON(&code); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("code-parsing", "Body Parsing Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	token, err := ch.service.VerifyEmailCode(ctxWithTimeout, code.Val, ctx.Param("token"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Token string `json:"token"`
	}{
		ginutil.Response{
			"The code has successfuly been verified",
			[]hypermedia.Link{
				verifyNumberLink,
				registerUserLink,
			},
		},
		token,
	})

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
		"The code has successfuly been sent.",
		hypermedia.Links{
			verifyEmailLink,
			registerUserLink,
			hypermedia.Link{"verifyNumberCode": hypermedia.Href{config.APIURL() + "/customer/verify-number-code/" + token, http.MethodPost}},
		},
	},
	)
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
			"The code has successfuly been verified",
			[]hypermedia.Link{
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
			"User has been successfuly logged in.",
			[]hypermedia.Link{deleteUserLink},
		},
		token,
		isNew,
	})
}

func (ch *customerHandler) register(ctx *gin.Context) {
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

	token, err := ch.service.Register(ctxWithTimeout, user, image, headers.EmailToken, headers.NumberToken)
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

func (ch *customerHandler) delete(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	err := ch.service.Delete(ctxWithTimeout, ctx.MustGet("userID").(uuid.UUID))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, ginutil.Response{
		"The user has successfuly been deleted.",
		hypermedia.Links{
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
		entity.UserSimplified `json:"user"`
	}{
		ginutil.Response{
			"The user has successfuly been found.",
			[]hypermedia.Link{deleteUserLink},
		},
		user,
	})
}

//Admin Hadler

//Declaration functions

func newcustomerHandler(service service.CustomerService) customerHandler {
	return customerHandler{userHandler: newUserHandler(service), service: service}
}
