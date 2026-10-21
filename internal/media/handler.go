package media

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxUploadSize = 15 * 1024 * 1024 // 15MB

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// UploadImage handles POST /api/upload/image
func (h *Handler) UploadImage(c *gin.Context) {
	// 1. Try "file" field, fallback to "image"
	fileHeader, err := c.FormFile("file")
	if err != nil {
		fileHeader, err = c.FormFile("image")
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid_request", "multipart form field 'file' or 'image' is required")
			return
		}
	}

	// 2. Size check
	if fileHeader.Size > maxUploadSize {
		writeError(c, http.StatusBadRequest, "file_too_large", "file size exceeds maximum limit of 15MB")
		return
	}

	// 3. Open file
	file, err := fileHeader.Open()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "file_open_error", "unable to read uploaded file")
		return
	}
	defer file.Close()

	// 4. Optimize and upload
	result, err := h.service.OptimizeAndUpload(c.Request.Context(), file, fileHeader.Filename, fileHeader.Size)
	if err != nil {
		if errors.Is(err, ErrInvalidImage) {
			writeError(c, http.StatusBadRequest, "invalid_image", "unsupported or corrupt image format")
			return
		}
		if errors.Is(err, ErrEmptyFile) {
			writeError(c, http.StatusBadRequest, "empty_file", "uploaded file is empty")
			return
		}
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
