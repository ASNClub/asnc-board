package http

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/service"
)

type FeedHandler struct {
	svc     *service.FeedService
	postSvc *service.PostService
}

func NewFeedHandler(svc *service.FeedService, postSvc *service.PostService) *FeedHandler {
	return &FeedHandler{svc: svc, postSvc: postSvc}
}

func (h *FeedHandler) Register(r *gin.Engine, auth, optAuth, admin gin.HandlerFunc) {
	r.GET("/api/v1/feed", optAuth, h.getFeed)
	r.GET("/api/v1/trending", h.getTrending)

	adminGrp := r.Group("/admin/feed", auth, admin)
	adminGrp.POST("/sources", h.addSource)
	adminGrp.GET("/sources", h.listSources)
}

func (h *FeedHandler) getFeed(c *gin.Context) {
	cursor := parseCursor(c)
	limit := parseLimit(c)

	var userID *uuid.UUID
	if u := currentUser(c); u != nil {
		id := u.ID
		userID = &id
	}

	items, err := h.svc.GetFeed(c.Request.Context(), userID, cursor, limit)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, items)
}

func (h *FeedHandler) getTrending(c *gin.Context) {
	limit := parseLimit(c)
	posts, err := h.postSvc.GetTrending(c.Request.Context(), limit)
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

func (h *FeedHandler) addSource(c *gin.Context) {
	var input domain.CreateSourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	src, err := h.svc.AddSource(c.Request.Context(), input)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Created(c, src)
}

func (h *FeedHandler) listSources(c *gin.Context) {
	sources, err := h.svc.ListSources(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, sources)
}

func parseCursor(c *gin.Context) *time.Time {
	v := c.Query("cursor")
	if v == "" {
		return nil
	}
	if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
		t := time.Unix(ts, 0).UTC()
		return &t
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return &t
	}
	return nil
}

func parseLimit(c *gin.Context) int {
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			return n
		}
	}
	return 20
}
