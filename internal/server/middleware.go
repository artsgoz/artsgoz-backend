package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDKey = "request_id"

func registerMiddleware(engine *gin.Engine, logger *slog.Logger) {
	engine.Use(requestID())
	engine.Use(requestLogger(logger))
	engine.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic recovered", "panic", recovered, requestIDKey, c.GetString(requestIDKey))
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse("internal", "internal server error"))
	}))
	engine.Use(errorHandler(logger))
}

func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(requestIDKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		logger.Info("http request",
			requestIDKey, c.GetString(requestIDKey),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(started).Milliseconds(),
		)
	}
}

func errorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}
		logger.Error("http request failed", "error", c.Errors.Last().Err, requestIDKey, c.GetString(requestIDKey))
		c.JSON(http.StatusInternalServerError, errorResponse("internal", "internal server error"))
	}
}

func errorResponse(code, message string) gin.H {
	return gin.H{"error": gin.H{"code": code, "message": message}}
}
