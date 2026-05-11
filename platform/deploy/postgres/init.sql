-- Создаём отдельные БД для других сервисов.
-- БД honeygarden создаётся через POSTGRES_DB в docker-compose.
-- БД zitadel создаётся самим Zitadel (start-from-init).
-- Схема honeygarden накатывается через migrator (apps/forum/migrations).

CREATE DATABASE wakapi;
