package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/service"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) Register(r *gin.Engine, auth gin.HandlerFunc) {
	authed := r.Group("/", auth)
	authed.GET("/api/v1/notifications", h.list)
	authed.GET("/api/v1/notifications/unread-count", h.unreadCount)
	authed.PUT("/api/v1/notifications/:id/read", h.markRead)
	authed.POST("/api/v1/notifications/read-all", h.markAllRead)
	authed.GET("/api/v1/users/me/notification-preferences", h.getPreferences)
	authed.PUT("/api/v1/users/me/notification-preferences", h.setPreference)
}

func (h *NotificationHandler) list(c *gin.Context) {
	limit, offset := pagination(c)
	notifications, err := h.svc.GetNotifications(c.Request.Context(), currentUserID(c), limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, notifications)
}

func (h *NotificationHandler) unreadCount(c *gin.Context) {
	count, err := h.svc.CountUnread(c.Request.Context(), currentUserID(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, map[string]int{"count": count})
}

func (h *NotificationHandler) markRead(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.MarkRead(c.Request.Context(), id, currentUserID(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *NotificationHandler) markAllRead(c *gin.Context) {
	if err := h.svc.MarkAllRead(c.Request.Context(), currentUserID(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *NotificationHandler) getPreferences(c *gin.Context) {
	prefs, err := h.svc.GetPreferences(c.Request.Context(), currentUserID(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, prefs)
}

func (h *NotificationHandler) setPreference(c *gin.Context) {
	var body struct {
		Type    string `json:"type"`
		Enabled bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Type == "" {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err := h.svc.SetPreference(c.Request.Context(), currentUserID(c), body.Type, body.Enabled); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}
