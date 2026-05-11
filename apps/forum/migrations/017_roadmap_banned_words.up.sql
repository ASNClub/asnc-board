-- Roadmap items (admin-managed, public-readable)
CREATE TABLE IF NOT EXISTS roadmap_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phase       TEXT NOT NULL DEFAULT 'next',       -- wip, next, later, done
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    tags        JSONB NOT NULL DEFAULT '[]',
    eta         TEXT,                                -- free-form: "ETA · 2 недели", "Q3", "шаг 22"
    featured    BOOLEAN NOT NULL DEFAULT false,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_roadmap_items_phase ON roadmap_items (phase, sort_order);

-- Banned words for username / community slug validation
CREATE TABLE IF NOT EXISTS banned_words (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    word       TEXT NOT NULL,
    scope      TEXT NOT NULL DEFAULT 'both',  -- username, slug, both
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_banned_words_word_scope UNIQUE (word, scope)
);

-- Seed existing roadmap items from STATIC_PHASES
INSERT INTO roadmap_items (phase, title, description, tags, eta, featured, sort_order) VALUES
-- wip
('wip', 'Чат v2 — общий + DM', 'Стабильный продакшн-чат поверх gorilla/websocket. Комнаты по сообществам, личные DM, реакции, упоминания, индикаторы typing. Бэкенд готов, фронт ждёт финального API.', '["websocket","realtime","bigfeature"]', 'ETA · 2-3 недели', true, 0),
('wip', '@mention пользователей', 'Тегать юзеров в постах и комментах через @username. Автокомплит в редакторе по searchUsers, рендер в ссылку на профиль, push в notifications.', '["editor","notif"]', 'ETA · ~3 дня', false, 1),
('wip', 'Тёмная тема — полный pass', 'Scaffold уже есть. Осталось: причесать оставшиеся виджеты, вычистить hardcoded цвета, проверить контрасты на каждой странице.', '["ui","polish"]', 'ETA · 1 неделя', false, 2),

-- next
('next', 'Глобальная репутация', 'Видимый счёт у каждого юзера. Растёт от апвоутов на постах и комментах, принятых ответов, активности в сообществах. Падает за минусы и баны. Прозрачные правила, история начислений в профиле.', '["community","backend","bigfeature"]', 'Q3', true, 0),
('next', 'Система вознаграждений', 'Разблокировки по репутации: кастомные бейджи, право создавать сообщество, доступ к bounty-постам, эмодзи-реакции, флаги «trusted». Без P2W — только за вклад.', '["gamification","community"]', 'Q3', true, 1),
('next', 'Post-pin UI для модераторов', 'Бэк уже готов (POST /posts/:id/pin). Нужна кнопка в dots-меню для owner/mod, индикатор в ленте сообщества.', '["community","quickwin"]', '~1 день', false, 2),
('next', 'User-block — полный уровень', 'Сейчас «средний»: посты/комменты блокированного скрыты двусторонне. Добавить: запрет фолловить блокера, фильтр упоминаний и комментов в notif-сервисе.', '["privacy","notif"]', '~2 дня', false, 3),
('next', 'Markdown-редактор v2', 'Drag-n-drop картинок прямо в textarea, UI для таблиц, подсветка синтаксиса в code-блоках, командная палитра. GitHub-style preview с syncscroll.', '["editor","ui"]', NULL, false, 4),
('next', 'Поиск с фильтрами', 'Сейчас Meilisearch отдаёт плоский список. Добавить фасеты: сообщество, тег, автор, тип поста, временной диапазон.', '["search","meilisearch"]', NULL, false, 5),
('next', 'Бейджи v2 — кастомные правила', 'Сейчас 12 хардкод-правил. Перенести в БД, дать админу UI для создания (триггер, условие, иконка, описание).', '["admin","gamification"]', NULL, false, 6),
('next', 'Email-дайджест', 'Еженедельная сводка: топ-посты в твоих сообществах, новые ответы, упоминания. Через Resend (уже подключён).', '["notif","email"]', NULL, false, 7),
('next', 'Wakapi OIDC через Zitadel', 'Сейчас Wakapi подключается через manual API-key. Перевести на OIDC — общий логин с форумом, без копирования ключа.', '["oidc","wakapi"]', NULL, false, 8),
('next', 'GitLab + Codeberg back', 'OAuth-адаптеры лежат dormant в бэке. Вернуть в UI после теста на GitHub. Профильные репо-пины с любого провайдера.', '["oauth","integrations"]', NULL, false, 9),
('next', 'Экспорт данных', 'Архив всех твоих постов, комментов, подписок, закладок одним ZIP. Background-задача, ссылка в email когда готово.', '["privacy","gdpr"]', NULL, false, 10),
('next', 'Мобильный UI', 'Адаптив для телефона: bottom-nav вместо TopNav, single-column фид, drawer для left/right rail, touch-friendly размеры.', '["responsive","ui"]', NULL, false, 11),

-- later
('later', 'PWA + offline-кеш', 'Service worker для оффлайн-чтения уже посещённых тредов. Установка как app на телефон и десктоп.', '["pwa"]', NULL, false, 0),
('later', 'i18n — английский', 'Сейчас всё на русском. Внедрить i18next, начать с ru/en. Выделить строки из jsx, держать словари в JSON.', '["i18n"]', NULL, false, 1),
('later', 'Federation (ActivityPub)', 'Подписка из Mastodon на сообщества HG, кросс-постинг через ActivityPub. Большая работа, прирост пользы — под вопросом.', '["fediverse","research"]', NULL, false, 2),
('later', 'Bounty-посты', 'Q&A с наградой репутацией за принятый ответ. Автор поста ставит сумму (sink его репы), победитель забирает после accept.', '["community","experiment"]', NULL, false, 3),

-- done
('done', 'Короткие URL постов', 'Вместо 36-символьного UUID — 8-символьный base62 short_id. Старые ссылки совместимы.', '["ux","backend"]', 'шаг 22', false, 0),
('done', 'Discover-лист сообществ', 'Endpoint GET /communities?sort=popular|active|new. Sort tabs в /c теперь реально тасуют.', '["community"]', 'шаг 22', false, 1),
('done', 'Скрытые теги в ленте', 'Settings → Лента → скрытые теги реально дропают посты с этими тегами в Feed и Community.', '["feed"]', 'шаг 22', false, 2),
('done', 'Редактор интересов в Settings', 'Раньше теги задавались только в onboarding — теперь pills в Settings → Лента.', '["settings"]', 'шаг 22', false, 3),
('done', 'User-block (средний)', 'Двусторонний фильтр постов и комментов в feed-SQL и EnrichPosts/Comments. UI-кнопка в профиле + dots-меню.', '["privacy"]', 'шаг 19', false, 4),
('done', 'Миграция на @tanstack/react-query', 'Cross-screen sync: голос/follow/bookmark в одном экране сразу обновляют все остальные.', '["frontend","perf"]', 'шаг 18', false, 5),
('done', 'Notif с контекстом', 'Уведомления раньше были сырыми UUID — теперь резолвят actor/post/community + snippet.', '["notif"]', 'шаг 20', false, 6),
('done', 'GitHub OAuth + repo pins', 'Подключение GitHub в Settings → Connections. Закреп до 6 репо на профиле через чекбокс-список.', '["oauth","integrations"]', 'шаг 21', false, 7),
('done', 'Бизнес-метрики Prometheus', 'Новые посты, голоса, регистрации — в /metrics. Grafana-дашборд для трекинга трендов.', '["observability"]', 'шаг 17', false, 8),
('done', 'RSS как обсуждаемый пост', 'RSS-источники в общей таблице posts, не fake-сообщества. Голоса и комменты на статьи работают как у обычных.', '["rss","backend"]', 'шаг 12', false, 9),
('done', 'Мок-переписка фронта 1:1', 'Все 12+ экранов переписаны под mock.html с canonical layout: Feed, Thread, Community, Profile, Editor, Settings и др.', '["ui","refactor"]', 'шаги 1-11', false, 10);

-- Seed some default banned words
INSERT INTO banned_words (word, scope) VALUES
('admin', 'both'),
('administrator', 'both'),
('moderator', 'both'),
('mod', 'slug'),
('support', 'both'),
('help', 'slug'),
('honeygarden', 'both'),
('system', 'both'),
('root', 'both'),
('api', 'slug'),
('www', 'slug'),
('mail', 'slug'),
('ftp', 'slug'),
('null', 'both'),
('undefined', 'both'),
('anonymous', 'both'),
('bot', 'both'),
('official', 'both')
ON CONFLICT (word, scope) DO NOTHING;
