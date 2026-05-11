package http

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/domain"
	"honeygarden/internal/service"
)

type IntegrationHandler struct {
	svc           *service.IntegrationService
	frontendURL   string // куда редиректить после OAuth-callback
}

// NewIntegrationHandler — frontendURL: например https://honeygarden.space.
// На него редиректит /callback с параметрами `?integration=<provider>&status=ok|error&msg=...`.
func NewIntegrationHandler(svc *service.IntegrationService, frontendURL string) *IntegrationHandler {
	return &IntegrationHandler{svc: svc, frontendURL: frontendURL}
}

func (h *IntegrationHandler) Register(r *gin.Engine, auth gin.HandlerFunc) {
	// Callback приходит из браузера юзера от внешнего провайдера — без auth.
	// Идентификация — через подписанный state.
	r.GET("/api/v1/integrations/:provider/callback", h.callback)

	authed := r.Group("/", auth)
	authed.GET("/api/v1/me/integrations", h.list)
	authed.POST("/api/v1/integrations/:provider/connect", h.beginConnect)
	authed.DELETE("/api/v1/me/integrations/:provider", h.disconnect)
	authed.GET("/api/v1/integrations/:provider/repos", h.listRepos)
	authed.PUT("/api/v1/me/pinned-repos", h.setPins)
}

func parseProvider(s string) (domain.GitProvider, bool) {
	switch domain.GitProvider(s) {
	case domain.GitProviderGitHub, domain.GitProviderGitLab, domain.GitProviderCodeberg:
		return domain.GitProvider(s), true
	}
	return "", false
}

func (h *IntegrationHandler) list(c *gin.Context) {
	accounts, err := h.svc.ListAccounts(c.Request.Context(), currentUserID(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, accounts)
}

func (h *IntegrationHandler) beginConnect(c *gin.Context) {
	provider, ok := parseProvider(c.Param("provider"))
	if !ok {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	authURL, err := h.svc.BeginConnect(c.Request.Context(), currentUserID(c), provider)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{"authUrl": authURL})
}

func (h *IntegrationHandler) callback(c *gin.Context) {
	provider, ok := parseProvider(c.Param("provider"))
	if !ok {
		h.redirectError(c, "unknown", "unknown provider")
		return
	}
	code := c.Query("code")
	state := c.Query("state")
	if errParam := c.Query("error"); errParam != "" {
		h.redirectError(c, string(provider), errParam)
		return
	}
	if code == "" || state == "" {
		h.redirectError(c, string(provider), "missing code or state")
		return
	}
	if _, err := h.svc.HandleCallback(c.Request.Context(), provider, code, state); err != nil {
		h.redirectError(c, string(provider), err.Error())
		return
	}
	h.redirectOK(c, string(provider))
}

func (h *IntegrationHandler) redirectOK(c *gin.Context, provider string) {
	q := url.Values{}
	q.Set("integration", provider)
	q.Set("status", "ok")
	c.Redirect(http.StatusFound, h.frontendURL+"/settings?"+q.Encode())
}

func (h *IntegrationHandler) redirectError(c *gin.Context, provider, msg string) {
	q := url.Values{}
	q.Set("integration", provider)
	q.Set("status", "error")
	q.Set("msg", msg)
	c.Redirect(http.StatusFound, h.frontendURL+"/settings?"+q.Encode())
}

func (h *IntegrationHandler) disconnect(c *gin.Context) {
	provider, ok := parseProvider(c.Param("provider"))
	if !ok {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err := h.svc.Disconnect(c.Request.Context(), currentUserID(c), provider); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}

func (h *IntegrationHandler) listRepos(c *gin.Context) {
	provider, ok := parseProvider(c.Param("provider"))
	if !ok {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	repos, err := h.svc.ListAvailableRepos(c.Request.Context(), currentUserID(c), provider)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, repos)
}

func (h *IntegrationHandler) setPins(c *gin.Context) {
	var body struct {
		Pins []service.PinSelection `json:"pins"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, domain.ErrInvalidInput)
		return
	}
	if err := h.svc.SetPins(c.Request.Context(), currentUserID(c), body.Pins); err != nil {
		response.Err(c, err)
		return
	}
	response.NoContent(c)
}
