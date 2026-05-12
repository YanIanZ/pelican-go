package client

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DatabaseHandler struct {
	DB *gorm.DB
}

func (h *DatabaseHandler) Index(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func (h *DatabaseHandler) Store(c *gin.Context)  { c.JSON(http.StatusCreated, gin.H{"data": gin.H{}}) }
func (h *DatabaseHandler) Delete(c *gin.Context) { c.Status(http.StatusNoContent) }
