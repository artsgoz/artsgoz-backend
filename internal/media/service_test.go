package media

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

type storageStub struct {
	uploadedPath string
	uploadedData []byte
	contentType  string
	returnURL    string
	err          error
}

func (s *storageStub) Upload(_ context.Context, relativePath string, data []byte, contentType string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	s.uploadedPath = relativePath
	s.uploadedData = data
	s.contentType = contentType
	if s.returnURL != "" {
		return s.returnURL, nil
	}
	return "/uploads/" + relativePath, nil
}

func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 100, A: 255})
		}
	}
	return img
}

func TestOptimizeAndUpload_ValidPNG(t *testing.T) {
	t.Parallel()

	storage := &storageStub{returnURL: "https://cdn.artsgoz.com/my-article.webp"}
	svc := NewService(storage)

	img := createTestImage(200, 150)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	res, err := svc.OptimizeAndUpload(context.Background(), &buf, "cover.png", int64(buf.Len()))
	if err != nil {
		t.Fatalf("OptimizeAndUpload error: %v", err)
	}

	if res.Format != "webp" {
		t.Errorf("expected format 'webp', got %q", res.Format)
	}
	if res.Width != 200 || res.Height != 150 {
		t.Errorf("expected 200x150, got %dx%d", res.Width, res.Height)
	}
	if res.URL != "https://cdn.artsgoz.com/my-article.webp" {
		t.Errorf("expected url %q, got %q", "https://cdn.artsgoz.com/my-article.webp", res.URL)
	}
	if len(storage.uploadedData) == 0 {
		t.Fatalf("expected data to be passed to storage provider")
	}
}

func TestOptimizeAndUpload_ValidJPEG(t *testing.T) {
	t.Parallel()

	storage := &storageStub{}
	svc := NewService(storage)

	img := createTestImage(300, 200)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("jpeg encode: %v", err)
	}

	res, err := svc.OptimizeAndUpload(context.Background(), &buf, "photo.jpeg", int64(buf.Len()))
	if err != nil {
		t.Fatalf("OptimizeAndUpload error: %v", err)
	}

	if res.Format != "webp" {
		t.Errorf("expected format 'webp', got %q", res.Format)
	}
	if res.Width != 300 || res.Height != 200 {
		t.Errorf("expected 300x200, got %dx%d", res.Width, res.Height)
	}
}

func TestOptimizeAndUpload_LargeDimensionDownscaling(t *testing.T) {
	t.Parallel()

	storage := &storageStub{}
	svc := NewService(storage)

	// Create 2400x1200 image (width exceeds 1920)
	img := createTestImage(2400, 1200)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	res, err := svc.OptimizeAndUpload(context.Background(), &buf, "huge-banner.png", int64(buf.Len()))
	if err != nil {
		t.Fatalf("OptimizeAndUpload error: %v", err)
	}

	if res.Width != MaxDimension {
		t.Errorf("expected width %d, got %d", MaxDimension, res.Width)
	}
	// Aspect ratio is 2:1, so height should be 1920 / 2 = 960
	if res.Height != 960 {
		t.Errorf("expected height 960, got %d", res.Height)
	}
}

func TestOptimizeAndUpload_InvalidImage(t *testing.T) {
	t.Parallel()

	storage := &storageStub{}
	svc := NewService(storage)

	invalidData := bytes.NewReader([]byte("not an image at all text file"))
	_, err := svc.OptimizeAndUpload(context.Background(), invalidData, "corrupt.png", 29)
	if !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("expected ErrInvalidImage, got %v", err)
	}
}

func TestOptimizeAndUpload_EmptyFile(t *testing.T) {
	t.Parallel()

	storage := &storageStub{}
	svc := NewService(storage)

	emptyData := bytes.NewReader([]byte{})
	_, err := svc.OptimizeAndUpload(context.Background(), emptyData, "empty.jpg", 0)
	if !errors.Is(err, ErrEmptyFile) {
		t.Fatalf("expected ErrEmptyFile, got %v", err)
	}
}
