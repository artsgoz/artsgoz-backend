package media

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	upload := router.Group("/upload")
	{
		upload.POST("/image", handler.UploadImage)
	}
}
