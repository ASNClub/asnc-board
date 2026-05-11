package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/service"
)

type AdminHandler struct {
	userSvc      *service.UserService
	postSvc      *service.PostService
	communitySvc *service.CommunityService
}

func NewAdminHandler(userSvc *service.UserService, postSvc *service.PostService, communitySvc *service.CommunityService) *AdminHandler {
	return &AdminHandler{userSvc: userSvc, postSvc: postSvc, communitySvc: communitySvc}
}

func (h *AdminHandler) Register(r *gin.Engine, auth, admin gin.HandlerFunc) {
	adm := r.Group("/api/v1/admin", auth, admin)
	adm.POST("/users/:username/ban", h.banUser)
	adm.DELETE("/users/:username/ban", h.unbanUser)
	adm.DELETE("/posts/:id", h.deletePost)
	adm.DELETE("/comments/:id", h.deleteComment)
	adm.DELETE("/communities/:slug", h.deleteCommunity)
}

func (h *AdminHandler) banUser(c *gin.Context) {
	if err := h.userSvc.BanUser(c.Request.Context(), c.Param("username"), true); err != nil {
		response.Err(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) unbanUser(c *gin.Context) {
	if err := h.userSvc.BanUser(c.Request.Context(), c.Param("username"), false); err != nil {
		response.Err(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) deletePost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}

	if err := h.postSvc.AdminDelete(c.Request.Context(), id); err != nil {
		response.Err(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) deleteComment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err := h.postSvc.AdminDeleteComment(c.Request.Context(), id); err != nil {
		response.Err(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) deleteCommunity(c *gin.Context) {
	if err := h.communitySvc.AdminDelete(c.Request.Context(), c.Param("slug")); err != nil {
		response.Err(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
