package httpserver

import (
	"github.com/gin-gonic/gin"
)

// New returns a Gin engine configured with default middleware (Logger and Recovery).
func New() *gin.Engine {
	return gin.Default()
}
