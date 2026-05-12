package remote

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ActivityHandler struct {
	DB *gorm.DB
}

func (h *ActivityHandler) Store(c *gin.Context) {
	var req struct {
		Data []struct {
			User      string                 `json:"user"`
			Server    string                 `json:"server"`
			Event     string                 `json:"event"`
			Metadata  map[string]interface{} `json:"metadata"`
			IP        string                 `json:"ip"`
			Timestamp string                 `json:"timestamp"`
		} `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": []gin.H{{"code": "ValidationError", "status": "400", "detail": err.Error()}}})
		return
	}
	c.Status(http.StatusOK)
}
