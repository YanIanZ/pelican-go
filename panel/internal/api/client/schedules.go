package client

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ScheduleHandler struct {
	DB *gorm.DB
}

func (h *ScheduleHandler) Index(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func (h *ScheduleHandler) Store(c *gin.Context)    { c.JSON(http.StatusCreated, gin.H{"data": gin.H{}}) }
func (h *ScheduleHandler) View(c *gin.Context)     { c.JSON(http.StatusOK, gin.H{"data": gin.H{}}) }
func (h *ScheduleHandler) Update(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{"data": gin.H{}}) }
func (h *ScheduleHandler) Delete(c *gin.Context)   { c.Status(http.StatusNoContent) }
func (h *ScheduleHandler) Execute(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"data": gin.H{}}) }
