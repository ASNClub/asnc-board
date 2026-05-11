package http

import (
	"github.com/gin-gonic/gin"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/service"
)

type BadgeHandler struct {
	svc     *service.BadgeService
	userSvc *service.UserService
}

func NewBadgeHandler(svc *service.BadgeService, userSvc *service.UserService) *BadgeHandler {
	return &BadgeHandler{svc: svc, userSvc: userSvc}
}

func (h *BadgeHandler) Register(r *gin.Engine) {
	r.GET("/api/v1/badges", h.listDefinitions)
	r.GET("/api/v1/users/:username/badges", h.getUserBadges)
}

func (h *BadgeHandler) listDefinitions(c *gin.Context) {
	defs, err := h.svc.ListDefinitions(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, defs)
}

func (h *BadgeHandler) getUserBadges(c *gin.Context) {
	user, err := h.userSvc.GetByUsername(c.Request.Context(), c.Param("username"))
	if err != nil {
		response.Err(c, domain.ErrNotFound)
		return
	}
	badges, err := h.svc.GetUserBadges(c.Request.Context(), user.ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, badges)
}
