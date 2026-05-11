package oauthclients

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"honeygarden/internal/port"
)

// Codeberg — Forgejo OAuth2 (тот же протокол, что у Gitea).
// По умолчанию — codeberg.org; instance_url не используем (одно публичное API).
type Codeberg struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

const (
	codebergBaseURL  = "https://codeberg.org"
	codebergAuthPath = "/login/oauth/authorize"
	codebergTokPath  = "/login/oauth/access_token"
	codebergAPIv1    = "https://codeberg.org/api/v1"
)

func (c *Codeberg) AuthURL(state string) string {
	q := url.Values{}
	q.Set("client_id", c.ClientID)
	q.Set("redirect_uri", c.RedirectURL)
	q.Set("response_type", "code")
	q.Set("state", state)
	return addQuery(codebergBaseURL+codebergAuthPath, q)
}

func (c *Codeberg) Exchange(ctx context.Context, code string) (*port.OAuthToken, error) {
	form := url.Values{}
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.RedirectURL)
	tok, err := postForm(ctx, codebergBaseURL+codebergTokPath, form)
	if err != nil {
		return nil, err
	}
	return tok.toToken(), nil
}

func (c *Codeberg) Refresh(ctx context.Context, refreshToken string) (*port.OAuthToken, error) {
	if refreshToken == "" {
		return nil, errors.New("codeberg: refresh_token not set")
	}
	form := url.Values{}
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	tok, err := postForm(ctx, codebergBaseURL+codebergTokPath, form)
	if err != nil {
		return nil, err
	}
	return tok.toToken(), nil
}

type codebergUser struct {
	Login string `json:"login"`
}

func (c *Codeberg) GetUsername(ctx context.Context, token string, _ *string) (string, error) {
	var u codebergUser
	if err := apiGet(ctx, codebergAPIv1+"/user", token, &u); err != nil {
		return "", err
	}
	return u.Login, nil
}

type codebergRepo struct {
	ID          int64    `json:"id"`
	FullName    string   `json:"full_name"`
	Description string   `json:"description"`
	HTMLURL     string   `json:"html_url"`
	Language    string   `json:"language"`
	Stars       int      `json:"stars_count"`
	Forks       int      `json:"forks_count"`
	Fork        bool     `json:"fork"`
	Private     bool     `json:"private"`
	Topics      []string `json:"topics"`
}

func (c *Codeberg) FetchRepos(ctx context.Context, token string, _ *string) ([]port.RepoData, error) {
	out := []port.RepoData{}
	for page := 1; page <= 5; page++ {
		q := url.Values{}
		q.Set("limit", "50")
		q.Set("page", strconv.Itoa(page))
		var batch []codebergRepo
		if err := apiGet(ctx, addQuery(codebergAPIv1+"/user/repos", q), token, &batch); err != nil {
			return nil, err
		}
		for _, r := range batch {
			if r.Private {
				continue
			}
			var desc *string
			if d := strings.TrimSpace(r.Description); d != "" {
				desc = &d
			}
			var lang *string
			if l := strings.TrimSpace(r.Language); l != "" {
				lang = &l
			}
			topics := r.Topics
			if topics == nil {
				topics = []string{}
			}
			out = append(out, port.RepoData{
				ExternalID:  strconv.FormatInt(r.ID, 10),
				Name:        r.FullName,
				Description: desc,
				URL:         r.HTMLURL,
				Language:    lang,
				StarsCount:  r.Stars,
				ForksCount:  r.Forks,
				IsFork:      r.Fork,
				Topics:      topics,
			})
		}
		if len(batch) < 50 {
			break
		}
	}
	return out, nil
}
