-- honeydrop: onboarding flag
-- golang-migrate: DOWN

ALTER TABLE users DROP COLUMN IF EXISTS onboarding_done;
