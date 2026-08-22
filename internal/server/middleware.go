package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDKey = "request_id"

func registerMiddleware(engine *gin.Engine) {
	engine.Use(requestID())
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(errorHandler())
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

func errorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("internal", "internal server error"))
	}
}

func errorResponse(code, message string) gin.H {
	return gin.H{"error": gin.H{"code": code, "message": message}}
}
