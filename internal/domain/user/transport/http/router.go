package transport

import (
	"maryan_api/internal/domain/user/repo"
	"maryan_api/internal/domain/user/service"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Admin struct {
	adminHandler adminHandler
}

type Customer struct {
	customerHandler customerHandler
}

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {

	//CUSTOMER ROUTES
	customer := Customer{newcustomerHandler(service.NewCustomerServiceImpl(repo.NewCustomerRepo(db), client))}
	authCustomerRouter := ginutil.CreateAuthRouter("/customer", customer.customerHandler.service.SecretKey(), s)
	customerRouter := s.Group("/customer")

	customerRouter.POST("/verify-email", customer.customerHandler.verifyEmail)
	customerRouter.POST("/verify-email-code/:token", customer.customerHandler.verifyEmailCode)
	customerRouter.POST("/verify-number", customer.customerHandler.verifyNumber)
	customerRouter.POST("/verify-number-code/:token", customer.customerHandler.verifyNumberCode)
	customerRouter.POST("/register", customer.customerHandler.register)

	customerRouter.POST("/login", customer.customerHandler.login)
	customerRouter.POST("/google-oauth", customer.customerHandler.googleOAUTH)

	authCustomerRouter.POST("/login-jwt", customer.customerHandler.loginJWT)
	authCustomerRouter.GET("", customer.customerHandler.get)

	authCustomerRouter.DELETE("", customer.customerHandler.delete)

	//ADMIN ROUTES
	admin := Admin{newAdminHandler(service.NewAdminServiceImpl(repo.NewAdminRepo(db), client))}
	authAdminRouter := ginutil.CreateAuthRouter("/admin", admin.adminHandler.service.SecretKey(), s)
	adminRouter := s.Group("/admin")

	adminRouter.POST("/login", admin.adminHandler.login)
	authAdminRouter.GET("/users/:page/:size/:role/:order-by/:order-way", admin.adminHandler.users)
	adminRouter.POST("/hash-password", admin.adminHandler.hashPassword)

}

var (
	guestLink = hypermedia.Link{
		"create": {Href: "/customer/guest", Method: "POST"},
	}

	verifyEmailLink = hypermedia.Link{
		"verifyEmail": {Href: "/customer/verify-email", Method: "POST"},
	}

	verifyNumberLink = hypermedia.Link{
		"verifyPhoneNumber": {Href: "/customer/verify-number", Method: "POST"},
	}

	verifyNumberCodeLink = hypermedia.Link{
		"codeVerification": {Href: "/customer/code-verification", Method: "POST"},
	}

	googleOAuthLink = hypermedia.Link{
		"loginOAuth": {Href: "/customer/google-oauth", Method: "POST"},
	}

	getUserLink = hypermedia.Link{
		"self": {Href: "/customer/user", Method: "GET"},
	}

	registerUserLink = hypermedia.Link{
		"register": {Href: "/customer/user", Method: "POST"},
	}

	deleteUserLink = hypermedia.Link{
		"delete": {Href: "/customer/user", Method: "DELETE"},
	}

	loginLink = hypermedia.Link{
		"login": {Href: "/customer/login", Method: "POST"},
	}

	loginJWTLink = hypermedia.Link{
		"loginJWT": {Href: "/customer/login-jwt", Method: "POST"},
	}
)
