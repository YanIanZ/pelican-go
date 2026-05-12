package client

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SubuserHandler struct {
	DB *gorm.DB
}

func (h *SubuserHandler) Index(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func (h *SubuserHandler) Store(c *gin.Context)   { c.JSON(http.StatusCreated, gin.H{"data": gin.H{}}) }
func (h *SubuserHandler) Update(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"data": gin.H{}}) }
func (h *SubuserHandler) Delete(c *gin.Context)  { c.Status(http.StatusNoContent) }
