package documents

import (
	"net/http"

	"github.com/nazarkurii/marshrutka_api/config"
	ginutil "github.com/nazarkurii/marshrutka_api/pkg/ginutils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {
	//-----------------------Trip Routes---------------------------------------
	var customer = s.Group("/customer")
	customer.GET("/legal-documents-uk.pdf", ginutil.ServeFileAttachment(config.RootPath()+"/static/pdf/legal-policy-uk.pdf", "legal-policy-uk.pdf"))
	customer.GET("/legal-documents-en.pdf", ginutil.ServeFileAttachment(config.RootPath()+"/static/pdf/legal-policy-en.pdf", "legal-policy-en.pdf"))
}
