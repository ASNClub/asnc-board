package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/service"
)

type FeedbackHandler struct {
	svc *service.FeedbackService
}

func NewFeedbackHandler(svc *service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{svc: svc}
}

func (h *FeedbackHandler) Register(r *gin.Engine, auth, optAuth, admin gin.HandlerFunc) {
	r.GET("/api/v1/feedback", optAuth, h.list)

	authed := r.Group("/", auth)
	authed.POST("/api/v1/feedback", h.create)
	authed.POST("/api/v1/feedback/:id/vote", h.vote)
	authed.DELETE("/api/v1/feedback/:id/vote", h.unvote)

	adm := r.Group("/", auth, admin)
	adm.PATCH("/api/v1/admin/feedback/:id", h.updateStatus)
	adm.DELETE("/api/v1/admin/feedback/:id", h.deleteFeedback)
}

func (h *FeedbackHandler) list(c *gin.Context) {
	limit, offset := pagination(c)
	sort := c.Query("sort")
	status := c.Query("status")
	items, err := h.svc.List(c.Request.Context(), sort, status, currentUserIDOpt(c), limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, items)
}

func (h *FeedbackHandler) create(c *gin.Context) {
	var input domain.CreateFeedbackInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	f, err := h.svc.Create(c.Request.Context(), currentUserID(c), input)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Created(c, f)
}

func (h *FeedbackHandler) vote(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.Vote(c.Request.Context(), currentUserID(c), id); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *FeedbackHandler) unvote(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.Unvote(c.Request.Context(), currentUserID(c), id); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *FeedbackHandler) deleteFeedback(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *FeedbackHandler) updateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	var body struct {
		Status domain.FeedbackStatus `json:"status"`
	}
	if err = c.ShouldBindJSON(&body); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.UpdateStatus(c.Request.Context(), id, body.Status); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}
