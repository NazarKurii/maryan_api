package http

import (
	"maryan_api/internal/domain/user/repo"
	"maryan_api/internal/domain/user/service"
	"maryan_api/pkg/auth"
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

type Driver struct {
	userhandler userHandler
}

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {

	//CUSTOMER ROUTES
	customer := Customer{newcustomerHandler(service.NewCustomerServiceImpl(repo.NewCustomerRepo(db), client))}
	authCustomerRouter := ginutil.CreateAuthRouter("/customer", customer.customerHandler.service.SecretKey(), s)
	customerRouter := s.Group("/customer")

	customerRouter.POST("/verify-email", customer.customerHandler.verifyEmailIfExists)
	customerRouter.POST("/verify-email-code/:token", customer.customerHandler.verifyEmailCode)
	customerRouter.POST("/verify-number", customer.customerHandler.verifyNumber)
	customerRouter.POST("/verify-number-code/:token", customer.customerHandler.verifyNumberCode)
	customerRouter.POST("/register", customer.customerHandler.register)
	customerRouter.POST("/change-password/verify-email", customer.customerHandler.verifyEmailChangePassword)
	customerRouter.POST("/change-password/verify-email-code/:token", customer.customerHandler.verifyEmailCodePasswordChanging)
	customerRouter.POST("/change-password", customer.customerHandler.changePassword)

	customerRouter.POST("/login", customer.customerHandler.login)
	customerRouter.POST("/google-oauth", customer.customerHandler.googleOAUTH)

	authCustomerRouter.POST("/login-jwt", customer.customerHandler.loginJWT)
	authCustomerRouter.GET("", customer.customerHandler.get)
	authCustomerRouter.PUT("/personal-info", customer.customerHandler.updatePersonalInfo)
	authCustomerRouter.PUT("/contact-info", customer.customerHandler.updateContactInfo)

	authCustomerRouter.POST("/verify-update", customer.customerHandler.verifyEmailCustomerUpdate)
	authCustomerRouter.POST("/verify-update-code/:token", customer.customerHandler.VerifyCustomerUpdateCode)

	authCustomerRouter.DELETE("", customer.customerHandler.delete)

	//ADMIN ROUTES
	admin := Admin{newAdminHandler(service.NewAdminServiceImpl(repo.NewAdminRepo(db), client))}
	authAdminRouter := ginutil.CreateAuthRouter("/admin", admin.adminHandler.service.SecretKey(), s)
	adminRouter := s.Group("/admin")

	adminRouter.POST("/login", admin.adminHandler.login)
	adminRouter.POST("/hash-password", admin.adminHandler.hashPassword)
	authAdminRouter.POST("/login-jwt", admin.adminHandler.loginJWT)
	authAdminRouter.GET("/users", admin.adminHandler.getUsers)
	authAdminRouter.GET("/user", admin.adminHandler.getUser)
	authAdminRouter.GET("", admin.adminHandler.get)
	authAdminRouter.POST("/driver", admin.adminHandler.newEmployee(auth.Driver))
	authAdminRouter.POST("/support", admin.adminHandler.newEmployee(auth.Support))
	authAdminRouter.POST("/admin", admin.adminHandler.newEmployee(auth.Admin))
	authAdminRouter.POST("/employee-schedule", admin.adminHandler.setEmployeeAvailability)
	authAdminRouter.GET("/available-employees", admin.adminHandler.getAvailableEmployees)
	authAdminRouter.GET("/free-drivers", admin.adminHandler.getFreeDrivers)

	//ADMIN ROUTES
	driver := Driver{newUserHandler(service.NewUserService(auth.Driver, repo.NewUserRepo(db)))}
	// authDriverRouter := ginutil.CreateAuthRouter("/driver", driver.userhandler.service.SecretKey(), s)
	driverRouter := s.Group("/driver")

	driverRouter.POST("/login", driver.userhandler.login)
}

var (
	guestLink = hypermedia.Link{
		Name: "guest",
		Data: hypermedia.LinkData{Href: "/customer/guest", Method: "POST"},
	}

	verifyEmailLink = hypermedia.Link{
		Name: "verifyEmail",
		Data: hypermedia.LinkData{Href: "/customer/verify-email", Method: "POST"},
	}

	verifyNumberLink = hypermedia.Link{
		Name: "verifyPhoneNumber",
		Data: hypermedia.LinkData{Href: "/customer/verify-number", Method: "POST"},
	}

	verifyNumberCodeLink = hypermedia.Link{
		Name: "codeVerification",
		Data: hypermedia.LinkData{Href: "/customer/code-verification", Method: "POST"},
	}

	googleOAuthLink = hypermedia.Link{
		Name: "loginOAuth",
		Data: hypermedia.LinkData{Href: "/customer/google-oauth", Method: "POST"},
	}

	getUserLink = hypermedia.Link{
		Name: "self",
		Data: hypermedia.LinkData{Href: "/customer/user", Method: "GET"},
	}

	registerUserLink = hypermedia.Link{
		Name: "register",
		Data: hypermedia.LinkData{Href: "/customer/user", Method: "POST"},
	}

	deleteUserLink = hypermedia.Link{
		Name: "delete",
		Data: hypermedia.LinkData{Href: "/customer/user", Method: "DELETE"},
	}

	loginLink = hypermedia.Link{
		Name: "login",
		Data: hypermedia.LinkData{Href: "/customer/login", Method: "POST"},
	}

	loginJWTLink = hypermedia.Link{
		Name: "loginJWT",
		Data: hypermedia.LinkData{Href: "/customer/login-jwt", Method: "POST"},
	}

	getUsersLink = hypermedia.Link{
		Name: "createDriver",
		Data: hypermedia.LinkData{Href: "/admin/users", Method: "GET"},
	}
)
