package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"
)

type Session struct {
	ID        string    `json:"id"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SessionManager struct {
	Redis      *redis.Client
	Lifetime   time.Duration
	CookieName string
}

func NewSessionManager(rdb *redis.Client) *SessionManager {
	return &SessionManager{
		Redis:      rdb,
		Lifetime:   24 * time.Hour,
		CookieName: "pelican_session",
	}
}

func (sm *SessionManager) Create(ctx *gin.Context, userID uint, username string) (*Session, error) {
	session := &Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Username:  username,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sm.Lifetime),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("marshal session: %w", err)
	}

	key := fmt.Sprintf("session:%s", session.ID)
	if err := sm.Redis.Set(ctx, key, data, sm.Lifetime).Err(); err != nil {
		return nil, fmt.Errorf("store session: %w", err)
	}

	secure := true
	if os.Getenv("PANEL_APP_ENV") == "local" || os.Getenv("PANEL_APP_ENV") == "development" {
		secure = false
	}
	ctx.SetCookie(sm.CookieName, session.ID, int(sm.Lifetime.Seconds()), "/", "", secure, true)
	return session, nil
}

func (sm *SessionManager) Get(ctx *gin.Context) (*Session, error) {
	cookie, err := ctx.Cookie(sm.CookieName)
	if err != nil {
		return nil, fmt.Errorf("no session cookie")
	}

	key := fmt.Sprintf("session:%s", cookie)
	data, err := sm.Redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		sm.Redis.Del(ctx, key)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

func (sm *SessionManager) Destroy(ctx *gin.Context) error {
	cookie, err := ctx.Cookie(sm.CookieName)
	if err != nil {
		return nil
	}
	key := fmt.Sprintf("session:%s", cookie)
	sm.Redis.Del(ctx, key)
	secure := true
	if os.Getenv("PANEL_APP_ENV") == "local" || os.Getenv("PANEL_APP_ENV") == "development" {
		secure = false
	}
	ctx.SetCookie(sm.CookieName, "", -1, "/", "", secure, true)
	return nil
}

func (sm *SessionManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := sm.Get(c)
		if err != nil {
			c.Next()
			return
		}
		c.Set("user_id", session.UserID)
		c.Set("username", session.Username)
		c.Next()
	}
}

func (sm *SessionManager) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := sm.Get(c)
		if err != nil {
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}
		c.Next()
	}
}
