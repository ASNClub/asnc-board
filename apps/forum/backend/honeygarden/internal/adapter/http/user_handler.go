package http

import (
	"github.com/gin-gonic/gin"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/port"
	"honeygarden/internal/service"
)

type UserHandler struct {
	svc           *service.UserService
	postSvc       *service.PostService
	activity      port.ActivityRepository
	adminAuthIDs  map[string]struct{}
}

func NewUserHandler(svc *service.UserService, postSvc *service.PostService, activity port.ActivityRepository, adminAuthIDs []string) *UserHandler {
	m := make(map[string]struct{}, len(adminAuthIDs))
	for _, id := range adminAuthIDs {
		m[id] = struct{}{}
	}
	return &UserHandler{svc: svc, postSvc: postSvc, activity: activity, adminAuthIDs: m}
}

func (h *UserHandler) Register(r *gin.Engine, auth, optAuth gin.HandlerFunc) {
	// Public
	r.GET("/api/v1/users/:username", h.getUser)
	r.GET("/api/v1/users/:username/followers", h.getFollowers)
	r.GET("/api/v1/users/:username/following", h.getFollowing)
	r.GET("/api/v1/users/:username/repos", h.getRepos)
	r.GET("/api/v1/users/:username/posts", h.getUserPosts)
	r.GET("/api/v1/users/:username/activity", h.getUserActivity)

	authed := r.Group("/", auth)
	authed.GET("/api/v1/users/me", h.getMe)
	authed.PUT("/api/v1/users/me", h.updateMe)
	authed.PUT("/api/v1/users/me/tags", h.setTags)
	authed.PUT("/api/v1/users/me/platforms", h.setPlatforms)

	authed.GET("/api/v1/users/me/bookmarks", h.getBookmarks)

	authed.POST("/api/v1/users/:username/follow", h.follow)
	authed.DELETE("/api/v1/users/:username/follow", h.unfollow)

	authed.POST("/api/v1/users/:username/block", h.block)
	authed.DELETE("/api/v1/users/:username/block", h.unblock)
	authed.GET("/api/v1/users/me/blocks", h.listBlocks)

	authed.POST("/api/v1/users/:username/friendship", h.requestFriendship)
	authed.POST("/api/v1/users/me/friendship/:username/accept", h.acceptFriendship)
	authed.POST("/api/v1/users/me/friendship/:username/reject", h.rejectFriendship)
	authed.GET("/api/v1/users/me/friends", h.getFriends)
	authed.GET("/api/v1/users/me/friendship/pending", h.getPendingRequests)

}

func (h *UserHandler) getBookmarks(c *gin.Context) {
	limit, offset := pagination(c)
	posts, err := h.postSvc.GetBookmarks(c.Request.Context(), currentUserID(c), limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	views, err := h.postSvc.EnrichPosts(c.Request.Context(), posts, currentUserIDOpt(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, views)
}

func (h *UserHandler) getUserPosts(c *gin.Context) {
	user, err := h.svc.GetByUsername(c.Request.Context(), c.Param("username"))
	if err != nil {
		response.Err(c, err)
		return
	}
	limit, offset := pagination(c)
	posts, err := h.postSvc.GetByAuthor(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	views, err := h.postSvc.EnrichPosts(c.Request.Context(), posts, currentUserIDOpt(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, views)
}

func (h *UserHandler) getUserActivity(c *gin.Context) {
	user, err := h.svc.GetByUsername(c.Request.Context(), c.Param("username"))
	if err != nil {
		response.Err(c, err)
		return
	}
	limit, offset := pagination(c)
	items, err := h.activity.GetByUser(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, items)
}

func (h *UserHandler) getMe(c *gin.Context) {
	u := currentUser(c)
	_, isAdmin := h.adminAuthIDs[u.AuthID]
	type meResponse struct {
		*domain.User
		IsAdmin bool `json:"isAdmin"`
	}
	response.OK(c, meResponse{User: u, IsAdmin: isAdmin})
}

func (h *UserHandler) getUser(c *gin.Context) {
	user, err := h.svc.GetByUsername(c.Request.Context(), c.Param("username"))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, user)
}

func (h *UserHandler) updateMe(c *gin.Context) {
	var input domain.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	user, err := h.svc.UpdateMe(c.Request.Context(), currentUserID(c), input)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, user)
}

func (h *UserHandler) setTags(c *gin.Context) {
	var body struct {
		Tags []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err := h.svc.SetTags(c.Request.Context(), currentUserID(c), body.Tags); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *UserHandler) setPlatforms(c *gin.Context) {
	var platforms []domain.UserPlatform
	if err := c.ShouldBindJSON(&platforms); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err := h.svc.SetPlatforms(c.Request.Context(), currentUserID(c), platforms); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *UserHandler) follow(c *gin.Context) {
	if err := h.svc.Follow(c.Request.Context(), currentUserID(c), c.Param("username")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *UserHandler) unfollow(c *gin.Context) {
	if err := h.svc.Unfollow(c.Request.Context(), currentUserID(c), c.Param("username")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *UserHandler) block(c *gin.Context) {
	if err := h.svc.Block(c.Request.Context(), currentUserID(c), c.Param("username")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *UserHandler) unblock(c *gin.Context) {
	if err := h.svc.Unblock(c.Request.Context(), currentUserID(c), c.Param("username")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *UserHandler) listBlocks(c *gin.Context) {
	users, err := h.svc.ListBlocks(c.Request.Context(), currentUserID(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, users)
}

func (h *UserHandler) getFollowers(c *gin.Context) {
	limit, offset := pagination(c)
	users, err := h.svc.GetFollowers(c.Request.Context(), c.Param("username"), limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, users)
}

func (h *UserHandler) getFollowing(c *gin.Context) {
	limit, offset := pagination(c)
	users, err := h.svc.GetFollowing(c.Request.Context(), c.Param("username"), limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, users)
}

func (h *UserHandler) requestFriendship(c *gin.Context) {
	if err := h.svc.RequestFriendship(c.Request.Context(), currentUserID(c), c.Param("username")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *UserHandler) acceptFriendship(c *gin.Context) {
	if err := h.svc.AcceptFriendship(c.Request.Context(), currentUserID(c), c.Param("username")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *UserHandler) rejectFriendship(c *gin.Context) {
	if err := h.svc.RejectFriendship(c.Request.Context(), currentUserID(c), c.Param("username")); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *UserHandler) getFriends(c *gin.Context) {
	limit, offset := pagination(c)
	users, err := h.svc.GetFriends(c.Request.Context(), currentUserID(c), limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, users)
}

func (h *UserHandler) getPendingRequests(c *gin.Context) {
	users, err := h.svc.GetPendingRequesters(c.Request.Context(), currentUserID(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, users)
}

func (h *UserHandler) getRepos(c *gin.Context) {
	user, err := h.svc.GetByUsername(c.Request.Context(), c.Param("username"))
	if err != nil {
		response.Err(c, err)
		return
	}
	repos, err := h.svc.GetPinnedRepos(c.Request.Context(), user.ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, repos)
}
