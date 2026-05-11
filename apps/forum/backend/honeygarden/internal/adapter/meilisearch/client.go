package meilisearch

import (
	"context"
	"fmt"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"honeygarden/internal/domain"
)

const (
	indexPosts       = "posts"
	indexCommunities = "communities"
)

type Client struct {
	ms meilisearch.ServiceManager
}

func NewClient(host, apiKey string) (*Client, error) {
	ms := meilisearch.New(host, meilisearch.WithAPIKey(apiKey))
	if _, err := ms.Health(); err != nil {
		return nil, fmt.Errorf("meilisearch health check failed: %w", err)
	}
	c := &Client{ms: ms}
	if err := c.ensureIndexes(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) ensureIndexes() error {
	if _, err := c.ms.GetIndex(indexPosts); err != nil {
		task, err := c.ms.CreateIndex(&meilisearch.IndexConfig{Uid: indexPosts, PrimaryKey: "id"})
		if err != nil {
			return err
		}
		c.ms.WaitForTask(task.TaskUID, 500*time.Millisecond)
	}
	c.ms.Index(indexPosts).UpdateSearchableAttributes(&[]string{"title", "content"})
	c.ms.Index(indexPosts).UpdateSortableAttributes(&[]string{"created_at"})

	if _, err := c.ms.GetIndex(indexCommunities); err != nil {
		task, err := c.ms.CreateIndex(&meilisearch.IndexConfig{Uid: indexCommunities, PrimaryKey: "id"})
		if err != nil {
			return err
		}
		c.ms.WaitForTask(task.TaskUID, 500*time.Millisecond)
	}
	c.ms.Index(indexCommunities).UpdateSearchableAttributes(&[]string{"name", "description", "tags"})
	c.ms.Index(indexCommunities).UpdateSortableAttributes(&[]string{"followers_count"})

	return nil
}

func (c *Client) IndexPost(_ context.Context, doc domain.PostDocument) error {
	_, err := c.ms.Index(indexPosts).AddDocuments([]domain.PostDocument{doc}, "id")
	return err
}

func (c *Client) IndexCommunity(_ context.Context, doc domain.CommunityDocument) error {
	_, err := c.ms.Index(indexCommunities).AddDocuments([]domain.CommunityDocument{doc}, "id")
	return err
}

func (c *Client) DeletePost(_ context.Context, id string) error {
	_, err := c.ms.Index(indexPosts).DeleteDocument(id)
	return err
}

func (c *Client) DeleteCommunity(_ context.Context, id string) error {
	_, err := c.ms.Index(indexCommunities).DeleteDocument(id)
	return err
}

func (c *Client) SearchPosts(_ context.Context, q string, limit, offset int) (*domain.SearchResult[domain.PostDocument], error) {
	res, err := c.ms.Index(indexPosts).Search(q, &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, err
	}
	var hits []domain.PostDocument
	for _, h := range res.Hits {
		m, _ := h.(map[string]any)
		doc := domain.PostDocument{
			ID:          str(m, "id"),
			CommunityID: str(m, "community_id"),
			AuthorID:    str(m, "author_id"),
			Content:     str(m, "content"),
		}
		if v, ok := m["title"].(string); ok {
			doc.Title = &v
		}
		hits = append(hits, doc)
	}
	if hits == nil {
		hits = []domain.PostDocument{}
	}
	return &domain.SearchResult[domain.PostDocument]{Hits: hits, Total: int(res.EstimatedTotalHits)}, nil
}

func (c *Client) SearchCommunities(_ context.Context, q string, limit, offset int) (*domain.SearchResult[domain.CommunityDocument], error) {
	res, err := c.ms.Index(indexCommunities).Search(q, &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, err
	}
	var hits []domain.CommunityDocument
	for _, h := range res.Hits {
		m, _ := h.(map[string]any)
		doc := domain.CommunityDocument{
			ID:   str(m, "id"),
			Slug: str(m, "slug"),
			Name: str(m, "name"),
		}
		if v, ok := m["description"].(string); ok {
			doc.Description = &v
		}
		hits = append(hits, doc)
	}
	if hits == nil {
		hits = []domain.CommunityDocument{}
	}
	return &domain.SearchResult[domain.CommunityDocument]{Hits: hits, Total: int(res.EstimatedTotalHits)}, nil
}

func str(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}
