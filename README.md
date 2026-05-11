# HoneyGarden

Community board for posts, RSS-fed links, small developer profiles, chat, roadmap feedback, badges, and Git hosting integrations.

![HoneyGarden forum preview](docs/honeygarden-preview.png)

## Stack

- **Frontend:** React, Vite, React Router, TanStack Query, OIDC client
- **Backend:** Go, Gin, pgx, NATS JetStream, Meilisearch, MinIO/S3
- **Identity:** Zitadel / OIDC
- **Storage:** PostgreSQL, MinIO
- **Search:** Meilisearch
- **Async:** NATS JetStream
- **Observability:** Prometheus, Grafana, Loki, Tempo, Grafana Alloy
- **Deploy:** Docker Compose, Caddy

## Project Tree

```text
.
├── apps
│   └── forum
│       ├── backend
│       │   └── honeygarden
│       │       ├── cmd
│       │       └── internal
│       │           ├── adapter
│       │           │   ├── http
│       │           │   ├── meilisearch
│       │           │   ├── nats
│       │           │   ├── oauthclients
│       │           │   ├── objectstore
│       │           │   ├── postgres
│       │           │   └── wakapi
│       │           ├── config
│       │           ├── domain
│       │           ├── metrics
│       │           ├── port
│       │           ├── service
│       │           └── worker
│       ├── frontend
│       │   ├── public
│       │   └── src
│       │       ├── components
│       │       ├── data
│       │       ├── lib
│       │       └── screens
│       └── migrations
├── platform
│   └── deploy
│       ├── alloy
│       ├── caddy
│       ├── grafana
│       ├── postgres
│       ├── prometheus
│       ├── tempo
│       └── zitadel
├── docker-compose.yml
└── Makefile
```

## Backend Services

- **UserService:** profile data, onboarding, follows, blocks, friends, admin bans, and pinned repository views.
- **CommunityService:** communities, membership, moderators, stars, community-level bans, and posting permissions.
- **PostService:** posts, comments, votes, bookmarks, pins, media links, RSS-backed external posts, and moderation actions.
- **FeedService:** aggregated feed from local posts and configured RSS sources.
- **SearchService:** indexes and queries posts and communities through Meilisearch.
- **NotificationService:** notification creation, unread counts, and notification preferences.
- **ChatService:** shared realtime chat history and WebSocket fan-out.
- **IntegrationService:** GitHub, GitLab, and Codeberg OAuth flows, repository sync, token refresh, and pinned repository selection.
- **WakapiService:** connects a user's Wakapi API key and exposes coding activity stats.
- **BadgeService:** badge definitions and earned badge stats.
- **RoadmapService:** public roadmap items and admin roadmap management.
- **FeedbackService:** feedback ideas, votes, status changes, and admin cleanup.
- **BannedWordService:** moderation dictionary for blocked words.
- **OnlineService:** heartbeat-based online presence.

## Runtime Services

- **honeygarden:** Go API, background workers, metrics endpoint, auth middleware, and business logic.
- **forum-web:** React frontend served behind nginx.
- **postgres:** main relational database for users, posts, chat, integrations, and settings.
- **migrator:** applies SQL migrations before the API starts.
- **zitadel:** OIDC identity provider.
- **nats:** JetStream event bus for notifications, indexing, and badges.
- **meilisearch:** full-text search index.
- **minio:** S3-compatible object storage for uploads.
- **wakapi:** optional coding activity service.
- **prometheus, grafana, loki, tempo, alloy:** metrics, dashboards, logs, traces, and collection.
- **caddy:** edge proxy and TLS configuration.
