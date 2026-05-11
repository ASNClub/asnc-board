package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/service"
)

type CommunityHandler struct {
	svc *service.CommunityService
}

func NewCommunityHandler(svc *service.CommunityService) *CommunityHandler {
	return &CommunityHandler{svc: svc}
}

func (h *CommunityHandler) Register(r *gin.Engine, auth, optAuth gin.HandlerFunc) {
	r.GET("/api/v1/communities", h.listCommunities)
	r.GET("/api/v1/communities/:slug", optAuth, h.getCommunity)

	authed := r.Group("/", auth)
	authed.POST("/api/v1/communities", h.createCommunity)
	authed.GET("/api/v1/communities/me", h.getMyCommunity)
	authed.PUT("/api/v1/communities/:slug", h.updateCommunity)
	authed.DELETE("/api/v1/communities/:slug", h.deleteCommunity)
	authed.PUT("/api/v1/communities/:slug/tags", h.setTags)

	authed.POST("/api/v1/communities/:slug/follow", h.follow)
	authed.DELETE("/api/v1/communities/:slug/follow", h.unfollow)
	authed.GET("/api/v1/users/me/communities", h.getFollowed)
	r.GET("/api/v1/communities/:slug/members", h.getMembers)

	authed.POST("/api/v1/communities/:slug/star", h.star)
	authed.DELETE("/api/v1/communities/:slug/star", h.unstar)

	r.GET("/api/v1/communities/:slug/moderators", h.getModerators)
	authed.POST("/api/v1/communities/:slug/moderators", h.addModerator)
	authed.DELETE("/api/v1/communities/:slug/moderators/:username", h.removeModerator)

	authed.POST("/api/v1/communities/:slug/bans", h.ban)
	authed.DELETE("/api/v1/communities/:slug/bans/:userID", h.unban)
	authed.GET("/api/v1/communities/:slug/bans", h.getBans)
}

func (h *CommunityHandler) listCommunities(c *gin.Context) {
	limit, offset := pagination(c)
	sort := c.Query("sort")
	communities, err := h.svc.List(c.Request.Context(), sort, limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, communities)
}

func (h *CommunityHandler) getCommunity(c *gin.Context) {
	comm, err := h.svc.GetBySlugForViewer(c.Request.Context(), c.Param("slug"), currentUserIDOpt(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, comm)
}

func (h *CommunityHandler) getMyCommunity(c *gin.Context) {
	comm, err := h.svc.GetByOwner(c.Request.Context(), currentUserID(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, comm)
}

func (h *CommunityHandler) createCommunity(c *gin.Context) {
	var input domain.CreateCommunityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	comm, err := h.svc.Create(c.Request.Context(), currentUserID(c), input)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Created(c, comm)
}

func (h *CommunityHandler) updateCommunity(c *gin.Context) {
	var input domain.UpdateCommunityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	comm, err := h.svc.Update(c.Request.Context(), currentUserID(c), c.Param("slug"), input)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, comm)
}

func (h *CommunityHandler) deleteCommunity(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), currentUserID(c), c.Param("slug")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *CommunityHandler) setTags(c *gin.Context) {
	var body struct {
		Tags []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err := h.svc.SetTags(c.Request.Context(), currentUserID(c), c.Param("slug"), body.Tags); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *CommunityHandler) follow(c *gin.Context) {
	if err := h.svc.Follow(c.Request.Context(), currentUserID(c), c.Param("slug")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *CommunityHandler) unfollow(c *gin.Context) {
	if err := h.svc.Unfollow(c.Request.Context(), currentUserID(c), c.Param("slug")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *CommunityHandler) getFollowed(c *gin.Context) {
	limit, offset := pagination(c)
	communities, err := h.svc.GetFollowed(c.Request.Context(), currentUserID(c), limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, communities)
}

func (h *CommunityHandler) star(c *gin.Context) {
	if err := h.svc.Star(c.Request.Context(), currentUserID(c), c.Param("slug")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *CommunityHandler) unstar(c *gin.Context) {
	if err := h.svc.Unstar(c.Request.Context(), currentUserID(c), c.Param("slug")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *CommunityHandler) getMembers(c *gin.Context) {
	limit, offset := pagination(c)
	users, err := h.svc.GetMembers(c.Request.Context(), c.Param("slug"), limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, users)
}

func (h *CommunityHandler) getModerators(c *gin.Context) {
	mods, err := h.svc.GetModerators(c.Request.Context(), c.Param("slug"))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, mods)
}

func (h *CommunityHandler) addModerator(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if body.Role == "" {
		body.Role = "moderator"
	}
	if err := h.svc.AddModerator(c.Request.Context(), currentUserID(c), c.Param("slug"), body.Username, body.Role); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *CommunityHandler) removeModerator(c *gin.Context) {
	if err := h.svc.RemoveModerator(c.Request.Context(), currentUserID(c), c.Param("slug"), c.Param("username")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *CommunityHandler) ban(c *gin.Context) {
	var body struct {
		UserID    string         `json:"user_id"`
		Type      domain.BanType `json:"type"`
		Reason    *string        `json:"reason"`
		ExpiresAt *string        `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	targetID, err := uuid.Parse(body.UserID)
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if body.Type != domain.BanTypeBan && body.Type != domain.BanTypeMute {
		body.Type = domain.BanTypeBan
	}
	var expiresAt *time.Time
	if body.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *body.ExpiresAt)
		if err != nil {
			response.Err(c, domain.ErrInvalidInput)
			return
		}
		expiresAt = &t
	}
	if err = h.svc.Ban(c.Request.Context(), currentUserID(c), c.Param("slug"), targetID, body.Type, body.Reason, expiresAt); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *CommunityHandler) unban(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.Unban(c.Request.Context(), currentUserID(c), c.Param("slug"), targetID); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *CommunityHandler) getBans(c *gin.Context) {
	bans, err := h.svc.GetBans(c.Request.Context(), currentUserID(c), c.Param("slug"))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, bans)
}
