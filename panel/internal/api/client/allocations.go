package client

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AllocationHandler struct {
	DB *gorm.DB
}

func (h *AllocationHandler) Index(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func (h *AllocationHandler) Store(c *gin.Context)  { c.JSON(http.StatusCreated, gin.H{"data": gin.H{}}) }
