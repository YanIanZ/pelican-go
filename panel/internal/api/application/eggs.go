package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type EggHandler struct {
	DB *gorm.DB
}

func (h *EggHandler) Index(c *gin.Context) {
	var eggs []models.Egg
	h.DB.Find(&eggs)
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": eggs})
}

func (h *EggHandler) View(c *gin.Context) {
	var egg models.Egg
	if err := h.DB.Preload("EggVariables").First(&egg, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errors(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"object": "egg", "attributes": egg})
}

func (h *EggHandler) Store(c *gin.Context) {
	var egg models.Egg
	if err := c.ShouldBindJSON(&egg); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	egg.UUID = genUUID()
	if err := h.DB.Create(&egg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errors(err))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"object": "egg", "attributes": egg})
}

func (h *EggHandler) Delete(c *gin.Context) {
	h.DB.Delete(&models.Egg{}, c.Param("id"))
	c.JSON(http.StatusNoContent, nil)
}
