package remote

import (
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestServerHandlerExists(t *testing.T) {
	var h *ServerHandler
	if h != nil {
		t.Error("nil handler should be nil")
	}

	h = &ServerHandler{DB: &gorm.DB{}}
	if h.DB == nil {
		t.Error("handler DB should not be nil when set")
	}
}

func TestRouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &ServerHandler{}

	r.GET("/api/remote/servers", h.List)
	r.GET("/api/remote/servers/:uuid", h.Get)
	r.POST("/api/remote/servers/reset", h.ResetState)

	routes := r.Routes()
	if len(routes) != 3 {
		t.Errorf("expected 3 routes, got %d", len(routes))
	}

	expectedPaths := map[string]bool{
		"/api/remote/servers":         false,
		"/api/remote/servers/:uuid":   false,
		"/api/remote/servers/reset":   false,
	}
	for _, route := range routes {
		expectedPaths[route.Path] = true
	}
	for path, found := range expectedPaths {
		if !found {
			t.Errorf("route not registered: %s", path)
		}
	}
}
