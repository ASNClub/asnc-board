package http

import (
	"github.com/gin-gonic/gin"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/service"
)

type SearchHandler struct {
	svc     *service.SearchService
	userSvc *service.UserService
}

func NewSearchHandler(svc *service.SearchService, userSvc *service.UserService) *SearchHandler {
	return &SearchHandler{svc: svc, userSvc: userSvc}
}

func (h *SearchHandler) Register(r *gin.Engine) {
	r.GET("/api/v1/search/posts", h.searchPosts)
	r.GET("/api/v1/search/communities", h.searchCommunities)
	r.GET("/api/v1/search/users", h.searchUsers)
}

func (h *SearchHandler) searchPosts(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	limit, offset := pagination(c)
	result, err := h.svc.SearchPosts(c.Request.Context(), q, limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, result)
}

func (h *SearchHandler) searchCommunities(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	limit, offset := pagination(c)
	result, err := h.svc.SearchCommunities(c.Request.Context(), q, limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, result)
}

func (h *SearchHandler) searchUsers(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	limit, _ := pagination(c)
	users, err := h.userSvc.Search(c.Request.Context(), q, limit)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, users)
}
