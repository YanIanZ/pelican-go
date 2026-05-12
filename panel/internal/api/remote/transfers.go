package remote

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TransferHandler struct {
	DB *gorm.DB
}

func (h *TransferHandler) Failure(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

func (h *TransferHandler) Success(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

func (h *TransferHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
