package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port string
	DevAuth      bool
	JWKSUrl      string
	JWKSHost     string
	AdminAuthIDs []string
	DatabaseURL string
	NatsURL string
	MeilisearchURL    string
	MeilisearchAPIKey string
	S3Endpoint   string
	S3Bucket     string
	S3AccessKey  string
	S3SecretKey  string
	S3UseSSL     bool
	S3Region     string
	S3PublicBase string // базовый URL, под которым клиенту видно бакет (например /uploads)
	RSSInterval time.Duration
	RSSOTimeout time.Duration
	FrontendURL    string // куда редиректить юзера после OAuth callback
	OAuthCallback  string // base URL для redirect_uri (например https://honeygarden.space/api)
	OAuthStateKey  string // секрет для подписи state-токена
	GitHubClientID     string
	GitHubClientSecret string
	GitLabClientID     string
	GitLabClientSecret string
	GitLabBaseURL      string // https://gitlab.com по умолчанию
	CodebergClientID     string
	CodebergClientSecret string
}

func Load() *Config {
	return &Config{
		Port: env("PORT", "8080"),

		DevAuth:      env("DEV_AUTH", "false") == "true",
		JWKSUrl:      env("JWKS_URL", ""),
		JWKSHost:     env("JWKS_HOST", ""),
		AdminAuthIDs: parseList(env("ADMIN_AUTH_IDS", "")),

		DatabaseURL: env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/honeygarden"),

		NatsURL: env("NATS_URL", "nats://localhost:4222"),

		MeilisearchURL:    env("MEILISEARCH_URL", "http://localhost:7700"),
		MeilisearchAPIKey: env("MEILISEARCH_API_KEY", ""),

		S3Endpoint:   env("S3_ENDPOINT", "minio:9000"),
		S3Bucket:     env("S3_BUCKET", "uploads"),
		S3AccessKey:  env("S3_ACCESS_KEY", ""),
		S3SecretKey:  env("S3_SECRET_KEY", ""),
		S3UseSSL:     env("S3_USE_SSL", "false") == "true",
		S3Region:     env("S3_REGION", "us-east-1"),
		S3PublicBase: env("S3_PUBLIC_BASE", "/uploads"),

		RSSInterval: envDuration("RSS_INTERVAL", 30*time.Minute),
		RSSOTimeout: envDuration("RSS_TIMEOUT", 15*time.Second),

		FrontendURL:   env("FRONTEND_URL", "https://honeygarden.space"),
		OAuthCallback: env("OAUTH_CALLBACK_BASE", "https://honeygarden.space/api/v1"),
		OAuthStateKey: env("OAUTH_STATE_KEY", ""),

		GitHubClientID:     env("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: env("GITHUB_CLIENT_SECRET", ""),

		GitLabClientID:     env("GITLAB_CLIENT_ID", ""),
		GitLabClientSecret: env("GITLAB_CLIENT_SECRET", ""),
		GitLabBaseURL:      env("GITLAB_BASE_URL", "https://gitlab.com"),

		CodebergClientID:     env("CODEBERG_CLIENT_ID", ""),
		CodebergClientSecret: env("CODEBERG_CLIENT_SECRET", ""),
	}
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func parseList(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Duration(n) * time.Second
		}
	}
	return fallback
}
