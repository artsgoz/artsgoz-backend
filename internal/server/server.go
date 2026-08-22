package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// New creates the Gin engine. Feature routes are registered by their modules.
func New(mode string) *gin.Engine {
	gin.SetMode(mode)
	engine := gin.New()
	registerMiddleware(engine)

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return engine
}
