package user

import (
	ginutil "maryan_api/pkg/ginutils"
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
	customerRouter := ginutil.CreateAuthRouter("/customer", customer.service.secretKey(), db, s)
	customerRouterNoAuth := ginutil.CreateRouter("/customer", db, s) //OK
	customerRouterNoAuth.POST("/guest", customer.guest)
	customerRouter.POST("/verify-email", customer.verifyEmail)   //OK
	customerRouter.POST("/verify-number", customer.verifyNumber) //OK
	customerRouter.POST("/google-oauth", customer.googleOAUTH)   //OK
	customerRouter.GET("/user", customer.getUser)                //OK
	customerRouter.POST("/user", customer.saveUser)              //OK
	customerRouter.DELETE("/user", customer.deleteUser)          //OK

	customerRouter.POST("/login", customer.login)        //OK
	customerRouter.POST("/login-jwt", customer.loginJWT) //OK
	//CUSTOMER ROUTES

	//ADMIN ROUTES
	admin := newAdminHandler(newAdminServiceImpl(newAdminRepoMySQL(db)))
	adminRouter := ginutil.CreateAuthRouter("/admin", admin.service.secretKey(), db, s) //OK
	adminRouter.POST("/login", admin.login)                                             //OK
	adminRouter.POST("/login-jwt", admin.loginJWT)                                      //OK
	adminRouter.GET("/user", admin.getUser)                                             //OK
	adminRouter.GET("users", admin.getUsers)                                            //Unfinished
	//ADMIN ROUTES

	//DRIVER ROUTES
	driver := newDriverHandler(newDriverServiceImpl(newDriverRepoMySQL(db)))
	driverRouter := ginutil.CreateAuthRouter("/driver", driver.service.secretKey(), db, s) //OK
	driverRouter.POST("/login", driver.login)                                              //OK
	driverRouter.POST("/loginJWT", driver.loginJWT)                                        //OK
	driverRouter.GET("/user", driver.getUser)                                              //OK
	//DRIVER ROUTES

	//SUPPORSTEMPLOYEE ROUTES
	supportEmployee := newSupportEmployeeHandler(newSupportEmployeeServiceImpl(newSupportEmployeeRepoMySQL(db)))
	supportEmployeeRouter := ginutil.CreateAuthRouter("/support-employee", admin.service.secretKey(), db, s) //OK
	supportEmployeeRouter.POST("/login", supportEmployee.login)                                              //OK
	supportEmployeeRouter.POST("/login-jwt", supportEmployee.loginJWT)                                       //OK
	supportEmployeeRouter.GET("/user", supportEmployee.getUser)                                              //OK
	//SUPPORSTEMPLOYEE ROUTES

}

type userHandler struct {
	service userService
}

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

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (uh userHandler) loginJWT(c *gin.Context) {
	id := c.MustGet("userID").(uuid.UUID)
	email := c.MustGet("email").(string)

	token, err := uh.service.loginJWT(id, email)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (uh userHandler) getUser(c *gin.Context) {
	user, err := uh.service.usersData(c.MustGet("userID").(uuid.UUID))
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user was successfuly found", "user": user})
}

type CustomerHandler struct {
	userHandler
	service customerService
}

func (ch CustomerHandler) guest(c *gin.Context) {
	token, err := ch.service.guest()
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (ch CustomerHandler) verifyEmail(c *gin.Context) {
	var email struct {
		Val string `json:"email"`
	}

	if err := c.ShouldBindJSON(&email); err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("email-parsing", "Body Parsing Error", err.Error()))
		return
	}

	verificationCode, exists, err := ch.service.verifyEmail(email.Val)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	if exists {
		c.JSON(http.StatusOK, gin.H{"exists": true})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": false, "verificationCode": verificationCode})

}

func (ch CustomerHandler) verifyNumber(c *gin.Context) {
	var number struct {
		Val string `json:"number"`
	}

	if err := c.ShouldBindJSON(&number); err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("number-parsing", "Body Parsing Error", err.Error()))
		return
	}

	verificationCode, err := ch.service.verifyNumber(number.Val)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"verificationCode": verificationCode})
}

func (ch CustomerHandler) googleOAUTH(c *gin.Context) {
	var request struct {
		Code string `json:"code"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("google-code-parsing", "Body Parsing Error", err.Error()))
		return
	}

	token, status, err := ch.service.googleOAUTH(request.Code, c.Request.Context(), c.MustGet("userID").(uuid.UUID))

	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "existed": status})
}

func (ch CustomerHandler) saveUser(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest("user-parsing", "Body Parsing Error", err.Error()))
		return
	}

	user.ID = c.MustGet("userID").(uuid.UUID)
	err := ch.service.save(&user)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user was successfuly saved"})
}

func (ch CustomerHandler) deleteUser(c *gin.Context) {

	err := ch.service.delete(c.MustGet("userID").(uuid.UUID))
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user was successfuly deleted"})

}

type AdminHandler struct {
	userHandler
	service adminService
}

func (ah AdminHandler) getUsers(c *gin.Context) {
	params, add, isNil := rfc7807.StartSettingInvalidParams()

	pageSize, err := ginutil.ParseIntKey(c, "pageSize")
	if err != nil {
		add("pageSize", err.Error())
	}

	pageNumber, err := ginutil.ParseIntKey(c, "pageNumber")
	if err != nil {
		add("pageSize", err.Error())
	}

	if !isNil() {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest(
			"users-order-data-parsing",
			"Users Order Data Error",
			"Could not parse user order data.").SetInvalidParams(*params),
		)
		return
	}

	users, pages, err := ah.service.getUsers(pageSize, pageNumber)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "users were successfuly found", "users": users, "pages": pages})
}

type DriverHandler struct {
	userHandler
	service driverService
}

type SupportEmployeeHandler struct {
	userHandler
	service supportEmployeeService
}

//Declaration functions

func newCustomerHandler(service customerService) CustomerHandler {
	return CustomerHandler{userHandler: userHandler{service.userService()}, service: service}
}

func newAdminHandler(service adminService) AdminHandler {
	return AdminHandler{userHandler: userHandler{service.userService()}, service: service}
}

func newDriverHandler(service driverService) DriverHandler {
	return DriverHandler{userHandler: userHandler{service.userService()}, service: service}
}

func newSupportEmployeeHandler(service supportEmployeeService) SupportEmployeeHandler {
	return SupportEmployeeHandler{userHandler: userHandler{service.userService()}, service: service}
}
