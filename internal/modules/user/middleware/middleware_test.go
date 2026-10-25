package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	router := gin.New()
	router.Use(middleware.AuthMiddleware(nil))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for missing header, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	router := gin.New()
	router.Use(middleware.AuthMiddleware(nil))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic some-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for invalid auth format, got %d", w.Code)
	}
}

func TestRequireRole_MissingUID(t *testing.T) {
	router := gin.New()
	router.Use(middleware.RequireRole(nil, "admin"))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 when uid is missing, got %d", w.Code)
	}
}

func TestRequireRole_EmptyUID(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("uid", "")
		c.Next()
	})
	router.Use(middleware.RequireRole(nil, "admin"))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 when uid is empty string, got %d", w.Code)
	}
}
