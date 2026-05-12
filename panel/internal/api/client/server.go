package client

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type ServerHandler struct {
	DB *gorm.DB
}

func (h *ServerHandler) Index(c *gin.Context) {
	var server models.Server
	if err := h.DB.Where("uuid = ?", c.Param("uuid")).Preload("Node").Preload("Allocation").First(&server).Error; err != nil {
		c.JSON(http.StatusNotFound, errorsMsg("Server not found"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"object": "server", "attributes": gin.H{
		"uuid":       server.UUID,
		"uuid_short": server.UUIDShort,
		"name":       server.Name,
		"description": server.Description,
		"node":       server.Node.Name,
		"memory":     server.Memory,
		"swap":       server.Swap,
		"disk":       server.Disk,
		"io":         server.IO,
		"cpu":        server.CPU,
		"oom_killer": server.OOMKiller,
		"status":     server.Status,
		"allocation": server.Allocation,
	}})
}

func (h *ServerHandler) Power(c *gin.Context) {
	var req struct {
		Signal string `json:"signal" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ServerHandler) Command(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (h *ServerHandler) Rename(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	h.DB.Model(&models.Server{}).Where("uuid = ?", c.Param("uuid")).Update("name", req.Name)
	c.JSON(http.StatusOK, gin.H{})
}
