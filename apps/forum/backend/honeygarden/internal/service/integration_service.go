package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"honeygarden/internal/domain"
	"honeygarden/internal/port"
)

// IntegrationService управляет подключением GitHub/GitLab/Codeberg-аккаунтов
// и выбором закреплённых репозиториев.
type IntegrationService struct {
	accounts    port.GitAccountRepository
	pinnedRepos port.PinnedRepoRepository
	oauths      map[domain.GitProvider]port.OAuthProvider
	apis        map[domain.GitProvider]port.GitProviderClient
	stateSecret []byte
	log         zerolog.Logger
}

func NewIntegrationService(
	accounts port.GitAccountRepository,
	pinnedRepos port.PinnedRepoRepository,
	oauths map[domain.GitProvider]port.OAuthProvider,
	apis map[domain.GitProvider]port.GitProviderClient,
	stateSecret []byte,
	log zerolog.Logger,
) *IntegrationService {
	return &IntegrationService{
		accounts:    accounts,
		pinnedRepos: pinnedRepos,
		oauths:      oauths,
		apis:        apis,
		stateSecret: stateSecret,
		log:         log,
	}
}

// ── State ────────────────────────────────────────────────────────────────────
//
// State кодирует userID и nonce, подписан HMAC-SHA256 от stateSecret.
// Формат: base64url(json{u,n,e}).hex(hmac).
// Нужен чтобы:
// - защититься от CSRF при возврате с провайдера,
// - узнать какой юзер инициировал flow (callback приходит без auth-cookie/JWT).

type stateClaims struct {
	UserID   string `json:"u"`
	Nonce    string `json:"n"`
	ExpiryTS int64  `json:"e"`
	Provider string `json:"p"`
}

const stateTTL = 10 * time.Minute

func (s *IntegrationService) signState(userID uuid.UUID, provider domain.GitProvider) (string, error) {
	nonceBytes := make([]byte, 12)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}
	claims := stateClaims{
		UserID:   userID.String(),
		Nonce:    hex.EncodeToString(nonceBytes),
		ExpiryTS: time.Now().Add(stateTTL).Unix(),
		Provider: string(provider),
	}
	body, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, s.stateSecret)
	mac.Write([]byte(encoded))
	return encoded + "." + hex.EncodeToString(mac.Sum(nil)), nil
}

func (s *IntegrationService) verifyState(state string, provider domain.GitProvider) (uuid.UUID, error) {
	parts := strings.SplitN(state, ".", 2)
	if len(parts) != 2 {
		return uuid.Nil, errors.New("integration: bad state format")
	}
	encoded, sig := parts[0], parts[1]
	mac := hmac.New(sha256.New, s.stateSecret)
	mac.Write([]byte(encoded))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return uuid.Nil, errors.New("integration: state signature mismatch")
	}
	body, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return uuid.Nil, err
	}
	var claims stateClaims
	if err := json.Unmarshal(body, &claims); err != nil {
		return uuid.Nil, err
	}
	if claims.Provider != string(provider) {
		return uuid.Nil, errors.New("integration: provider mismatch")
	}
	if time.Now().Unix() > claims.ExpiryTS {
		return uuid.Nil, errors.New("integration: state expired")
	}
	return uuid.Parse(claims.UserID)
}

// ── Public API ───────────────────────────────────────────────────────────────

// BeginConnect формирует URL OAuth-провайдера, на который нужно отправить юзера.
func (s *IntegrationService) BeginConnect(ctx context.Context, userID uuid.UUID, provider domain.GitProvider) (string, error) {
	oauth, ok := s.oauths[provider]
	if !ok {
		return "", fmt.Errorf("integration: provider %q not configured", provider)
	}
	state, err := s.signState(userID, provider)
	if err != nil {
		return "", err
	}
	return oauth.AuthURL(state), nil
}

// HandleCallback меняет code на токен, читает username, апсёртит аккаунт.
// Возвращает свежесозданный/обновлённый GitAccount.
func (s *IntegrationService) HandleCallback(ctx context.Context, provider domain.GitProvider, code, state string) (*domain.GitAccount, error) {
	userID, err := s.verifyState(state, provider)
	if err != nil {
		return nil, err
	}
	oauth, ok := s.oauths[provider]
	if !ok {
		return nil, fmt.Errorf("integration: provider %q not configured", provider)
	}
	api, ok := s.apis[provider]
	if !ok {
		return nil, fmt.Errorf("integration: api client %q not configured", provider)
	}
	tok, err := oauth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("integration: exchange code: %w", err)
	}
	username, err := api.GetUsername(ctx, tok.AccessToken, nil)
	if err != nil {
		return nil, fmt.Errorf("integration: fetch username: %w", err)
	}
	var refresh *string
	if tok.RefreshToken != "" {
		rt := tok.RefreshToken
		refresh = &rt
	}
	account := &domain.GitAccount{
		ID:           uuid.New(),
		UserID:       userID,
		Provider:     provider,
		AccessToken:  tok.AccessToken,
		RefreshToken: refresh,
		ExpiresAt:    tok.ExpiresAt,
		Username:     username,
	}
	if err := s.accounts.Upsert(ctx, account); err != nil {
		return nil, err
	}
	got, err := s.accounts.GetByUserProvider(ctx, userID, provider, nil)
	if err != nil {
		return nil, err
	}
	return got, nil
}

// ListAccounts — все подключённые аккаунты юзера.
func (s *IntegrationService) ListAccounts(ctx context.Context, userID uuid.UUID) ([]domain.GitAccount, error) {
	return s.accounts.GetByUser(ctx, userID)
}

// Disconnect — отключить провайдера. Каскад удалит и пины.
func (s *IntegrationService) Disconnect(ctx context.Context, userID uuid.UUID, provider domain.GitProvider) error {
	a, err := s.accounts.GetByUserProvider(ctx, userID, provider, nil)
	if err != nil {
		return err
	}
	return s.accounts.Delete(ctx, a.ID)
}

func (s *IntegrationService) ListAvailableRepos(ctx context.Context, userID uuid.UUID, provider domain.GitProvider) ([]port.RepoData, error) {
	a, err := s.accounts.GetByUserProvider(ctx, userID, provider, nil)
	if err != nil {
		return nil, err
	}
	api, ok := s.apis[provider]
	if !ok {
		return nil, fmt.Errorf("integration: api client %q not configured", provider)
	}

	if err := s.refreshIfExpired(ctx, a); err != nil {
		return nil, err
	}
	return api.FetchRepos(ctx, a.AccessToken, a.InstanceURL)
}

type PinSelection struct {
	Provider   domain.GitProvider `json:"provider"`
	ExternalID string             `json:"externalId"`
}

func (s *IntegrationService) SetPins(ctx context.Context, userID uuid.UUID, selections []PinSelection) error {
	if len(selections) > 50 {
		return domain.ErrInvalidInput
	}
	if err := s.pinnedRepos.DeleteByUser(ctx, userID); err != nil {
		return err
	}
	if len(selections) == 0 {
		return nil
	}

	byProvider := make(map[domain.GitProvider][]string)
	for _, sel := range selections {
		byProvider[sel.Provider] = append(byProvider[sel.Provider], sel.ExternalID)
	}

	type cached struct {
		data      port.RepoData
		accountID uuid.UUID
	}
	cache := make(map[domain.GitProvider]map[string]cached, len(byProvider))

	for prov := range byProvider {
		a, err := s.accounts.GetByUserProvider(ctx, userID, prov, nil)
		if err != nil {
			return fmt.Errorf("integration: %s not connected: %w", prov, err)
		}
		if err := s.refreshIfExpired(ctx, a); err != nil {
			return err
		}
		api, ok := s.apis[prov]
		if !ok {
			return fmt.Errorf("integration: api client %q not configured", prov)
		}
		repos, err := api.FetchRepos(ctx, a.AccessToken, a.InstanceURL)
		if err != nil {
			return fmt.Errorf("integration: fetch repos for %s: %w", prov, err)
		}
		idx := make(map[string]cached, len(repos))
		for _, r := range repos {
			idx[r.ExternalID] = cached{data: r, accountID: a.ID}
		}
		cache[prov] = idx
	}

	for i, sel := range selections {
		idx := cache[sel.Provider]
		entry, ok := idx[sel.ExternalID]
		if !ok {
			s.log.Warn().
				Str("provider", string(sel.Provider)).
				Str("external_id", sel.ExternalID).
				Msg("pin selection not in fetched repos, skipping")
			continue
		}
		pin := &domain.PinnedRepo{
			ID:           uuid.New(),
			UserID:       userID,
			GitAccountID: entry.accountID,
			ExternalID:   entry.data.ExternalID,
			Name:         entry.data.Name,
			Description:  entry.data.Description,
			URL:          entry.data.URL,
			Language:     entry.data.Language,
			StarsCount:   entry.data.StarsCount,
			ForksCount:   entry.data.ForksCount,
			IsFork:       entry.data.IsFork,
			Topics:       entry.data.Topics,
			SortOrder:    i,
		}
		if err := s.pinnedRepos.Upsert(ctx, pin); err != nil {
			return err
		}
	}
	return nil
}

func (s *IntegrationService) GetPins(ctx context.Context, userID uuid.UUID) ([]domain.PinnedRepo, error) {
	return s.pinnedRepos.GetByUser(ctx, userID)
}

func (s *IntegrationService) refreshIfExpired(ctx context.Context, a *domain.GitAccount) error {
	if a.ExpiresAt == nil {
		return nil
	}
	if time.Now().Before(a.ExpiresAt.Add(-time.Minute)) {
		return nil
	}
	if a.RefreshToken == nil || *a.RefreshToken == "" {
		return nil
	}
	oauth, ok := s.oauths[a.Provider]
	if !ok {
		return nil
	}
	tok, err := oauth.Refresh(ctx, *a.RefreshToken)
	if err != nil {
		return fmt.Errorf("integration: refresh token: %w", err)
	}
	var refresh *string
	if tok.RefreshToken != "" {
		rt := tok.RefreshToken
		refresh = &rt
	} else {
		refresh = a.RefreshToken
	}
	if err := s.accounts.UpdateTokens(ctx, a.ID, tok.AccessToken, refresh, tok.ExpiresAt); err != nil {
		return err
	}
	a.AccessToken = tok.AccessToken
	a.RefreshToken = refresh
	a.ExpiresAt = tok.ExpiresAt
	return nil
}
