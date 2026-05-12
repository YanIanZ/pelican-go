package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type NodeHandler struct {
	DB *gorm.DB
}

func (h *NodeHandler) Index(c *gin.Context) {
	var nodes []models.Node
	h.DB.Find(&nodes)
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": nodes})
}

func (h *NodeHandler) View(c *gin.Context) {
	var node models.Node
	if err := h.DB.First(&node, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errors(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"object": "node", "attributes": node})
}

func (h *NodeHandler) Store(c *gin.Context) {
	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	node.UUID = genUUID()
	node.DaemonToken = randomStr(64)
	node.DaemonTokenID = randomStr(16)
	if err := h.DB.Create(&node).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errors(err))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"object": "node", "attributes": node})
}

func (h *NodeHandler) Update(c *gin.Context) {
	var node models.Node
	if err := h.DB.First(&node, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errors(err))
		return
	}
	var updates models.Node
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	h.DB.Model(&node).Omit("UUID", "DaemonToken", "DaemonTokenID").Updates(updates)
	c.JSON(http.StatusOK, gin.H{"object": "node", "attributes": node})
}

func (h *NodeHandler) Delete(c *gin.Context) {
	h.DB.Delete(&models.Node{}, c.Param("id"))
	c.JSON(http.StatusNoContent, nil)
}

func (h *NodeHandler) Configuration(c *gin.Context) {
	var node models.Node
	if err := h.DB.First(&node, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errors(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"debug":    false,
		"uuid":     node.UUID,
		"token_id": node.DaemonTokenID,
		"token":    node.DaemonToken,
		"api": gin.H{
			"host": "0.0.0.0",
			"port": node.DaemonListen,
			"ssl":  gin.H{"enabled": node.Scheme == "https", "cert": "", "key": ""},
		},
		"system": gin.H{
			"data": node.DaemonBase,
			"sftp": gin.H{"bind_port": node.DaemonSFTP},
		},
	})
}
