package middleware

import "github.com/gin-gonic/gin"

func RateLimit(maxRequests int, windowSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
