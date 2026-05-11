DROP INDEX IF EXISTS idx_posts_short_id;
ALTER TABLE posts DROP COLUMN IF EXISTS short_id;
