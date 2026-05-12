package remote

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SFTPHandler struct {
	DB *gorm.DB
}

func (h *SFTPHandler) Auth(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": []gin.H{{"code": "ValidationError", "status": "400", "detail": err.Error()}}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":        "",
		"server":      "",
		"permissions": []string{},
	})
}
