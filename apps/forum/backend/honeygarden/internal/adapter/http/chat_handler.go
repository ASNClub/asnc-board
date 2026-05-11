package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/service"
)

type ChatHandler struct {
	svc      *service.ChatService
	upgrader websocket.Upgrader
	log      zerolog.Logger
}

func NewChatHandler(svc *service.ChatService, allowedOrigin string, log zerolog.Logger) *ChatHandler {
	return &ChatHandler{
		svc: svc,
		log: log,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				return origin == "" || origin == allowedOrigin
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 4096,
		},
	}
}

func (h *ChatHandler) Register(r *gin.Engine, auth gin.HandlerFunc) {
	r.GET("/api/v1/chat/messages", h.list)
	authed := r.Group("/", auth)
	authed.POST("/api/v1/chat/messages", h.send)
	authed.GET("/api/v1/chat/ws", h.ws)
}

func (h *ChatHandler) list(c *gin.Context) {
	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	msgs, err := h.svc.List(c.Request.Context(), limit)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, msgs)
}

func (h *ChatHandler) send(c *gin.Context) {
	var body struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	m, err := h.svc.Send(c.Request.Context(), currentUserID(c), body.Content)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, m)
}

func (h *ChatHandler) ws(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Warn().Err(err).Msg("chat ws upgrade failed")
		return
	}
	defer conn.Close()

	conn.SetReadLimit(2048)
	conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(70 * time.Second))
		return nil
	})

	sub := h.svc.Subscribe()
	defer h.svc.Unsubscribe(sub)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.NextReader(); err != nil {
				return
			}
		}
	}()

	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()

	for {
		select {
		case msg, ok := <-sub:
			if !ok {
				return
			}
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ping.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}
