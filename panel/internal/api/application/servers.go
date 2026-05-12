package application

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type ServerHandler struct {
	DB *gorm.DB
}

func (h *ServerHandler) Index(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}

	var total int64
	h.DB.Model(&models.Server{}).Count(&total)

	var servers []models.Server
	h.DB.Offset((page - 1) * perPage).Limit(perPage).
		Preload("Node").Preload("Owner").Preload("Allocation").Preload("Egg").
		Find(&servers)

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   servers,
		"meta": gin.H{
			"pagination": gin.H{
				"total":        total,
				"count":        len(servers),
				"per_page":     perPage,
				"current_page": page,
				"total_pages":  (total + int64(perPage) - 1) / int64(perPage),
			},
		},
	})
}

func (h *ServerHandler) View(c *gin.Context) {
	var server models.Server
	if err := h.DB.Preload("Node").Preload("Owner").Preload("Allocation").Preload("Egg").First(&server, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errors(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"object": "server", "attributes": server})
}

type CreateServerRequest struct {
	Name            string  `json:"name" binding:"required"`
	UserID          uint    `json:"user_id" binding:"required"`
	NodeID          uint    `json:"node_id" binding:"required"`
	EggID           uint    `json:"egg_id" binding:"required"`
	Memory          int64   `json:"memory" binding:"required"`
	Swap            int64   `json:"swap"`
	Disk            int64   `json:"disk" binding:"required"`
	IO              int     `json:"io"`
	CPU             int64   `json:"cpu" binding:"required"`
	Image           string  `json:"image" binding:"required"`
	Startup         string  `json:"startup" binding:"required"`
	Threads         *string `json:"threads"`
	OOMKiller       bool    `json:"oom_killer"`
	DatabaseLimit   *int    `json:"database_limit"`
	AllocationLimit *int    `json:"allocation_limit"`
	BackupLimit     int     `json:"backup_limit"`
	AllocationID    *uint   `json:"allocation_id"`
	SkipScripts     bool    `json:"skip_scripts"`
}

func (h *ServerHandler) Store(c *gin.Context) {
	var req CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	server := models.Server{
		UUID:            genUUID(),
		UUIDShort:       genUUIDShort(),
		OwnerID:         req.UserID,
		NodeID:          req.NodeID,
		Name:            req.Name,
		Memory:          req.Memory,
		Swap:            req.Swap,
		Disk:            req.Disk,
		IO:              req.IO,
		CPU:             req.CPU,
		Threads:         req.Threads,
		OOMKiller:       req.OOMKiller,
		EggID:           req.EggID,
		Image:           req.Image,
		Startup:         req.Startup,
		SkipScripts:     req.SkipScripts,
		DatabaseLimit:   req.DatabaseLimit,
		AllocationLimit: req.AllocationLimit,
		BackupLimit:     req.BackupLimit,
		AllocationID:    req.AllocationID,
	}
	if err := h.DB.Create(&server).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errors(err))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"object": "server", "attributes": server})
}

func (h *ServerHandler) Delete(c *gin.Context) {
	h.DB.Delete(&models.Server{}, c.Param("id"))
	c.JSON(http.StatusNoContent, nil)
}

func (h *ServerHandler) Suspend(c *gin.Context) {
	s := strPtr("suspended")
	h.DB.Model(&models.Server{}).Where("id = ?", c.Param("id")).Update("Status", s)
	c.JSON(http.StatusOK, gin.H{})
}

func (h *ServerHandler) Unsuspend(c *gin.Context) {
	h.DB.Model(&models.Server{}).Where("id = ?", c.Param("id")).Update("Status", nil)
	c.JSON(http.StatusOK, gin.H{})
}

func (h *ServerHandler) Reinstall(c *gin.Context) {
	h.DB.Model(&models.Server{}).Where("id = ?", c.Param("id")).Updates(map[string]interface{}{"Status": nil, "InstalledAt": nil})
	c.JSON(http.StatusOK, gin.H{})
}

func (h *ServerHandler) UpdateDetails(c *gin.Context) {
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		OwnerID     *uint   `json:"owner_id"`
		ExternalID  *string `json:"external_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["Name"] = *req.Name
	}
	if req.Description != nil {
		updates["Description"] = req.Description
	}
	if req.OwnerID != nil {
		updates["OwnerID"] = *req.OwnerID
	}
	if req.ExternalID != nil {
		updates["ExternalID"] = req.ExternalID
	}
	h.DB.Model(&models.Server{}).Where("id = ?", c.Param("id")).Updates(updates)
	c.JSON(http.StatusOK, gin.H{})
}

func (h *ServerHandler) UpdateBuild(c *gin.Context) {
	var req struct {
		Memory          *int64  `json:"memory"`
		Swap            *int64  `json:"swap"`
		Disk            *int64  `json:"disk"`
		IO              *int    `json:"io"`
		CPU             *int64  `json:"cpu"`
		Threads         *string `json:"threads"`
		OOMKiller       *bool   `json:"oom_killer"`
		DatabaseLimit   *int    `json:"database_limit"`
		AllocationLimit *int    `json:"allocation_limit"`
		BackupLimit     *int    `json:"backup_limit"`
		AllocationID    *uint   `json:"allocation_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	updates := map[string]interface{}{}
	if req.Memory != nil {
		updates["Memory"] = *req.Memory
	}
	if req.Swap != nil {
		updates["Swap"] = *req.Swap
	}
	if req.Disk != nil {
		updates["Disk"] = *req.Disk
	}
	if req.IO != nil {
		updates["IO"] = *req.IO
	}
	if req.CPU != nil {
		updates["CPU"] = *req.CPU
	}
	if req.Threads != nil {
		updates["Threads"] = *req.Threads
	}
	if req.OOMKiller != nil {
		updates["OOMKiller"] = *req.OOMKiller
	}
	if req.DatabaseLimit != nil {
		updates["DatabaseLimit"] = *req.DatabaseLimit
	}
	if req.AllocationLimit != nil {
		updates["AllocationLimit"] = *req.AllocationLimit
	}
	if req.BackupLimit != nil {
		updates["BackupLimit"] = *req.BackupLimit
	}
	if req.AllocationID != nil {
		updates["AllocationID"] = *req.AllocationID
	}
	h.DB.Model(&models.Server{}).Where("id = ?", c.Param("id")).Updates(updates)
	c.JSON(http.StatusOK, gin.H{})
}

func (h *ServerHandler) UpdateStartup(c *gin.Context) {
	var req struct {
		Startup     *string `json:"startup"`
		Image       *string `json:"image"`
		SkipScripts *bool   `json:"skip_scripts"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	updates := map[string]interface{}{}
	if req.Startup != nil {
		updates["Startup"] = *req.Startup
	}
	if req.Image != nil {
		updates["Image"] = *req.Image
	}
	if req.SkipScripts != nil {
		updates["SkipScripts"] = *req.SkipScripts
	}
	h.DB.Model(&models.Server{}).Where("id = ?", c.Param("id")).Updates(updates)
	c.JSON(http.StatusOK, gin.H{})
}

func errors(err error) gin.H {
	return gin.H{"errors": []gin.H{{"code": "Error", "status": "500", "detail": err.Error()}}}
}
func errorsMsg(msg string) gin.H {
	return gin.H{"errors": []gin.H{{"code": "ValidationError", "status": "400", "detail": msg}}}
}
