package remote

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type InstallHandler struct {
	DB *gorm.DB
}

func (h *InstallHandler) Index(c *gin.Context) {
	uuid := c.Param("uuid")

	var server models.Server
	if err := h.DB.Where("uuid = ?", uuid).Preload("Egg").First(&server).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"errors": []gin.H{{"code": "HttpNotFoundException", "status": "404", "detail": "Server not found"}}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"container_image": server.Egg.ScriptContainer,
		"entrypoint":      server.Egg.ScriptEntry,
		"script":          server.Egg.ScriptInstall,
	})
}

func (h *InstallHandler) Store(c *gin.Context) {
	uuid := c.Param("uuid")

	var req struct {
		Successful bool `json:"successful"`
		Reinstall  bool `json:"reinstall"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": []gin.H{{"code": "ValidationError", "status": "400", "detail": err.Error()}}})
		return
	}

	var server models.Server
	if err := h.DB.Where("uuid = ?", uuid).First(&server).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"errors": []gin.H{{"code": "HttpNotFoundException", "status": "404", "detail": "Server not found"}}})
		return
	}

	if req.Successful {
		now := time.Now()
		server.Status = nil
		server.InstalledAt = &now
	} else {
		status := "install_failed"
		if req.Reinstall {
			status = "reinstall_failed"
		}
		server.Status = &status
	}

	h.DB.Save(&server)
	c.JSON(http.StatusNoContent, nil)
}
