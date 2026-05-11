package oauthclients

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"honeygarden/internal/port"
)

type GitHub struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

const (
	githubAuthURL  = "https://github.com/login/oauth/authorize"
	githubTokenURL = "https://github.com/login/oauth/access_token"
	githubAPIURL   = "https://api.github.com"
	// scopes:
	//   read:user — читать профиль (для GetUsername)
	//   public_repo — читать публичные репы
	//   repo (опционально) — читать приватные; пока берём только public_repo
	githubScope = "read:user public_repo"
)

func (g *GitHub) AuthURL(state string) string {
	q := url.Values{}
	q.Set("client_id", g.ClientID)
	q.Set("redirect_uri", g.RedirectURL)
	q.Set("scope", githubScope)
	q.Set("state", state)
	q.Set("allow_signup", "false")
	return addQuery(githubAuthURL, q)
}

func (g *GitHub) Exchange(ctx context.Context, code string) (*port.OAuthToken, error) {
	form := url.Values{}
	form.Set("client_id", g.ClientID)
	form.Set("client_secret", g.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", g.RedirectURL)
	tok, err := postForm(ctx, githubTokenURL, form)
	if err != nil {
		return nil, err
	}
	return tok.toToken(), nil
}

func (g *GitHub) Refresh(ctx context.Context, refreshToken string) (*port.OAuthToken, error) {
	if refreshToken == "" {
		return nil, errors.New("github: refresh_token not set")
	}
	form := url.Values{}
	form.Set("client_id", g.ClientID)
	form.Set("client_secret", g.ClientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	tok, err := postForm(ctx, githubTokenURL, form)
	if err != nil {
		return nil, err
	}
	return tok.toToken(), nil
}

type githubUser struct {
	Login string `json:"login"`
}

func (g *GitHub) GetUsername(ctx context.Context, token string, _ *string) (string, error) {
	var u githubUser
	if err := apiGet(ctx, githubAPIURL+"/user", token, &u); err != nil {
		return "", err
	}
	return u.Login, nil
}

type githubRepo struct {
	ID          int64    `json:"id"`
	FullName    string   `json:"full_name"`
	Description *string  `json:"description"`
	HTMLURL     string   `json:"html_url"`
	Language    *string  `json:"language"`
	Stars       int      `json:"stargazers_count"`
	Forks       int      `json:"forks_count"`
	Fork        bool     `json:"fork"`
	Topics      []string `json:"topics"`
	Private     bool     `json:"private"`
}

func (g *GitHub) FetchRepos(ctx context.Context, token string, _ *string) ([]port.RepoData, error) {
	out := []port.RepoData{}
	for page := 1; page <= 5; page++ { // максимум 500 репов
		q := url.Values{}
		q.Set("per_page", "100")
		q.Set("page", strconv.Itoa(page))
		q.Set("sort", "updated")
		q.Set("affiliation", "owner,collaborator,organization_member")
		var batch []githubRepo
		if err := apiGet(ctx, addQuery(githubAPIURL+"/user/repos", q), token, &batch); err != nil {
			return nil, err
		}
		for _, r := range batch {
			if r.Private {
				continue // на профиле показываем только публичные
			}
			topics := r.Topics
			if topics == nil {
				topics = []string{}
			}
			out = append(out, port.RepoData{
				ExternalID:  strconv.FormatInt(r.ID, 10),
				Name:        r.FullName,
				Description: r.Description,
				URL:         r.HTMLURL,
				Language:    r.Language,
				StarsCount:  r.Stars,
				ForksCount:  r.Forks,
				IsFork:      r.Fork,
				Topics:      topics,
			})
		}
		if len(batch) < 100 {
			break
		}
	}
	return out, nil
}
