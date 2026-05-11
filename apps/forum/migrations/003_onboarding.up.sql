-- honeydrop: onboarding flag
-- golang-migrate: UP

ALTER TABLE users ADD COLUMN onboarding_done BOOL NOT NULL DEFAULT FALSE;
