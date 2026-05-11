package wakapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"honeygarden/internal/domain"
)

type Client struct {
	http *http.Client
}

func New() *Client {
	return &Client{http: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) FetchStats(ctx context.Context, instanceURL, apiKey, username string) (domain.WakapiStats, error) {
	base := strings.TrimRight(instanceURL, "/")
	base = strings.TrimSuffix(base, "/api")
	url := fmt.Sprintf("%s/api/compat/wakatime/v1/users/%s/stats/last_7_days", base, username)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.WakapiStats{}, err
	}
	req.Header.Set("Authorization", "Basic "+basicAuth(apiKey))

	resp, err := c.http.Do(req)
	if err != nil {
		return domain.WakapiStats{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.WakapiStats{}, fmt.Errorf("wakapi: status %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			TotalSeconds      float64 `json:"total_seconds"`
			DailyAverage      float64 `json:"daily_average"`
			Languages         []struct {
				Name    string  `json:"name"`
				Percent float64 `json:"percent"`
			} `json:"languages"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return domain.WakapiStats{}, err
	}

	stats := domain.WakapiStats{
		TotalSeconds: body.Data.TotalSeconds,
		DailyAverage: body.Data.DailyAverage,
	}

	for _, l := range body.Data.Languages {
		stats.Languages = append(stats.Languages, domain.WakapiLang{
			Name:    l.Name,
			Percent: l.Percent,
			Color:   langColor(l.Name),
		})
	}

	return stats, nil
}

func basicAuth(apiKey string) string {
	return base64.StdEncoding.EncodeToString([]byte(apiKey + ":"))
}

var langColors = map[string]string{
	"Go":         "#00ADD8",
	"TypeScript": "#3178C6",
	"JavaScript": "#F7DF1E",
	"Python":     "#3572A5",
	"Rust":       "#DEA584",
	"Java":       "#B07219",
	"C":          "#555555",
	"C++":        "#F34B7D",
	"C#":         "#178600",
	"Ruby":       "#701516",
	"PHP":        "#4F5D95",
	"Swift":      "#F05138",
	"Kotlin":     "#A97BFF",
	"Dart":       "#00B4AB",
	"HTML":       "#E34C26",
	"CSS":        "#563D7C",
	"SCSS":       "#C6538C",
	"Shell":      "#89E051",
	"Bash":       "#89E051",
	"SQL":        "#E38C00",
	"Lua":        "#000080",
	"Vue":        "#41B883",
	"Svelte":     "#FF3E00",
	"YAML":       "#CB171E",
	"JSON":       "#292929",
	"Markdown":   "#083FA1",
	"Docker":     "#384D54",
	"Zig":        "#EC915C",
}

func langColor(name string) string {
	if c, ok := langColors[name]; ok {
		return c
	}
	return "#6E7681"
}
