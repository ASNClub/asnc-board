package worker

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog"
	"honeygarden/internal/domain"
	"honeygarden/internal/port"
)

const userAgent = "HoneyGarden/1.0 RSS Reader"

var ogImageRe = regexp.MustCompile(`(?i)<meta[^>]+property=["']og:image["'][^>]+content=["']([^"']+)["']`)
var ogImageReAlt = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:image["']`)
var imgTagRe = regexp.MustCompile(`(?i)<img[^>]+src=["']([^"']+)["']`)

type RSSWorker struct {
	sources    port.SourceRepository
	posts      port.PostRepository
	interval   time.Duration
	timeout    time.Duration
	httpClient *http.Client
	log        zerolog.Logger
}

func NewRSSWorker(
	sources port.SourceRepository,
	posts port.PostRepository,
	interval time.Duration,
	timeout time.Duration,
	log zerolog.Logger,
) *RSSWorker {
	return &RSSWorker{
		sources:    sources,
		posts:      posts,
		interval:   interval,
		timeout:    timeout,
		httpClient: NewSafeHTTPClient(timeout),
		log:        log,
	}
}

func (w *RSSWorker) Start(ctx context.Context) func() {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		w.run(ctx)
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.run(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
	return cancel
}

func (w *RSSWorker) run(ctx context.Context) {
	sources, err := w.sources.List(ctx)
	if err != nil {
		w.log.Error().Err(err).Msg("rss: failed to list sources")
		return
	}
	w.log.Info().Int("sources", len(sources)).Msg("rss: fetching feeds")
	for _, s := range sources {
		if ctx.Err() != nil {
			return
		}
		w.fetchSource(ctx, s)
	}
}

func (w *RSSWorker) fetchSource(ctx context.Context, s domain.RSSSource) {
	fetchCtx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	fp := gofeed.NewParser()
	fp.Client = w.httpClient

	feed, err := fp.ParseURLWithContext(s.URL, fetchCtx)
	if err != nil {
		w.log.Warn().Err(err).Str("source", s.Name).Msg("rss: failed to parse feed")
		return
	}

	saved := 0
	for _, item := range feed.Items {
		if ctx.Err() != nil {
			return
		}
		if item.Link == "" {
			continue
		}

		publishedAt := time.Now()
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			publishedAt = *item.UpdatedParsed
		}

		if time.Since(publishedAt) > 90*24*time.Hour {
			continue
		}

		coverImage := w.extractCoverImage(item)
		if coverImage == "" {
			coverImage = w.fetchOGImage(fetchCtx, item.Link)
		}

		ext := &domain.ExternalPost{
			ID:            uuid.New(),
			SourceID:      s.ID,
			Title:         item.Title,
			URL:           item.Link,
			Summary:       "",
			CoverImageURL: coverImage,
			Tags:          mergeTags(s.Tags, item.Categories),
			PublishedAt:   publishedAt,
		}

		if err = w.posts.UpsertExternal(fetchCtx, ext); err != nil {
			w.log.Error().Err(err).Str("url", item.Link).Msg("rss: failed to upsert external post")
			continue
		}
		saved++
	}

	_ = w.sources.UpdateLastFetched(ctx, s.ID, time.Now())
	w.log.Info().Str("source", s.Name).Int("saved", saved).Msg("rss: feed processed")
}

func (w *RSSWorker) extractCoverImage(item *gofeed.Item) string {
	if item.Image != nil && item.Image.URL != "" {
		return item.Image.URL
	}
	for _, enc := range item.Enclosures {
		if strings.HasPrefix(enc.Type, "image/") && enc.URL != "" {
			return enc.URL
		}
	}
	// media:thumbnail / media:content (Yahoo Media RSS)
	if media, ok := item.Extensions["media"]; ok {
		for _, key := range []string{"thumbnail", "content"} {
			for _, ext := range media[key] {
				if u := ext.Attrs["url"]; u != "" {
					return u
				}
			}
		}
	}
	if item.ITunesExt != nil && item.ITunesExt.Image != "" {
		return item.ITunesExt.Image
	}
	// первый <img> внутри content:encoded или description
	for _, html := range []string{item.Content, item.Description} {
		if html == "" {
			continue
		}
		if m := imgTagRe.FindStringSubmatch(html); len(m) > 1 {
			return m[1]
		}
	}
	return ""
}

func (w *RSSWorker) fetchOGImage(ctx context.Context, pageURL string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		return ""
	}

	buf, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	html := string(buf)

	if m := ogImageRe.FindStringSubmatch(html); len(m) > 1 {
		return m[1]
	}
	if m := ogImageReAlt.FindStringSubmatch(html); len(m) > 1 {
		return m[1]
	}
	return ""
}

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

func stripHTML(s string) string {
	return strings.TrimSpace(htmlTagRe.ReplaceAllString(s, ""))
}

func mergeTags(sourceTags, categories []string) []string {
	seen := make(map[string]struct{}, len(sourceTags)+len(categories))
	result := make([]string, 0, len(sourceTags)+len(categories))
	for _, t := range sourceTags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if _, ok := seen[t]; !ok {
			seen[t] = struct{}{}
			result = append(result, t)
		}
	}
	for _, c := range categories {
		c = strings.ToLower(strings.TrimSpace(c))
		if c == "" || len(c) > 32 {
			continue
		}
		if _, ok := seen[c]; !ok {
			seen[c] = struct{}{}
			result = append(result, c)
		}
	}
	return result
}
