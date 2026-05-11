CREATE TABLE feedback (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  type        VARCHAR(16) NOT NULL CHECK (type IN ('idea','bug','question','other')),
  title       VARCHAR(200) NOT NULL,
  body        TEXT NOT NULL,
  author_id   UUID REFERENCES users(id) ON DELETE SET NULL,
  is_anon     BOOLEAN NOT NULL DEFAULT FALSE,
  status      VARCHAR(16) NOT NULL DEFAULT 'open'
              CHECK (status IN ('open','planned','in_progress','done','rejected')),
  votes_count INT NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_feedback_status ON feedback(status, created_at DESC);
CREATE INDEX idx_feedback_votes ON feedback(votes_count DESC, created_at DESC);

CREATE TABLE feedback_votes (
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  feedback_id UUID NOT NULL REFERENCES feedback(id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, feedback_id)
);
