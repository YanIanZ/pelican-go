package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type MountHandler struct {
	DB *gorm.DB
}

func (h *MountHandler) Index(c *gin.Context) {
	var mounts []models.Mount
	h.DB.Find(&mounts)
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": mounts})
}

func (h *MountHandler) View(c *gin.Context) {
	var mount models.Mount
	if err := h.DB.First(&mount, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errors(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"object": "mount", "attributes": mount})
}

func (h *MountHandler) Store(c *gin.Context) {
	var mount models.Mount
	if err := c.ShouldBindJSON(&mount); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	mount.UUID = genUUID()
	if err := h.DB.Create(&mount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errors(err))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"object": "mount", "attributes": mount})
}

func (h *MountHandler) Update(c *gin.Context) {
	var mount models.Mount
	if err := h.DB.First(&mount, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errors(err))
		return
	}
	var updates models.Mount
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	h.DB.Model(&mount).Omit("UUID").Updates(updates)
	c.JSON(http.StatusOK, gin.H{"object": "mount", "attributes": mount})
}

func (h *MountHandler) Delete(c *gin.Context) {
	h.DB.Delete(&models.Mount{}, c.Param("id"))
	c.JSON(http.StatusNoContent, nil)
}
