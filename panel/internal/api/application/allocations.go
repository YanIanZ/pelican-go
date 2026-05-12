package application

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type AllocationHandler struct {
	DB *gorm.DB
}

func (h *AllocationHandler) Index(c *gin.Context) {
	nodeID := c.Param("nodeID")
	var allocs []models.Allocation
	h.DB.Where("node_id = ?", nodeID).Find(&allocs)
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": allocs})
}

func (h *AllocationHandler) Store(c *gin.Context) {
	nodeID := c.Param("nodeID")
	var req struct {
		IP    string `json:"ip" binding:"required"`
		Alias string `json:"alias"`
		Ports []int  `json:"ports" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	for _, port := range req.Ports {
		alloc := models.Allocation{NodeID: parseUint(nodeID), IP: req.IP, Port: port}
		if req.Alias != "" {
			alloc.Alias = &req.Alias
		}
		h.DB.Create(&alloc)
	}
	c.JSON(http.StatusCreated, gin.H{})
}

func (h *AllocationHandler) Delete(c *gin.Context) {
	h.DB.Delete(&models.Allocation{}, c.Param("allocationID"))
	c.JSON(http.StatusNoContent, nil)
}

func parseUint(s string) uint {
	var id uint
	fmt.Sscanf(s, "%d", &id)
	return id
}
