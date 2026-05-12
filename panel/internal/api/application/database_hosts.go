package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type DatabaseHostHandler struct {
	DB *gorm.DB
}

func (h *DatabaseHostHandler) Index(c *gin.Context) {
	var hosts []models.DatabaseHost
	h.DB.Find(&hosts)
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": hosts})
}

func (h *DatabaseHostHandler) View(c *gin.Context) {
	var host models.DatabaseHost
	if err := h.DB.First(&host, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errors(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"object": "database_host", "attributes": host})
}

func (h *DatabaseHostHandler) Store(c *gin.Context) {
	var host models.DatabaseHost
	if err := c.ShouldBindJSON(&host); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	if err := h.DB.Create(&host).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errors(err))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"object": "database_host", "attributes": host})
}

func (h *DatabaseHostHandler) Update(c *gin.Context) {
	var host models.DatabaseHost
	if err := h.DB.First(&host, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errors(err))
		return
	}
	var updates models.DatabaseHost
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	h.DB.Model(&host).Omit("Password").Updates(updates)
	c.JSON(http.StatusOK, gin.H{"object": "database_host", "attributes": host})
}

func (h *DatabaseHostHandler) Delete(c *gin.Context) {
	h.DB.Delete(&models.DatabaseHost{}, c.Param("id"))
	c.JSON(http.StatusNoContent, nil)
}
