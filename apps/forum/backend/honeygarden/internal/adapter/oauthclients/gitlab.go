package oauthclients

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"honeygarden/internal/port"
)

// GitLab реализует port.OAuthProvider + port.GitProviderClient.
// BaseURL — корень инстанса GitLab (например https://gitlab.com).
type GitLab struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	BaseURL      string // без слеша на конце
}

const gitlabScope = "read_api read_user"

func (g *GitLab) base() string {
	if g.BaseURL == "" {
		return "https://gitlab.com"
	}
	return strings.TrimSuffix(g.BaseURL, "/")
}

func (g *GitLab) AuthURL(state string) string {
	q := url.Values{}
	q.Set("client_id", g.ClientID)
	q.Set("redirect_uri", g.RedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", gitlabScope)
	q.Set("state", state)
	return addQuery(g.base()+"/oauth/authorize", q)
}

func (g *GitLab) Exchange(ctx context.Context, code string) (*port.OAuthToken, error) {
	form := url.Values{}
	form.Set("client_id", g.ClientID)
	form.Set("client_secret", g.ClientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", g.RedirectURL)
	tok, err := postForm(ctx, g.base()+"/oauth/token", form)
	if err != nil {
		return nil, err
	}
	return tok.toToken(), nil
}

func (g *GitLab) Refresh(ctx context.Context, refreshToken string) (*port.OAuthToken, error) {
	if refreshToken == "" {
		return nil, errors.New("gitlab: refresh_token not set")
	}
	form := url.Values{}
	form.Set("client_id", g.ClientID)
	form.Set("client_secret", g.ClientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	tok, err := postForm(ctx, g.base()+"/oauth/token", form)
	if err != nil {
		return nil, err
	}
	return tok.toToken(), nil
}

type gitlabUser struct {
	Username string `json:"username"`
}

func (g *GitLab) baseFor(instance *string) string {
	if instance != nil && *instance != "" {
		return strings.TrimSuffix(*instance, "/")
	}
	return g.base()
}

func (g *GitLab) GetUsername(ctx context.Context, token string, instance *string) (string, error) {
	var u gitlabUser
	if err := apiGet(ctx, g.baseFor(instance)+"/api/v4/user", token, &u); err != nil {
		return "", err
	}
	return u.Username, nil
}

type gitlabProject struct {
	ID                int64    `json:"id"`
	PathWithNamespace string   `json:"path_with_namespace"`
	Description       *string  `json:"description"`
	WebURL            string   `json:"web_url"`
	StarsCount        int      `json:"star_count"`
	ForksCount        int      `json:"forks_count"`
	Topics            []string `json:"topics"`
	Visibility        string   `json:"visibility"`
	ForkedFromProject *struct {
		ID int64 `json:"id"`
	} `json:"forked_from_project"`
}

func (g *GitLab) FetchRepos(ctx context.Context, token string, instance *string) ([]port.RepoData, error) {
	base := g.baseFor(instance)
	out := []port.RepoData{}
	for page := 1; page <= 5; page++ {
		q := url.Values{}
		q.Set("membership", "true")
		q.Set("per_page", "100")
		q.Set("page", strconv.Itoa(page))
		q.Set("order_by", "updated_at")
		var batch []gitlabProject
		if err := apiGet(ctx, addQuery(base+"/api/v4/projects", q), token, &batch); err != nil {
			return nil, err
		}
		for _, p := range batch {
			if p.Visibility == "private" {
				continue
			}
			topics := p.Topics
			if topics == nil {
				topics = []string{}
			}
			isFork := p.ForkedFromProject != nil
			out = append(out, port.RepoData{
				ExternalID:  strconv.FormatInt(p.ID, 10),
				Name:        p.PathWithNamespace,
				Description: p.Description,
				URL:         p.WebURL,
				StarsCount:  p.StarsCount,
				ForksCount:  p.ForksCount,
				IsFork:      isFork,
				Topics:      topics,
			})
		}
		if len(batch) < 100 {
			break
		}
	}
	return out, nil
}
