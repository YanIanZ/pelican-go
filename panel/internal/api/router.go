package api

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/api/application"
	clientapi "github.com/pelican-dev/panel/internal/api/client"
	"github.com/pelican-dev/panel/internal/api/middleware"
	"github.com/pelican-dev/panel/internal/api/remote"
	"github.com/pelican-dev/panel/internal/api/web"
	"github.com/pelican-dev/panel/internal/auth"
	"github.com/pelican-dev/panel/internal/config"
	"github.com/pelican-dev/panel/internal/websockets"
)

type API struct {
	Router *gin.Engine
	DB     *gorm.DB
	Config *config.Config
	JWT    *auth.JWTManager
}

func NewAPI(db *gorm.DB, cfg *config.Config, jwtManager *auth.JWTManager, rdb *redis.Client) *API {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	sessionManager := auth.NewSessionManager(rdb)
	r.Use(sessionManager.Middleware())

	wsHub := websockets.NewHub()
	wsHandler := websockets.NewHandler(wsHub)

	api := &API{
		Router: r,
		DB:     db,
		Config: cfg,
		JWT:    jwtManager,
	}

	api.registerRoutes(sessionManager, wsHandler)
	return api
}

func (api *API) registerRoutes(sm *auth.SessionManager, wsHandler *websockets.Handler) {
	r := api.Router

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	rHandler := &remote.ServerHandler{DB: api.DB}
	iHandler := &remote.InstallHandler{DB: api.DB}
	bHandler := &remote.BackupHandler{DB: api.DB}
	tHandler := &remote.TransferHandler{DB: api.DB}
	sHandler := &remote.SFTPHandler{DB: api.DB}
	aHandler := &remote.ActivityHandler{DB: api.DB}

	remoteGroup := r.Group("/api/remote", middleware.NodeAuth(api.DB))
	{
		remoteGroup.POST("/sftp/auth", sHandler.Auth)
		remoteGroup.GET("/servers", rHandler.List)
		remoteGroup.POST("/servers/reset", rHandler.ResetState)
		remoteGroup.POST("/activity", aHandler.Store)
		remoteGroup.GET("/servers/:uuid", rHandler.Get)
		remoteGroup.GET("/servers/:uuid/install", iHandler.Index)
		remoteGroup.POST("/servers/:uuid/install", iHandler.Store)
		remoteGroup.POST("/servers/:uuid/transfer/failure", tHandler.Failure)
		remoteGroup.POST("/servers/:uuid/transfer/success", tHandler.Success)
		remoteGroup.POST("/servers/:uuid/container/status", tHandler.Status)
		remoteGroup.GET("/backups/:uuid", bHandler.Upload)
		remoteGroup.POST("/backups/:uuid", bHandler.Status)
		remoteGroup.POST("/backups/:uuid/restore", bHandler.Restore)
	}

	appGroup := r.Group("/api/application", middleware.ApplicationAuth(api.DB))
	client := r.Group("/api/client", middleware.ClientAuth(api.DB))

	serverHandler := &application.ServerHandler{DB: api.DB}
	nodeHandler := &application.NodeHandler{DB: api.DB}
	userHandler := &application.UserHandler{DB: api.DB}
	eggHandler := &application.EggHandler{DB: api.DB}
	mountHandler := &application.MountHandler{DB: api.DB}
	allocHandler := &application.AllocationHandler{DB: api.DB}
	dbHostHandler := &application.DatabaseHostHandler{DB: api.DB}

	{
		appGroup.GET("/servers", serverHandler.Index)
		appGroup.GET("/servers/:id", serverHandler.View)
		appGroup.POST("/servers", serverHandler.Store)
		appGroup.DELETE("/servers/:id", serverHandler.Delete)
		appGroup.POST("/servers/:id/suspend", serverHandler.Suspend)
		appGroup.POST("/servers/:id/unsuspend", serverHandler.Unsuspend)
		appGroup.POST("/servers/:id/reinstall", serverHandler.Reinstall)
		appGroup.PATCH("/servers/:id/details", serverHandler.UpdateDetails)
		appGroup.PATCH("/servers/:id/build", serverHandler.UpdateBuild)
		appGroup.PATCH("/servers/:id/startup", serverHandler.UpdateStartup)

		appGroup.GET("/nodes", nodeHandler.Index)
		appGroup.GET("/nodes/:id", nodeHandler.View)
		appGroup.GET("/nodes/:id/configuration", nodeHandler.Configuration)
		appGroup.POST("/nodes", nodeHandler.Store)
		appGroup.PATCH("/nodes/:id", nodeHandler.Update)
		appGroup.DELETE("/nodes/:id", nodeHandler.Delete)

		appGroup.GET("/users", userHandler.Index)
		appGroup.GET("/users/:id", userHandler.View)
		appGroup.POST("/users", userHandler.Store)
		appGroup.PATCH("/users/:id", userHandler.Update)
		appGroup.DELETE("/users/:id", userHandler.Delete)

		appGroup.GET("/eggs", eggHandler.Index)
		appGroup.GET("/eggs/:id", eggHandler.View)
		appGroup.POST("/eggs", eggHandler.Store)
		appGroup.DELETE("/eggs/:id", eggHandler.Delete)

		appGroup.GET("/mounts", mountHandler.Index)
		appGroup.GET("/mounts/:id", mountHandler.View)
		appGroup.POST("/mounts", mountHandler.Store)
		appGroup.PATCH("/mounts/:id", mountHandler.Update)
		appGroup.DELETE("/mounts/:id", mountHandler.Delete)

		appGroup.GET("/nodes/:nodeID/allocations", allocHandler.Index)
		appGroup.POST("/nodes/:nodeID/allocations", allocHandler.Store)
		appGroup.DELETE("/nodes/:nodeID/allocations/:allocationID", allocHandler.Delete)

		appGroup.GET("/database-hosts", dbHostHandler.Index)
		appGroup.GET("/database-hosts/:id", dbHostHandler.View)
		appGroup.POST("/database-hosts", dbHostHandler.Store)
		appGroup.PATCH("/database-hosts/:id", dbHostHandler.Update)
		appGroup.DELETE("/database-hosts/:id", dbHostHandler.Delete)
	}

	clAccountHandler := &clientapi.AccountHandler{DB: api.DB}
	clServerHandler := &clientapi.ServerHandler{DB: api.DB}
	clFileHandler := &clientapi.FileHandler{DB: api.DB}
	clBackupHandler := &clientapi.BackupHandler{DB: api.DB}
	clScheduleHandler := &clientapi.ScheduleHandler{DB: api.DB}
	clDBHandler := &clientapi.DatabaseHandler{DB: api.DB}
	clAllocHandler := &clientapi.AllocationHandler{DB: api.DB}
	clSubuserHandler := &clientapi.SubuserHandler{DB: api.DB}

	{
		client.GET("/", clAccountHandler.Permissions)
		client.GET("/permissions", clAccountHandler.Permissions)
		client.GET("/account", clAccountHandler.Index)
		client.PUT("/account/email", clAccountHandler.UpdateEmail)
		client.PUT("/account/username", clAccountHandler.UpdateUsername)
		client.PUT("/account/password", clAccountHandler.UpdatePassword)

		client.GET("/servers/:uuid", clServerHandler.Index)
		client.POST("/servers/:uuid/power", clServerHandler.Power)
		client.POST("/servers/:uuid/command", clServerHandler.Command)
		client.POST("/servers/:uuid/settings/rename", clServerHandler.Rename)

		client.GET("/servers/:uuid/files/list", clFileHandler.List)
		client.GET("/servers/:uuid/files/contents", clFileHandler.Contents)
		client.POST("/servers/:uuid/files/write", clFileHandler.Write)
		client.POST("/servers/:uuid/files/delete", clFileHandler.Delete)
		client.POST("/servers/:uuid/files/create-folder", clFileHandler.CreateFolder)
		client.POST("/servers/:uuid/files/compress", clFileHandler.Compress)
		client.POST("/servers/:uuid/files/decompress", clFileHandler.Decompress)
		client.PUT("/servers/:uuid/files/rename", clFileHandler.Rename)
		client.POST("/servers/:uuid/files/copy", clFileHandler.Copy)
		client.GET("/servers/:uuid/files/upload", clFileHandler.Upload)
		client.POST("/servers/:uuid/files/chmod", clFileHandler.Chmod)
		client.POST("/servers/:uuid/files/pull", clFileHandler.Pull)

		client.GET("/servers/:uuid/backups", clBackupHandler.Index)
		client.POST("/servers/:uuid/backups", clBackupHandler.Store)
		client.DELETE("/servers/:uuid/backups/:backup", clBackupHandler.Delete)
		client.POST("/servers/:uuid/backups/:backup/restore", clBackupHandler.Restore)

		client.GET("/servers/:uuid/schedules", clScheduleHandler.Index)
		client.POST("/servers/:uuid/schedules", clScheduleHandler.Store)
		client.GET("/servers/:uuid/schedules/:schedule", clScheduleHandler.View)
		client.POST("/servers/:uuid/schedules/:schedule", clScheduleHandler.Update)
		client.DELETE("/servers/:uuid/schedules/:schedule", clScheduleHandler.Delete)
		client.POST("/servers/:uuid/schedules/:schedule/execute", clScheduleHandler.Execute)

		client.GET("/servers/:uuid/databases", clDBHandler.Index)
		client.POST("/servers/:uuid/databases", clDBHandler.Store)
		client.DELETE("/servers/:uuid/databases/:database", clDBHandler.Delete)

		client.GET("/servers/:uuid/network/allocations", clAllocHandler.Index)
		client.POST("/servers/:uuid/network/allocations", clAllocHandler.Store)

		client.GET("/servers/:uuid/users", clSubuserHandler.Index)
		client.POST("/servers/:uuid/users", clSubuserHandler.Store)
		client.POST("/servers/:uuid/users/:userUuid", clSubuserHandler.Update)
		client.DELETE("/servers/:uuid/users/:userUuid", clSubuserHandler.Delete)
	}

	authCtrl := web.NewAuthController(api.DB, sm)
	adminCtrl := web.NewAdminController(api.DB, sm)
	serverCtrl := web.NewServerController(api.DB, sm)

	admin := r.Group("/admin", sm.RequireAuth())
	{
		admin.GET("/", adminCtrl.Dashboard)
		admin.GET("/servers", adminCtrl.ServersList)
	}
	server := r.Group("/servers", sm.RequireAuth())
	{
		server.GET("/:uuid/console", serverCtrl.Console)
		server.GET("/:uuid/files", serverCtrl.Files)
		server.GET("/:uuid/settings", serverCtrl.Settings)
		server.GET("/:uuid/backups", serverCtrl.Backups)
	}

	r.GET("/", sm.RequireAuth(), adminCtrl.Dashboard)
	r.GET("/auth/login", authCtrl.LoginPage)
	r.POST("/auth/login", authCtrl.Login)
	r.GET("/auth/logout", authCtrl.Logout)

	// WebSocket console relay
	r.GET("/api/client/servers/:uuid/ws", wsHandler.Handle)
}
