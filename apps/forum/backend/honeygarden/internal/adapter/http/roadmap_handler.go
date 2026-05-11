package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/service"
)

type RoadmapHandler struct {
	svc *service.RoadmapService
}

func NewRoadmapHandler(svc *service.RoadmapService) *RoadmapHandler {
	return &RoadmapHandler{svc: svc}
}

func (h *RoadmapHandler) Register(r *gin.Engine, auth, admin gin.HandlerFunc) {
	r.GET("/api/v1/roadmap", h.list)

	adm := r.Group("/", auth, admin)
	adm.POST("/api/v1/admin/roadmap", h.create)
	adm.PUT("/api/v1/admin/roadmap/:id", h.update)
	adm.DELETE("/api/v1/admin/roadmap/:id", h.delete)
}

func (h *RoadmapHandler) list(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	if items == nil {
		items = []domain.RoadmapItem{}
	}
	response.OK(c, items)
}

func (h *RoadmapHandler) create(c *gin.Context) {
	var input domain.CreateRoadmapItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	item, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Created(c, item)
}

func (h *RoadmapHandler) update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	var input domain.UpdateRoadmapItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	item, err := h.svc.Update(c.Request.Context(), id, input)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, item)
}

func (h *RoadmapHandler) delete(c *gin.Context) {
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
