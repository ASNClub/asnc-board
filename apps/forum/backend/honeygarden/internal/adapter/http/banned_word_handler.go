package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/service"
)

type BannedWordHandler struct {
	svc *service.BannedWordService
}

func NewBannedWordHandler(svc *service.BannedWordService) *BannedWordHandler {
	return &BannedWordHandler{svc: svc}
}

func (h *BannedWordHandler) Register(r *gin.Engine, auth, admin gin.HandlerFunc) {
	adm := r.Group("/", auth, admin)
	adm.GET("/api/v1/admin/banned-words", h.list)
	adm.POST("/api/v1/admin/banned-words", h.create)
	adm.DELETE("/api/v1/admin/banned-words/:id", h.delete)
}

func (h *BannedWordHandler) list(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	if items == nil {
		items = []domain.BannedWord{}
	}
	response.OK(c, items)
}

func (h *BannedWordHandler) create(c *gin.Context) {
	var input domain.CreateBannedWordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	bw, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Created(c, bw)
}

func (h *BannedWordHandler) delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}
