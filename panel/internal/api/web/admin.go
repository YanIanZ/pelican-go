package web

import (
	"html/template"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/auth"
	"github.com/pelican-dev/panel/internal/models"
	"github.com/pelican-dev/panel/internal/templates"
)

type AdminController struct {
	DB      *gorm.DB
	Session *auth.SessionManager
	dashTpl *template.Template
	listTpl *template.Template
}

func NewAdminController(db *gorm.DB, sm *auth.SessionManager) *AdminController {
	dashTpl := template.Must(template.ParseFS(templates.FS, "layout.html", "admin/dashboard.html"))
	listTpl := template.Must(template.ParseFS(templates.FS, "layout.html", "admin/servers/list.html"))
	return &AdminController{DB: db, Session: sm, dashTpl: dashTpl, listTpl: listTpl}
}

func (ac *AdminController) Dashboard(c *gin.Context) {
	var serverCount, nodeCount, userCount int64
	ac.DB.Model(&models.Server{}).Count(&serverCount)
	ac.DB.Model(&models.Node{}).Count(&nodeCount)
	ac.DB.Model(&models.User{}).Count(&userCount)

	ac.dashTpl.ExecuteTemplate(c.Writer, "layout", gin.H{
		"ServerCount": serverCount,
		"NodeCount":   nodeCount,
		"UserCount":   userCount,
	})
}

func (ac *AdminController) ServersList(c *gin.Context) {
	var servers []models.Server
	ac.DB.Preload("Owner").Preload("Node").Find(&servers)
	ac.listTpl.ExecuteTemplate(c.Writer, "layout", gin.H{"Servers": servers, "User": gin.H{"Username": c.GetString("username")}})
}
