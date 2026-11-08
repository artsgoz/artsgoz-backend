package app_test

import (
	"github.com/artsgoz/artsgoz-backend/internal/app"
	"github.com/artsgoz/artsgoz-backend/internal/platform/config"
	"github.com/artsgoz/artsgoz-backend/internal/server"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnifiedRoutesAndAuthentication(t *testing.T) {
	engine := server.New(gin.TestMode)
	app.RegisterRoutes(engine, nil, nil, config.Config{})
	routes := map[string]bool{}
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, prefix := range []string{"/api", "/api/v1"} {
		for _, route := range []struct{ method, path string }{
			{"POST", "/auth/firebase"}, {"GET", "/articles"}, {"GET", "/articles/:id"}, {"POST", "/articles/:id/views"},
			{"POST", "/contacts"}, {"GET", "/faqs"}, {"GET", "/admin/contacts"}, {"GET", "/admin/reviews"},
			{"GET", "/admin/comments"}, {"PATCH", "/admin/comments/:id"}, {"GET", "/admin/tags"},
			{"POST", "/admin/media/upload"}, {"GET", "/admin/dashboard/stats"}, {"GET", "/admin/users"},
		} {
			if !routes[route.method+" "+prefix+route.path] {
				t.Errorf("missing %s %s%s", route.method, prefix, route.path)
			}
		}
		for _, path := range []string{"/admin/users", "/admin/articles", "/admin/contacts", "/admin/reviews", "/admin/dashboard/stats"} {
			request := httptest.NewRequest(http.MethodGet, prefix+path, nil)
			request.Header.Set("X-User-Role", "editor")
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Errorf("unauthenticated %s: %d", path, response.Code)
			}
		}
	}
	request := httptest.NewRequest(http.MethodOptions, "/api/auth/firebase", nil)
	request.Header.Set("Origin", "http://localhost:4300")
	request.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != 204 || response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:4300" {
		t.Fatalf("CORS failed: %d", response.Code)
	}
}
