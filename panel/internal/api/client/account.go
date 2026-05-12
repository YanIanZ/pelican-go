package client

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type AccountHandler struct {
	DB *gorm.DB
}

func (h *AccountHandler) Index(c *gin.Context) {
	userID := c.GetUint("user_id")
	var user models.User
	h.DB.First(&user, userID)
	c.JSON(http.StatusOK, gin.H{"object": "user", "attributes": gin.H{
		"id":         user.ID,
		"uuid":       user.UUID,
		"username":   user.Username,
		"email":      user.Email,
		"language":   user.Language,
		"timezone":   user.Timezone,
		"created_at": user.CreatedAt,
	}})
}

func (h *AccountHandler) Permissions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"permissions": []string{}}})
}

func (h *AccountHandler) UpdateEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	h.DB.Model(&models.User{}).Where("id = ?", c.GetUint("user_id")).Update("email", req.Email)
	c.JSON(http.StatusOK, gin.H{})
}

func (h *AccountHandler) UpdateUsername(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	h.DB.Model(&models.User{}).Where("id = ?", c.GetUint("user_id")).Update("username", req.Username)
	c.JSON(http.StatusOK, gin.H{})
}

func (h *AccountHandler) UpdatePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		Password        string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorsMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func errorsMsg(msg string) gin.H {
	return gin.H{"errors": []gin.H{{"code": "ValidationError", "status": "400", "detail": msg}}}
}
