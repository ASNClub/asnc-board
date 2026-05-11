package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
	"honeygarden/internal/adapter/http/response"
	"honeygarden/internal/config"
	"honeygarden/internal/domain"
	"honeygarden/internal/port"
	"honeygarden/internal/service"
)

const ctxUser = "current_user"

// hostOverrideTransport подменяет Host заголовок — нужно когда Zitadel
// доступен по внутреннему адресу, но ожидает внешний Host.
type hostOverrideTransport struct {
	base http.RoundTripper
	host string
}

func (t *hostOverrideTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r2 := req.Clone(req.Context())
	r2.Host = t.host
	return t.base.RoundTrip(r2)
}

func newJWKSStorage(ctx context.Context, rawURL, hostOverride string) (jwkset.Storage, error) {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid JWKS URL %q: %w", rawURL, err)
	}
	var httpClient *http.Client
	if hostOverride != "" {
		httpClient = &http.Client{
			Transport: &hostOverrideTransport{base: http.DefaultTransport, host: hostOverride},
		}
	}
	refreshErrHandler := func(ctx context.Context, err error) {
		slog.Default().ErrorContext(ctx, "JWKS refresh failed", "error", err, "url", rawURL)
	}
	perURL, err := jwkset.NewStorageFromHTTP(parsed, jwkset.HTTPClientStorageOptions{
		Client:                    httpClient,
		Ctx:                       ctx,
		NoErrorReturnFirstHTTPReq: true,
		RefreshErrorHandler:       refreshErrHandler,
		RefreshInterval:           time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("JWKS storage for %q: %w", rawURL, errors.Join(err, jwkset.ErrNewClient))
	}
	return jwkset.NewHTTPClient(jwkset.HTTPClientOptions{
		HTTPURLs:          map[string]jwkset.Storage{rawURL: perURL},
		RateLimitWaitMax:  time.Minute,
		RefreshUnknownKID: rate.NewLimiter(rate.Every(5*time.Minute), 1),
	})
}

// stripAuthHeaders удаляет identity-заголовки, которые могли прийти от клиента.
// Вызывается в самом начале JWTMiddleware до любой логики, чтобы исключить
// spoofing через X-Auth-*
func stripAuthHeaders(c *gin.Context) {
	c.Request.Header.Del("X-Auth-ID")
	c.Request.Header.Del("X-Auth-Username")
	c.Request.Header.Del("X-Auth-Display-Name")
	c.Request.Header.Del("X-Auth-Email")
}

func JWTMiddleware(cfg *config.Config) gin.HandlerFunc {
	if cfg.DevAuth {
		return func(c *gin.Context) {
			stripAuthHeaders(c)
			if id := c.GetHeader("X-Dev-Auth-ID"); id != "" {
				c.Request.Header.Set("X-Auth-ID", id)
			}
			c.Next()
		}
	}

	if cfg.JWKSUrl == "" {
		return func(c *gin.Context) {
			stripAuthHeaders(c)
			c.Next()
		}
	}

	ctx := context.Background()
	storage, err := newJWKSStorage(ctx, cfg.JWKSUrl, cfg.JWKSHost)
	if err != nil {
		log.Fatalf("failed to init JWKS: %v", err)
	}
	jwks, err := keyfunc.New(keyfunc.Options{Ctx: ctx, Storage: storage})
	if err != nil {
		log.Fatalf("failed to create keyfunc: %v", err)
	}

	return func(c *gin.Context) {
		stripAuthHeaders(c)
		bearer := c.GetHeader("Authorization")
		if bearer == "" {
			bearer = bearerFromWebSocketProtocol(c.GetHeader("Sec-WebSocket-Protocol"))
		}
		if bearer == "" || !strings.HasPrefix(bearer, "Bearer ") {
			c.Next()
			return
		}
		token, err := jwt.Parse(strings.TrimPrefix(bearer, "Bearer "), jwks.Keyfunc)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			c.Abort()
			return
		}
		if sub, _ := claims["sub"].(string); sub != "" {
			c.Request.Header.Set("X-Auth-ID", sub)
		}
		username, _ := claims["preferred_username"].(string)
		if username == "" {
			username, _ = claims["nickname"].(string)
		}
		if username != "" {
			c.Request.Header.Set("X-Auth-Username", username)
		}
		if name, _ := claims["name"].(string); name != "" {
			c.Request.Header.Set("X-Auth-Display-Name", name)
		}
		if email, _ := claims["email"].(string); email != "" {
			c.Request.Header.Set("X-Auth-Email", email)
		}
		c.Next()
	}
}

func bearerFromWebSocketProtocol(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.Split(header, ",")
	for i := 0; i < len(parts)-1; i++ {
		if strings.TrimSpace(parts[i]) == "bearer" {
			if token := strings.TrimSpace(parts[i+1]); token != "" {
				return "Bearer " + token
			}
		}
	}
	return ""
}

func Logger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Msg("request")
	}
}

func RequireAuth(svc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := provisionUser(c, svc)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": domain.ErrForbidden.Error()})
			c.Abort()
			return
		}
		if user.BannedAt != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "account banned"})
			c.Abort()
			return
		}
		c.Set(ctxUser, user)
		c.Next()
	}
}

// RequireAdmin проверяет, что currentUser входит в список admin auth-id.
// Должно применяться ПОСЛЕ ПОСЛЕ ПОСЛЕ RequireAuth.
func RequireAdmin(adminAuthIDs []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(adminAuthIDs))
	for _, id := range adminAuthIDs {
		allowed[id] = struct{}{}
	}
	return func(c *gin.Context) {
		u := currentUser(c)
		if u == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": domain.ErrForbidden.Error()})
			c.Abort()
			return
		}
		if _, ok := allowed[u.AuthID]; !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": domain.ErrForbidden.Error()})
			c.Abort()
			return
		}
		c.Next()
	}
}

func OptionalAuth(svc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, ok := provisionUser(c, svc); ok {
			c.Set(ctxUser, user)
		}
		c.Next()
	}
}

func provisionUser(c *gin.Context, svc *service.UserService) (*domain.User, bool) {
	authID := c.GetHeader("X-Auth-ID")
	if authID == "" {
		return nil, false
	}
	username := c.GetHeader("X-Auth-Username")
	if username == "" {
		username = authID
	}
	displayName := c.GetHeader("X-Auth-Display-Name")
	if displayName == "" {
		displayName = username
	}
	user, err := svc.GetOrProvision(c.Request.Context(), authID, username, displayName)
	if err != nil {
		response.Err(c, err)
		return nil, false
	}
	return user, true
}

func currentUser(c *gin.Context) *domain.User {
	v, _ := c.Get(ctxUser)
	u, _ := v.(*domain.User)
	return u
}

func currentUserID(c *gin.Context) uuid.UUID {
	if u := currentUser(c); u != nil {
		return u.ID
	}
	return uuid.UUID{}
}

func currentUserIDOpt(c *gin.Context) *uuid.UUID {
	if u := currentUser(c); u != nil {
		id := u.ID
		return &id
	}
	return nil
}

func LastSeenMiddleware(repo port.UserRepository) gin.HandlerFunc {
	var mu sync.Map // map[uuid.UUID]time.Time

	return func(c *gin.Context) {
		c.Next()

		u := currentUser(c)
		if u == nil {
			return
		}
		now := time.Now()
		if v, ok := mu.Load(u.ID); ok {
			if now.Sub(v.(time.Time)) < time.Minute {
				return
			}
		}
		mu.Store(u.ID, now)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = repo.TouchLastSeen(ctx, u.ID)
		}()
	}
}
