-- Post kind: discussion / article / question.
-- External (RSS) posts остаются с kind='discussion' (default), для них kind не релевантен.

ALTER TABLE posts ADD COLUMN kind TEXT NOT NULL DEFAULT 'discussion'
    CHECK (kind IN ('discussion', 'article', 'question'));

CREATE INDEX idx_posts_community_kind ON posts(community_id, kind, created_at DESC)
    WHERE community_id IS NOT NULL;
