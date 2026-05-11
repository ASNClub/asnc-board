-- Снос Forjego (Forgejo). Таблицы пересоздадутся в фазе интеграций
-- GitHub/GitLab/Codeberg с обновлённой схемой (refresh_token, expires_at).
DROP TABLE IF EXISTS pinned_repos;
DROP TABLE IF EXISTS user_git_accounts;
DROP INDEX IF EXISTS idx_user_git_accounts_user;
DROP INDEX IF EXISTS idx_pinned_repos_user;
