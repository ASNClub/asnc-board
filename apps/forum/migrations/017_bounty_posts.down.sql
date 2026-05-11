-- Remove indexes
DROP INDEX IF EXISTS idx_posts_accepted_answer;
DROP INDEX IF EXISTS idx_posts_bounty;

-- Remove columns
ALTER TABLE posts DROP COLUMN IF EXISTS accepted_answer_id;
ALTER TABLE users DROP COLUMN IF EXISTS reputation;
ALTER TABLE posts DROP COLUMN IF EXISTS bounty_amount;