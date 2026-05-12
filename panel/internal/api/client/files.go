package client

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FileHandler struct {
	DB *gorm.DB
}

func (h *FileHandler) List(c *gin.Context)         { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func (h *FileHandler) Contents(c *gin.Context)     { c.JSON(http.StatusOK, gin.H{"data": gin.H{}}) }
func (h *FileHandler) Write(c *gin.Context)        { c.JSON(http.StatusOK, gin.H{}) }
func (h *FileHandler) Delete(c *gin.Context)       { c.Status(http.StatusOK) }
func (h *FileHandler) CreateFolder(c *gin.Context) { c.JSON(http.StatusOK, gin.H{}) }
func (h *FileHandler) Compress(c *gin.Context)     { c.JSON(http.StatusOK, gin.H{}) }
func (h *FileHandler) Decompress(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{}) }
func (h *FileHandler) Rename(c *gin.Context)       { c.JSON(http.StatusOK, gin.H{}) }
func (h *FileHandler) Copy(c *gin.Context)         { c.JSON(http.StatusOK, gin.H{}) }
func (h *FileHandler) Upload(c *gin.Context)       { c.JSON(http.StatusOK, gin.H{}) }
func (h *FileHandler) Chmod(c *gin.Context)        { c.JSON(http.StatusOK, gin.H{}) }
func (h *FileHandler) Pull(c *gin.Context)         { c.JSON(http.StatusOK, gin.H{}) }
