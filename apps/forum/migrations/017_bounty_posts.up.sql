-- Add bounty_amount column to posts table
ALTER TABLE posts ADD COLUMN bounty_amount INTEGER DEFAULT 0;

-- Add reputation column to users table
ALTER TABLE users ADD COLUMN reputation INTEGER DEFAULT 0;

-- Add accepted_answer_id column to posts table for marking accepted answers
ALTER TABLE posts ADD COLUMN accepted_answer_id INTEGER REFERENCES posts(id);

-- Create index for bounty posts
CREATE INDEX idx_posts_bounty ON posts(bounty_amount) WHERE bounty_amount > 0;

-- Create index for accepted answers
CREATE INDEX idx_posts_accepted_answer ON posts(accepted_answer_id);