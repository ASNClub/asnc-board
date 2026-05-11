package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/service"
)

type PostHandler struct {
	svc *service.PostService
}

func NewPostHandler(svc *service.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

func (h *PostHandler) Register(r *gin.Engine, auth, optAuth gin.HandlerFunc) {
	// Public (optAuth — чтобы IsVoted заполнялся, если пользователь залогинен)
	r.GET("/api/v1/communities/:slug/posts", optAuth, h.listPosts)
	r.GET("/api/v1/posts/:id", optAuth, h.getPost)
	r.GET("/api/v1/posts/:id/comments", h.listComments)

	authed := r.Group("/", auth)
	authed.POST("/api/v1/communities/:slug/posts", h.createPost)
	authed.PUT("/api/v1/posts/:id", h.updatePost)
	authed.DELETE("/api/v1/posts/:id", h.deletePost)
	authed.POST("/api/v1/posts/:id/vote", h.votePost)
	authed.DELETE("/api/v1/posts/:id/vote", h.unvotePost)
	authed.POST("/api/v1/posts/:id/pin", h.pinPost)
	authed.DELETE("/api/v1/posts/:id/pin", h.unpinPost)
	authed.POST("/api/v1/posts/:id/bookmark", h.bookmark)
	authed.DELETE("/api/v1/posts/:id/bookmark", h.unbookmark)
	authed.POST("/api/v1/posts/:id/comments", h.createComment)
	authed.DELETE("/api/v1/comments/:id", h.deleteComment)
	authed.POST("/api/v1/comments/:id/vote", h.voteComment)
	authed.DELETE("/api/v1/comments/:id/vote", h.unvoteComment)
}

func (h *PostHandler) listPosts(c *gin.Context) {
	limit, offset := pagination(c)
	var posts []domain.Post
	var err error
	if kindStr := c.Query("kind"); kindStr != "" {
		kind := domain.PostKind(kindStr)
		if !kind.Valid() {
			response.Err(c, domain.ErrInvalidInput)
			return
		}
		posts, err = h.svc.GetByCommunityKind(c.Request.Context(), c.Param("slug"), kind, limit, offset)
	} else {
		posts, err = h.svc.GetByCommunity(c.Request.Context(), c.Param("slug"), limit, offset)
	}
	if err != nil {
		response.Err(c, err)
		return
	}
	views, err := h.svc.EnrichPosts(c.Request.Context(), posts, currentUserIDOpt(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, views)
}

func (h *PostHandler) getPost(c *gin.Context) {
	post, err := h.svc.GetByIDOrShort(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Err(c, err)
		return
	}
	view, err := h.svc.EnrichPost(c.Request.Context(), post, currentUserIDOpt(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, view)
}

func (h *PostHandler) createPost(c *gin.Context) {
	var input domain.CreatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	post, err := h.svc.Create(c.Request.Context(), currentUserID(c), c.Param("slug"), input)
	if err != nil {
		response.Err(c, err)
		return
	}
	view, err := h.svc.EnrichPost(c.Request.Context(), post, currentUserIDOpt(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Created(c, view)
}

func (h *PostHandler) updatePost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	var input domain.UpdatePostInput
	if err = c.ShouldBindJSON(&input); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	post, err := h.svc.Update(c.Request.Context(), currentUserID(c), id, input)
	if err != nil {
		response.Err(c, err)
		return
	}
	view, err := h.svc.EnrichPost(c.Request.Context(), post, currentUserIDOpt(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, view)
}

func (h *PostHandler) deletePost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.Delete(c.Request.Context(), currentUserID(c), id); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *PostHandler) votePost(c *gin.Context) {
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

func (h *PostHandler) unvotePost(c *gin.Context) {
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

func (h *PostHandler) pinPost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.Pin(c.Request.Context(), currentUserID(c), id); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *PostHandler) unpinPost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.Unpin(c.Request.Context(), currentUserID(c), id); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *PostHandler) listComments(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	limit, offset := pagination(c)
	comments, err := h.svc.GetComments(c.Request.Context(), id, limit, offset)
	if err != nil {
		response.Err(c, err)
		return
	}
	views, err := h.svc.EnrichComments(c.Request.Context(), comments, currentUserIDOpt(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, views)
}

func (h *PostHandler) createComment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	var input domain.CreateCommentInput
	if err = c.ShouldBindJSON(&input); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	comment, err := h.svc.CreateComment(c.Request.Context(), currentUserID(c), id, input)
	if err != nil {
		response.Err(c, err)
		return
	}
	view, err := h.svc.EnrichComment(c.Request.Context(), comment)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Created(c, view)
}

func (h *PostHandler) deleteComment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.DeleteComment(c.Request.Context(), currentUserID(c), id); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *PostHandler) bookmark(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.Bookmark(c.Request.Context(), currentUserID(c), id); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *PostHandler) unbookmark(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.Unbookmark(c.Request.Context(), currentUserID(c), id); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *PostHandler) voteComment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.VoteComment(c.Request.Context(), currentUserID(c), id); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *PostHandler) unvoteComment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err = h.svc.UnvoteComment(c.Request.Context(), currentUserID(c), id); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}
