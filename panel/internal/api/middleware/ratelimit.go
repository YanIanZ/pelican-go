package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	Client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{Client: client}
}

func (r *RateLimiter) Limit(maxRequests int, windowSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if r.Client == nil {
			c.Next()
			return
		}

		key := "rate:" + c.ClientIP() + ":" + c.FullPath()
		count, err := r.Client.Incr(c, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			r.Client.Expire(c, key, time.Duration(windowSeconds)*time.Second)
		}

		if count > int64(maxRequests) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"errors": []gin.H{{"code": "TooManyRequests", "status": "429", "detail": "Rate limit exceeded. Try again in " + strconv.Itoa(windowSeconds) + " seconds."}},
			})
			return
		}

		c.Next()
	}
}
