// Package oauthclients реализует OAuth2-клиенты для GitHub / GitLab / Codeberg.
//
// Каждый клиент реализует одновременно port.OAuthProvider (для flow подключения)
// и port.GitProviderClient (для подтягивания списка репов).
package oauthclients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"honeygarden/internal/port"
)

const httpTimeout = 15 * time.Second

func httpClient() *http.Client {
	return &http.Client{Timeout: httpTimeout}
}

// tokenResponse — общая форма ответа на code-exchange и refresh у GitLab/Codeberg.
// GitHub возвращает то же самое при Accept: application/json.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // секунды; 0 = бессрочно
	Scope        string `json:"scope"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

func (t *tokenResponse) toToken() *port.OAuthToken {
	tok := &port.OAuthToken{AccessToken: t.AccessToken}
	if t.RefreshToken != "" {
		tok.RefreshToken = t.RefreshToken
	}
	if t.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
		tok.ExpiresAt = &exp
	}
	return tok
}

func postForm(ctx context.Context, endpoint string, form url.Values) (*tokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, fmt.Errorf("oauth: parse token response: %w", err)
	}
	if tok.Error != "" {
		return nil, fmt.Errorf("oauth: %s: %s", tok.Error, tok.ErrorDesc)
	}
	if tok.AccessToken == "" {
		return nil, errors.New("oauth: empty access_token in response")
	}
	return &tok, nil
}

func apiGet(ctx context.Context, endpoint, token string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api %s: %d %s", endpoint, resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

func addQuery(rawURL string, params url.Values) string {
	if !strings.Contains(rawURL, "?") {
		return rawURL + "?" + params.Encode()
	}
	return rawURL + "&" + params.Encode()
}
