package http

import (
	"github.com/gin-gonic/gin"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/service"
)

type WakapiHandler struct {
	svc     *service.WakapiService
	userSvc *service.UserService
}

func NewWakapiHandler(svc *service.WakapiService, userSvc *service.UserService) *WakapiHandler {
	return &WakapiHandler{svc: svc, userSvc: userSvc}
}

func (h *WakapiHandler) Register(r *gin.Engine, auth gin.HandlerFunc) {
	r.GET("/api/v1/users/:username/wakapi", h.getStats)

	authed := r.Group("/", auth)
	authed.POST("/api/v1/users/me/wakapi", h.connect)
	authed.DELETE("/api/v1/users/me/wakapi", h.disconnect)
}

func (h *WakapiHandler) connect(c *gin.Context) {
	var in service.ConnectWakapiInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Err(c, err)
		return
	}
	if err := h.svc.Connect(c.Request.Context(), currentUserID(c), in); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *WakapiHandler) disconnect(c *gin.Context) {
	if err := h.svc.Disconnect(c.Request.Context(), currentUserID(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *WakapiHandler) getStats(c *gin.Context) {
	user, err := h.userSvc.GetByUsername(c.Request.Context(), c.Param("username"))
	if err != nil {
		response.Err(c, err)
		return
	}

	if !h.svc.IsConnected(c.Request.Context(), user.ID) {
		response.OK(c, gin.H{"connected": false})
		return
	}

	stats, err := h.svc.GetStats(c.Request.Context(), user.ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{"connected": true, "stats": stats})
}
