DROP INDEX IF EXISTS idx_posts_community_kind;
ALTER TABLE posts DROP COLUMN IF EXISTS kind;
