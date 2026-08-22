package server

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// New creates the Gin engine. Feature routes are registered by their modules.
func New(logger *slog.Logger, mode string) *gin.Engine {
	gin.SetMode(mode)
	engine := gin.New()
	registerMiddleware(engine, logger)

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return engine
}
