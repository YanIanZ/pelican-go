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

type AuthController struct {
	DB      *gorm.DB
	Session *auth.SessionManager
	tmpl    *template.Template
}

func NewAuthController(db *gorm.DB, sm *auth.SessionManager) *AuthController {
	tmpl := template.Must(template.ParseFS(templates.FS, "layout.html", "auth/login.html"))
	return &AuthController{DB: db, Session: sm, tmpl: tmpl}
}

func (ac *AuthController) LoginPage(c *gin.Context) {
	ac.tmpl.ExecuteTemplate(c.Writer, "layout", nil)
}

func (ac *AuthController) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	var user models.User
	if err := ac.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.String(http.StatusOK, `<span style="color:red">Invalid credentials</span>`)
		return
	}

	if !auth.VerifyPassword(password, user.Password) {
		c.String(http.StatusOK, `<span style="color:red">Invalid credentials</span>`)
		return
	}

	if user.MFAAppSecret != nil && *user.MFAAppSecret != "" {
		c.Redirect(http.StatusFound, "/auth/2fa")
		return
	}

	_, err := ac.Session.Create(c, user.ID, user.Username)
	if err != nil {
		c.String(http.StatusOK, `<span style="color:red">Failed to create session</span>`)
		return
	}

	c.Header("HX-Redirect", "/")
	c.Status(http.StatusOK)
}

func (ac *AuthController) Logout(c *gin.Context) {
	ac.Session.Destroy(c)
	c.Redirect(http.StatusFound, "/auth/login")
}
