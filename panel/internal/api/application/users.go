package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type UserHandler struct {
	DB *gorm.DB
}

func (h *UserHandler) Index(c *gin.Context) {
	var users []models.User
	h.DB.Find(&users)
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": users})
}

func (h *UserHandler) View(c *gin.Context) {
	var user models.User
	if err := h.DB.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errors(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"object": "user", "attributes": user})
}

func (h *UserHandler) Store(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	user.UUID = genUUID()
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errors(err))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"object": "user", "attributes": user})
}

func (h *UserHandler) Update(c *gin.Context) {
	var user models.User
	if err := h.DB.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errors(err))
		return
	}
	var updates models.User
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	h.DB.Model(&user).Omit("Password", "UUID", "MFAAppSecret").Updates(updates)
	c.JSON(http.StatusOK, gin.H{"object": "user", "attributes": user})
}

func (h *UserHandler) Delete(c *gin.Context) {
	h.DB.Delete(&models.User{}, c.Param("id"))
	c.JSON(http.StatusNoContent, nil)
}
