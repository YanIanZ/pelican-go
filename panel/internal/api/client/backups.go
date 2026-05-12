package client

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type BackupHandler struct {
	DB *gorm.DB
}

func (h *BackupHandler) Index(c *gin.Context) {
	var server models.Server
	h.DB.Where("uuid = ?", c.Param("uuid")).First(&server)
	var backups []models.Backup
	h.DB.Where("server_id = ?", server.ID).Find(&backups)
	c.JSON(http.StatusOK, gin.H{"data": backups})
}

func (h *BackupHandler) Store(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{})
}

func (h *BackupHandler) Delete(c *gin.Context) {
	h.DB.Where("uuid = ?", c.Param("backup")).Delete(&models.Backup{})
	c.Status(http.StatusNoContent)
}

func (h *BackupHandler) Restore(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
