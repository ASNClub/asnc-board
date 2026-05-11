package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/metrics"
)
 
var allowedExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

var allowedMIMEs = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

const maxUploadSize = 5 << 20

type ObjectStore interface {
	Put(ctx context.Context, key, contentType string, r io.Reader, size int64) error
}

type UploadHandler struct {
	store      ObjectStore
	publicBase string
}

func NewUploadHandler(store ObjectStore, publicBase string) *UploadHandler {
	return &UploadHandler{store: store, publicBase: strings.TrimRight(publicBase, "/")}
}

func (h *UploadHandler) Register(r *gin.Engine, auth gin.HandlerFunc) {
	authed := r.Group("/", auth)
	authed.POST("/api/v1/upload", h.upload)
}

func (h *UploadHandler) upload(c *gin.Context) {
	c.Request.Body = io.NopCloser(io.LimitReader(c.Request.Body, maxUploadSize+1))

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	defer file.Close()

	if header.Size > maxUploadSize {
		response.Err(c, domain.ErrInvalidInput)
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExts[ext] {
		response.Err(c, domain.ErrInvalidInput)
		return
	}

	// Читаем весь файл — нужен и для sniff, и для S3 (PutObject требует size).
	buf, err := io.ReadAll(file)
	if err != nil {
		response.Err(c, err)
		return
	}
	if int64(len(buf)) > maxUploadSize {
		response.Err(c, domain.ErrInvalidInput)
		return
	}

	sniffLen := 512
	if len(buf) < sniffLen {
		sniffLen = len(buf)
	}
	mime := strings.Split(http.DetectContentType(buf[:sniffLen]), ";")[0]
	if !allowedMIMEs[mime] {
		response.Err(c, domain.ErrInvalidInput)
		return
	}

	key := uuid.New().String() + ext
	if err := h.store.Put(c.Request.Context(), key, mime, bytes.NewReader(buf), int64(len(buf))); err != nil {
		response.Err(c, err)
		return
	}

	url := fmt.Sprintf("%s/%s", h.publicBase, key)
	metrics.UploadsTotal.Inc()
	response.OK(c, gin.H{"url": url})
}
