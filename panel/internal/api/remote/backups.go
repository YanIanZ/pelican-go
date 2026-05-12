package remote

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type BackupHandler struct {
	DB *gorm.DB
}

func (h *BackupHandler) Upload(c *gin.Context) {
	uuid := c.Param("uuid")

	var backup models.Backup
	if err := h.DB.Where("uuid = ?", uuid).First(&backup).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"errors": []gin.H{{"code": "HttpNotFoundException", "status": "404", "detail": "Backup not found"}}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"parts": []string{}, "part_size": 0})
}

func (h *BackupHandler) Status(c *gin.Context) {
	uuid := c.Param("uuid")

	var req struct {
		Successful bool   `json:"successful"`
		Size       int64  `json:"size"`
		Checksum   string `json:"checksum"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": []gin.H{{"code": "ValidationError", "status": "400", "detail": err.Error()}}})
		return
	}

	var backup models.Backup
	if err := h.DB.Where("uuid = ?", uuid).First(&backup).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"errors": []gin.H{{"code": "HttpNotFoundException", "status": "404", "detail": "Backup not found"}}})
		return
	}

	backup.IsSuccessful = &req.Successful
	backup.Size = req.Size
	backup.Checksum = &req.Checksum
	now := time.Now()
	backup.CompletedAt = &now
	h.DB.Save(&backup)
	c.JSON(http.StatusNoContent, nil)
}

func (h *BackupHandler) Restore(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}
