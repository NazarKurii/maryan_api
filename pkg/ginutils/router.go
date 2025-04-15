package ginutil

import (
	"maryan_api/pkg/auth"

	"github.com/gin-gonic/gin"
)

func CreateAuthRouter(groupName string, secretKey []byte, s *gin.Engine) *gin.RouterGroup {
	group := s.Group(groupName)
	group.Use(auth.Authorize(secretKey))
	return group
}
