.PHONY: up down restart logs ps build pull deploy migrate seed migrate-down

up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose restart

logs:
	docker compose logs -f --tail=100

ps:
	docker compose ps

build:
	docker compose build

build-forum:
	docker compose build api-gateway forum-web user-service community-service \
	    post-service notification-service billing-service search-service feed-service

pull:
	docker compose pull

deploy: pull build-forum
	docker compose up -d --remove-orphans

# Накатить миграции вручную
migrate:
	docker compose run --rm migrator

migrate-down:
	docker compose run --rm migrator \
	    -path /migrations \
	    -database "postgres://$${POSTGRES_USER}:$${POSTGRES_PASSWORD}@postgres:5432/honeydrop?sslmode=disable" \
	    down 1

# Залить начальные RSS источники (запускать после migrate)
seed:
	docker compose exec postgres psql -U $${POSTGRES_USER} -d honeydrop \
	    -f /seeds/rss_sources_seed.sql

caddy-reload:
	docker compose exec caddy caddy reload --config /etc/caddy/Caddyfile

zitadel-logs:
	docker compose logs -f zitadel
