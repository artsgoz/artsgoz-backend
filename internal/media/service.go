package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen2brain/webp"
	"github.com/google/uuid"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

var (
	ErrInvalidImage    = errors.New("invalid or unsupported image file")
	ErrFileTooLarge    = errors.New("file size exceeds maximum allowed limit")
	ErrEmptyFile       = errors.New("uploaded file is empty")
)

const (
	MaxDimension = 1920
	WebPQuality  = 80
)

type UploadResult struct {
	URL              string `json:"url"`
	Format           string `json:"format"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	Size             int64  `json:"size"`
	OriginalFilename string `json:"original_filename"`
}

type Service struct {
	storage StorageProvider
}

func NewService(storage StorageProvider) *Service {
	return &Service{storage: storage}
}

func (s *Service) OptimizeAndUpload(ctx context.Context, reader io.Reader, originalFilename string, size int64) (*UploadResult, error) {
	if size == 0 {
		return nil, ErrEmptyFile
	}

	// Read full file content
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}

	if len(data) == 0 {
		return nil, ErrEmptyFile
	}

	// 1. Decode original image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, ErrInvalidImage
	}

	origBounds := img.Bounds()
	origWidth := origBounds.Dx()
	origHeight := origBounds.Dy()

	if origWidth == 0 || origHeight == 0 {
		return nil, ErrInvalidImage
	}

	// 2. Resize if image dimensions exceed maximum threshold (1920px)
	var finalImg image.Image = img
	newWidth, newHeight := origWidth, origHeight

	if origWidth > MaxDimension || origHeight > MaxDimension {
		if origWidth > origHeight {
			newWidth = MaxDimension
			newHeight = int(float64(origHeight) * (float64(MaxDimension) / float64(origWidth)))
		} else {
			newHeight = MaxDimension
			newWidth = int(float64(origWidth) * (float64(MaxDimension) / float64(origHeight)))
		}

		dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.BiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
		finalImg = dst
	}

	// 3. Compress and encode to WebP
	var webpBuf bytes.Buffer
	opts := webp.Options{
		Quality: WebPQuality,
	}

	if err := webp.Encode(&webpBuf, finalImg, opts); err != nil {
		return nil, fmt.Errorf("encode webp: %w", err)
	}

	webpBytes := webpBuf.Bytes()

	// 4. Generate unique storage path: articles/YYYY/MM/<uuid>.webp
	now := time.Now().UTC()
	uniqueID := uuid.NewString()
	cleanExt := filepath.Ext(originalFilename)
	baseClean := strings.TrimSuffix(filepath.Base(originalFilename), cleanExt)
	if baseClean == "" {
		baseClean = "image"
	}

	relPath := fmt.Sprintf("articles/%d/%02d/%s.webp", now.Year(), int(now.Month()), uniqueID)

	// 5. Upload via Storage Provider
	url, err := s.storage.Upload(ctx, relPath, webpBytes, "image/webp")
	if err != nil {
		return nil, fmt.Errorf("upload to storage: %w", err)
	}

	return &UploadResult{
		URL:              url,
		Format:           "webp",
		Width:            newWidth,
		Height:           newHeight,
		Size:             int64(len(webpBytes)),
		OriginalFilename: originalFilename,
	}, nil
}
