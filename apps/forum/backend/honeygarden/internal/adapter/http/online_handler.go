package http

import (
	"github.com/gin-gonic/gin"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/service"
)

type OnlineHandler struct {
	svc     *service.OnlineService
	userSvc *service.UserService
}

func NewOnlineHandler(svc *service.OnlineService, userSvc *service.UserService) *OnlineHandler {
	return &OnlineHandler{svc: svc, userSvc: userSvc}
}

func (h *OnlineHandler) Register(r *gin.Engine, auth gin.HandlerFunc) {
	r.GET("/api/v1/online/count", h.count)
	r.GET("/api/v1/users/:username/online", h.isOnline)

	authed := r.Group("/", auth)
	authed.POST("/api/v1/users/me/heartbeat", h.heartbeat)
}

func (h *OnlineHandler) heartbeat(c *gin.Context) {
	h.svc.Heartbeat(currentUserID(c))
	response.NoContent(c)
}

func (h *OnlineHandler) isOnline(c *gin.Context) {
	user, err := h.userSvc.GetByUsername(c.Request.Context(), c.Param("username"))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{"online": h.svc.IsOnline(user.ID)})
}

func (h *OnlineHandler) count(c *gin.Context) {
	response.OK(c, gin.H{"count": h.svc.OnlineCount()})
}
