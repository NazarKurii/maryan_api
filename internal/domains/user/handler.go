package user

import (
	"maryan_api/config"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Routes Register
func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {

	//CUSTOMER ROUTES
	customer := newCustomerHandler(newCustomerServiceImpl(newCustomerRepoMySQL(db), client))
	authCustomerRouter := ginutil.CreateAuthRouter("/customer", customer.service.secretKey(), s)
	customerRouter := s.Group("/customer")

	customerRouter.POST("/verify-email", customer.verifyEmail)
	customerRouter.POST("/verify-email-code/:token", customer.verifyEmailCode)
	customerRouter.POST("/verify-number", customer.verifyNumber)
	customerRouter.POST("/verify-number-code/:token", customer.verifyNumberCode)
	customerRouter.POST("/register", customer.register)

	customerRouter.POST("/login", customer.login)
	customerRouter.POST("/google-oauth", customer.googleOAUTH)

	authCustomerRouter.POST("/login-jwt", customer.loginJWT)
	authCustomerRouter.GET("", customer.get)

	authCustomerRouter.DELETE("", customer.delete)

	//ADMIN ROUTES

}

type userHandler struct {
	service userService
}

func mai() {}

func (uh userHandler) login(c *gin.Context) {
	var credentials struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		ginutil.HandlerProblemAbort(
			c,
			rfc7807.BadRequest(
				"login-creadentials-parsing",
				"Credentials Parsing Error",
				err.Error(),
			),
		)
		return
	}

	token, err := uh.service.login(credentials.Email, credentials.Password)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, struct {
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

func (uh userHandler) loginJWT(c *gin.Context) {
	id := c.MustGet("userID").(uuid.UUID)
	email := c.MustGet("email").(string)

	token, err := uh.service.loginJWT(id, email)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, struct {
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

type CustomerHandler struct {
	userHandler
	service customerService
}

func (ch CustomerHandler) verifyEmail(c *gin.Context) {
	var email struct {
		Val string `json:"email"`
	}

	if err := c.ShouldBindJSON(&email); err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("email-parsing", "Body Parsing Error", err.Error()))
		return
	}

	token, exists, err := ch.service.verifyEmail(email.Val)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
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

	c.JSON(http.StatusOK, resp)
}

func (ch CustomerHandler) verifyEmailCode(c *gin.Context) {
	var code struct {
		Val string `json:"code"`
	}

	if err := c.ShouldBindJSON(&code); err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("code-parsing", "Body Parsing Error", err.Error()))
		return
	}

	token := c.Param("token")

	token, err := ch.service.verifyEmailCode(code.Val, token)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, struct {
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

func (ch CustomerHandler) verifyNumber(c *gin.Context) {
	var number struct {
		Val string `json:"number"`
	}

	if err := c.ShouldBindJSON(&number); err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("number-parsing", "Body Parsing Error", err.Error()))
		return
	}

	token, err := ch.service.verifyNumber(number.Val)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, ginutil.Response{
		"The code has successfuly been sent.",
		hypermedia.Links{
			verifyEmailLink,
			registerUserLink,
			hypermedia.Link{"verifyNumberCode": hypermedia.Href{config.APIURL() + "/customer/verify-number-code/" + token, http.MethodPost}},
		},
	},
	)
}

func (ch CustomerHandler) verifyNumberCode(c *gin.Context) {
	var code struct {
		Val string `json:"code"`
	}

	if err := c.ShouldBindJSON(&code); err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("code-parsing", "Body Parsing Error", err.Error()))
		return
	}

	token := c.Param("token")

	token, err := ch.service.verifyNumberCode(code.Val, token)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, struct {
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

func (ch CustomerHandler) googleOAUTH(c *gin.Context) {
	var request struct {
		Code string `json:"code"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("google-code-parsing", "Body Parsing Error", err.Error()))
		return
	}

	token, isNew, err := ch.service.googleOAUTH(request.Code, c.Request.Context(), c.MustGet("userID").(uuid.UUID))

	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, struct {
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

func (ch CustomerHandler) register(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("user-parsing", "Body Parsing Error", err.Error()))
		return
	}

	image, err := c.FormFile("image")
	if err != nil {
		if err.Error() != "no multipart boundary param in Content-Type" {
			ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("image-forming-error", "Image Froming Error", err.Error()))
		}
	}

	type Headers struct {
		EmailToken  string `header:"X-Email-Access-Token" binding:"required"`
		NumberToken string `header:"X-Number-Access-Token" binding:"required"`
	}

	var headers Headers
	if err := c.ShouldBindHeader(&headers); err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("headers-parsing-error", "Headers Error", err.Error()))
		return
	}

	token, err := ch.service.register(&user, image, headers.EmailToken, headers.NumberToken)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, struct {
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

func (ch CustomerHandler) delete(c *gin.Context) {

	err := ch.service.delete(c.MustGet("userID").(uuid.UUID))
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, ginutil.Response{
		"The user has successfuly been deleted.",
		hypermedia.Links{
			registerUserLink,
			verifyEmailLink,
			verifyNumberLink,
		},
	})

}
func (uh CustomerHandler) get(c *gin.Context) {
	user, err := uh.service.get(c.MustGet("userID").(uuid.UUID))
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, struct {
		ginutil.Response
		ShortUser `json:"user"`
	}{
		ginutil.Response{
			"The user has successfuly been found.",
			[]hypermedia.Link{deleteUserLink},
		},
		user,
	})
}

//Declaration functions

func newCustomerHandler(service customerService) CustomerHandler {
	return CustomerHandler{userHandler: userHandler{service.userService()}, service: service}
}
