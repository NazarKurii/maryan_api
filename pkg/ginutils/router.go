package ginutil

import (
	"maryan_api/pkg/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateAuthRouter(groupName string, secretKey []byte, db *gorm.DB, s *gin.Engine) *gin.RouterGroup {
	group := s.Group(groupName)
	group.Use(auth.AuthorizeLog(secretKey, db))
	return group
}

func CreateRouter(groupName string, db *gorm.DB, s *gin.Engine) *gin.RouterGroup {
	group := s.Group(groupName)
	group.Use(auth.Log(db))
	return group
}
