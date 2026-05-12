package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

func NodeAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"errors": []gin.H{
					{"code": "MissingAuthorizationHeader", "status": "401", "detail": "Missing authorization header"},
				},
			})
			return
		}

		parts := strings.SplitN(token, ".", 2)
		if len(parts) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"errors": []gin.H{
					{"code": "InvalidAuthorizationFormat", "status": "401", "detail": "Invalid authorization token format"},
				},
			})
			return
		}
		tokenID := parts[0]
		tokenSecret := parts[1]

		var node models.Node
		if err := db.Where("daemon_token_id = ?", tokenID).First(&node).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"errors": []gin.H{
					{"code": "InvalidNodeCredentials", "status": "401", "detail": "Invalid node credentials"},
				},
			})
			return
		}

		if subtle.ConstantTimeCompare([]byte(node.DaemonToken), []byte(tokenSecret)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"errors": []gin.H{
					{"code": "InvalidNodeCredentials", "status": "403", "detail": "Invalid node credentials"},
				},
			})
			return
		}

		c.Set("node_id", node.ID)
		c.Set("node_uuid", node.UUID)
		c.Next()
	}
}

func ApplicationAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"errors": []gin.H{
					{"code": "MissingAuthorizationHeader", "status": "401", "detail": "Missing authorization header"},
				},
			})
			return
		}

		identifier := ""
		if len(token) >= 16 {
			identifier = token[:16]
		}
		var apiKey models.APIKey
		if err := db.Where("identifier = ? AND key_type = ?", identifier, models.APIKeyTypeApplication).First(&apiKey).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"errors": []gin.H{
					{"code": "InvalidApiKey", "status": "401", "detail": "Invalid API key"},
				},
			})
			return
		}

		tokenSecret := ""
		if len(token) > 16 {
			tokenSecret = token[16:]
		}
		hashed := sha256.Sum256([]byte(tokenSecret))
		if subtle.ConstantTimeCompare([]byte(apiKey.Token), []byte(fmt.Sprintf("%x", hashed))) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"errors": []gin.H{
					{"code": "InvalidApiKey", "status": "403", "detail": "Invalid API key"},
				},
			})
			return
		}

		c.Set("api_key_id", apiKey.ID)
		if apiKey.UserID != nil {
			c.Set("user_id", *apiKey.UserID)
		}
		c.Next()
	}
}

func ClientAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"errors": []gin.H{
					{"code": "MissingAuthorizationHeader", "status": "401", "detail": "Missing authorization header"},
				},
			})
			return
		}

		identifier := ""
		if len(token) >= 16 {
			identifier = token[:16]
		}
		var apiKey models.APIKey
		if err := db.Where("identifier = ? AND key_type = ?", identifier, models.APIKeyTypeAccount).First(&apiKey).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"errors": []gin.H{
					{"code": "InvalidApiKey", "status": "401", "detail": "Invalid API key"},
				},
			})
			return
		}

		tokenSecret := ""
		if len(token) > 16 {
			tokenSecret = token[16:]
		}
		hashed := sha256.Sum256([]byte(tokenSecret))
		if subtle.ConstantTimeCompare([]byte(apiKey.Token), []byte(fmt.Sprintf("%x", hashed))) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"errors": []gin.H{
					{"code": "InvalidApiKey", "status": "403", "detail": "Invalid API key"},
				},
			})
			return
		}

		c.Set("api_key_id", apiKey.ID)
		if apiKey.UserID != nil {
			c.Set("user_id", *apiKey.UserID)
		}
		c.Next()
	}
}

func extractBearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}
