package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/api/middleware"
	"github.com/pelican-dev/panel/internal/auth"
	"github.com/pelican-dev/panel/internal/config"
)

type API struct {
	Router *gin.Engine
	DB     *gorm.DB
	Config *config.Config
	JWT    *auth.JWTManager
}

func NewAPI(db *gorm.DB, cfg *config.Config, jwtManager *auth.JWTManager) *API {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	api := &API{
		Router: r,
		DB:     db,
		Config: cfg,
		JWT:    jwtManager,
	}

	api.registerRoutes()
	return api
}

func (api *API) registerRoutes() {
	r := api.Router

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	remote := r.Group("/api/remote", middleware.NodeAuth(api.DB))
	application := r.Group("/api/application", middleware.ApplicationAuth(api.DB))
	client := r.Group("/api/client", middleware.ClientAuth(api.DB))

	_ = remote
	_ = application
	_ = client
}
