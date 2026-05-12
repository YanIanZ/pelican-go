package web

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/auth"
	"github.com/pelican-dev/panel/internal/models"
	"github.com/pelican-dev/panel/internal/templates"
)

type ServerController struct {
	DB          *gorm.DB
	Session     *auth.SessionManager
	consoleTpl  *template.Template
	filesTpl    *template.Template
	settingsTpl *template.Template
	backupsTpl  *template.Template
}

func NewServerController(db *gorm.DB, sm *auth.SessionManager) *ServerController {
	consoleTpl := template.Must(template.ParseFS(templates.FS, "layout.html", "server/console.html"))
	filesTpl := template.Must(template.ParseFS(templates.FS, "layout.html", "server/files.html"))
	settingsTpl := template.Must(template.ParseFS(templates.FS, "layout.html", "server/settings.html"))
	backupsTpl := template.Must(template.ParseFS(templates.FS, "layout.html", "server/backups.html"))
	return &ServerController{
		DB:          db,
		Session:     sm,
		consoleTpl:  consoleTpl,
		filesTpl:    filesTpl,
		settingsTpl: settingsTpl,
		backupsTpl:  backupsTpl,
	}
}

func (sc *ServerController) Console(c *gin.Context) {
	uuid := c.Param("uuid")
	var server models.Server
	if err := sc.DB.Where("uuid = ?", uuid).First(&server).Error; err != nil {
		c.String(http.StatusNotFound, "Server not found")
		return
	}
	sc.consoleTpl.ExecuteTemplate(c.Writer, "layout", gin.H{
		"Server": server,
		"User":   gin.H{"Username": c.GetString("username")},
	})
}

func (sc *ServerController) Files(c *gin.Context) {
	uuid := c.Param("uuid")
	var server models.Server
	sc.DB.Where("uuid = ?", uuid).First(&server)
	sc.filesTpl.ExecuteTemplate(c.Writer, "layout", gin.H{
		"Server": server,
		"User":   gin.H{"Username": c.GetString("username")},
	})
}

func (sc *ServerController) Settings(c *gin.Context) {
	uuid := c.Param("uuid")
	var server models.Server
	sc.DB.Where("uuid = ?", uuid).First(&server)
	sc.settingsTpl.ExecuteTemplate(c.Writer, "layout", gin.H{
		"Server": server,
		"User":   gin.H{"Username": c.GetString("username")},
	})
}

func (sc *ServerController) Backups(c *gin.Context) {
	uuid := c.Param("uuid")
	var server models.Server
	sc.DB.Where("uuid = ?", uuid).First(&server)
	sc.backupsTpl.ExecuteTemplate(c.Writer, "layout", gin.H{
		"Server": server,
		"User":   gin.H{"Username": c.GetString("username")},
	})
}
