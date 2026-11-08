package media

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestMediaRouter() (*gin.Engine, *storageStub) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	storage := &storageStub{returnURL: "/uploads/articles/demo.webp"}
	svc := NewService(storage)
	handler := NewHandler(svc)

	api := engine.Group("/api")
	RegisterRoutes(api, handler)

	return engine, storage
}

func createTestPNGBytes(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestUploadImage_Success_FileField(t *testing.T) {
	t.Parallel()

	router, _ := setupTestMediaRouter()

	pngData := createTestPNGBytes(100, 100)

	var reqBody bytes.Buffer
	writer := multipart.NewWriter(&reqBody)
	part, err := writer.CreateFormFile("file", "test-cover.png")
	if err != nil {
		t.Fatalf("CreateFormFile error: %v", err)
	}
	_, _ = part.Write(pngData)
	_ = writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/upload/image", &reqBody)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res UploadResult
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if res.Format != "webp" || res.URL == "" || res.Width != 100 || res.Height != 100 {
		t.Fatalf("unexpected upload result: %#v", res)
	}
}

func TestUploadImage_Success_ImageField(t *testing.T) {
	t.Parallel()

	router, _ := setupTestMediaRouter()

	pngData := createTestPNGBytes(120, 80)

	var reqBody bytes.Buffer
	writer := multipart.NewWriter(&reqBody)
	part, _ := writer.CreateFormFile("image", "article-body.png")
	_, _ = part.Write(pngData)
	_ = writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/upload/image", &reqBody)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUploadImage_MissingFile(t *testing.T) {
	t.Parallel()

	router, _ := setupTestMediaRouter()

	var reqBody bytes.Buffer
	writer := multipart.NewWriter(&reqBody)
	_ = writer.WriteField("other_field", "value")
	_ = writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/upload/image", &reqBody)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d", rec.Code)
	}
}

func TestUploadImage_InvalidFormat(t *testing.T) {
	t.Parallel()

	router, _ := setupTestMediaRouter()

	var reqBody bytes.Buffer
	writer := multipart.NewWriter(&reqBody)
	part, _ := writer.CreateFormFile("file", "document.pdf")
	_, _ = part.Write([]byte("%PDF-1.4 not an image"))
	_ = writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/upload/image", &reqBody)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request for non-image, got %d", rec.Code)
	}
}
