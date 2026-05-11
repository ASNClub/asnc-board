-- Запускать после миграций: psql -f rss_sources_seed.sql
-- ON CONFLICT (url) DO UPDATE — повторный запуск перезаписывает name/site_url/
-- favicon_url/tags на seed-значения. Источники, добавленные пользователями через
-- API, не трогаются (у них другой URL).

INSERT INTO rss_sources (url, name, site_url, favicon_url, tags) VALUES

-- Go
('https://go.dev/blog/feed.atom',
 'The Go Blog',
 'https://go.dev/blog',
 'https://go.dev/favicon.ico',
 ARRAY['go', 'backend']),

('https://golangweekly.com/rss/',
 'Golang Weekly',
 'https://golangweekly.com',
 'https://golangweekly.com/favicon.ico',
 ARRAY['go', 'backend']),

('https://threedots.tech/index.xml',
 'Three Dots Labs',
 'https://threedots.tech',
 'https://threedots.tech/favicon.ico',
 ARRAY['go', 'backend', 'architecture']),

('https://www.ardanlabs.com/blog/index.xml',
 'Ardan Labs Blog',
 'https://www.ardanlabs.com/blog',
 'https://www.ardanlabs.com/favicon.png',
 ARRAY['go', 'backend']),

('https://bitfieldconsulting.com/golang?format=rss',
 'Bitfield Consulting — Go',
 'https://bitfieldconsulting.com/golang',
 'https://bitfieldconsulting.com/favicon.ico',
 ARRAY['go', 'backend']),

-- Базы данных
('https://brandur.org/articles.atom',
 'brandur.org',
 'https://brandur.org',
 'https://brandur.org/assets/images/favicon/favicon-128.jpg',
 ARRAY['postgres', 'databases', 'distributed', 'backend']),

('https://www.percona.com/blog/feed/',
 'Percona Blog',
 'https://www.percona.com/blog',
 'https://www.percona.com/favicon.ico',
 ARRAY['databases', 'sql', 'postgres', 'performance']),

-- Кеши / стриминг
('https://redis.io/blog/feed/',
 'Redis Blog',
 'https://redis.io/blog',
 'https://redis.io/favicon.ico',
 ARRAY['redis', 'databases', 'performance']),

('https://www.confluent.io/feed/',
 'Confluent Blog',
 'https://www.confluent.io/blog',
 'https://www.confluent.io/favicon.ico',
 ARRAY['kafka', 'distributed', 'backend']),

-- Архитектура
('https://highscalability.com/feed/',
 'High Scalability',
 'https://highscalability.com',
 'https://highscalability.com/favicon.ico',
 ARRAY['architecture', 'distributed', 'performance', 'backend']),

('https://medium.com/feed/netflix-techblog',
 'Netflix Tech Blog',
 'https://netflixtechblog.com',
 'https://miro.medium.com/v2/1*m-R_BkNf1Qjr1YbyOIJY2w.png',
 ARRAY['architecture', 'distributed', 'backend']),

('https://stripe.com/blog/feed.rss',
 'Stripe Blog',
 'https://stripe.com/blog',
 'https://stripe.com/favicon.ico',
 ARRAY['architecture', 'backend', 'performance']),

('https://github.blog/category/engineering/feed/',
 'GitHub Engineering',
 'https://github.blog/category/engineering',
 'https://github.blog/favicon.ico',
 ARRAY['architecture', 'backend', 'devops']),

('https://blog.cloudflare.com/rss/',
 'Cloudflare Blog',
 'https://blog.cloudflare.com',
 'https://blog.cloudflare.com/favicon.ico',
 ARRAY['networking', 'cloud', 'performance', 'security']),

('https://newsletter.pragmaticengineer.com/feed',
 'The Pragmatic Engineer',
 'https://newsletter.pragmaticengineer.com',
 'https://newsletter.pragmaticengineer.com/favicon.ico',
 ARRAY['architecture', 'backend']),

-- Хуябр
('https://habr.com/ru/rss/hubs/go/articles/?fl=ru',
 'Хабр — Go',
 'https://habr.com/ru/hub/go/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'go', 'backend']),

('https://habr.com/ru/rss/hubs/postgresql/articles/?fl=ru',
 'Хабр — PostgreSQL',
 'https://habr.com/ru/hub/postgresql/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'postgres', 'databases', 'sql']),

('https://habr.com/ru/rss/hubs/sql/articles/?fl=ru',
 'Хабр — SQL',
 'https://habr.com/ru/hub/sql/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'sql', 'databases']),

('https://habr.com/ru/rss/hubs/nosql/articles/?fl=ru',
 'Хабр — NoSQL',
 'https://habr.com/ru/hub/nosql/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'nosql', 'databases']),

('https://habr.com/ru/rss/hubs/kubernetes/articles/?fl=ru',
 'Хабр — Kubernetes',
 'https://habr.com/ru/hub/kubernetes/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'kubernetes', 'devops']),

('https://habr.com/ru/rss/hubs/devops/articles/?fl=ru',
 'Хабр — DevOps',
 'https://habr.com/ru/hub/devops/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'devops']),

('https://habr.com/ru/rss/hubs/microservices/articles/?fl=ru',
 'Хабр — Микросервисы',
 'https://habr.com/ru/hub/microservices/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'microservices', 'architecture', 'backend']),

('https://habr.com/ru/rss/hubs/system_programming/articles/?fl=ru',
 'Хабр — Системное программирование',
 'https://habr.com/ru/hub/system_programming/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'cpp', 'performance']),

-- Rust(осуждаю)
('https://blog.rust-lang.org/feed.xml',
 'Rust Blog',
 'https://blog.rust-lang.org',
 'https://www.rust-lang.org/static/images/favicon-32x32.png',
 ARRAY['rust', 'backend']),

('https://this-week-in-rust.org/rss.xml',
 'This Week in Rust',
 'https://this-week-in-rust.org',
 'https://this-week-in-rust.org/favicon.ico',
 ARRAY['rust', 'backend']),

('https://fasterthanli.me/index.xml',
 'fasterthanli.me',
 'https://fasterthanli.me',
 'https://fasterthanli.me/favicon-32x32.png',
 ARRAY['rust', 'performance', 'backend']),

-- Python(тоже но сфиля любит питончик гонять)
('https://realpython.com/atom.xml',
 'Real Python',
 'https://realpython.com',
 'https://realpython.com/static/favicon.68cbf4197b0c.png',
 ARRAY['python', 'backend']),

('https://blog.python.org/feeds/posts/default',
 'Python Insider',
 'https://blog.python.org',
 'https://www.python.org/static/favicon.ico',
 ARRAY['python']),

-- JS / TS / Web(кам)
('https://2ality.com/feeds/posts.atom',
 '2ality — Axel Rauschmayer',
 'https://2ality.com',
 'https://2ality.com/favicon.ico',
 ARRAY['javascript', 'typescript', 'webdev']),

('https://devblogs.microsoft.com/typescript/feed/',
 'TypeScript Blog',
 'https://devblogs.microsoft.com/typescript',
 'https://devblogs.microsoft.com/wp-content/uploads/2018/12/cropped-Microsoft-Logo-32x32.png',
 ARRAY['typescript', 'javascript', 'webdev']),

('https://v8.dev/blog.atom',
 'V8 Dev Blog',
 'https://v8.dev/blog',
 'https://v8.dev/_img/favicon-32x32.png',
 ARRAY['javascript', 'performance']),

('https://web.dev/feed.xml',
 'web.dev',
 'https://web.dev',
 'https://web.dev/images/favicon.ico',
 ARRAY['webdev', 'performance', 'frontend']),

('https://css-tricks.com/feed/',
 'CSS-Tricks',
 'https://css-tricks.com',
 'https://css-tricks.com/favicon.ico',
 ARRAY['webdev', 'frontend']),

('https://www.smashingmagazine.com/feed/',
 'Smashing Magazine',
 'https://www.smashingmagazine.com',
 'https://www.smashingmagazine.com/images/favicon/favicon.ico',
 ARRAY['webdev', 'frontend']),

-- C / C++
('https://herbsutter.com/feed/',
 'Sutter''s Mill',
 'https://herbsutter.com',
 'https://herbsutter.com/favicon.ico',
 ARRAY['cpp', 'performance']),

-- architecture
('https://www.allthingsdistributed.com/atom.xml',
 'All Things Distributed',
 'https://www.allthingsdistributed.com',
 'https://www.allthingsdistributed.com/images/favicon.ico',
 ARRAY['distributed', 'architecture', 'cloud']),

('https://blog.acolyer.org/feed/',
 'The Morning Paper',
 'https://blog.acolyer.org',
 'https://blog.acolyer.org/favicon.ico',
 ARRAY['distributed', 'architecture']),

('https://martinfowler.com/feed.atom',
 'martinfowler.com',
 'https://martinfowler.com',
 'https://martinfowler.com/favicon.ico',
 ARRAY['architecture', 'backend']),

('https://microservices.io/index.xml',
 'microservices.io',
 'https://microservices.io',
 'https://microservices.io/favicon.ico',
 ARRAY['microservices', 'architecture', 'backend']),

-- Observability
('https://www.honeycomb.io/feed',
 'Honeycomb Blog',
 'https://www.honeycomb.io/blog',
 'https://www.honeycomb.io/favicon.ico',
 ARRAY['observability', 'distributed']),

-- Postgres
('https://www.crunchydata.com/blog/rss.xml',
 'Crunchy Data',
 'https://www.crunchydata.com/blog',
 'https://www.crunchydata.com/favicon.ico',
 ARRAY['postgres', 'databases', 'sql']),

('https://planet.postgresql.org/rss20.xml',
 'Planet PostgreSQL',
 'https://planet.postgresql.org',
 'https://planet.postgresql.org/favicon.ico',
 ARRAY['postgres', 'databases', 'sql']),

-- Kubernetes (майнкрафт)
('https://kubernetes.io/feed.xml',
 'Kubernetes Blog',
 'https://kubernetes.io/blog',
 'https://kubernetes.io/images/favicon.png',
 ARRAY['kubernetes', 'devops', 'cloud']),

('https://aws.amazon.com/blogs/architecture/feed/',
 'AWS Architecture Blog',
 'https://aws.amazon.com/blogs/architecture',
 'https://a0.awsstatic.com/main/images/site/favicon.ico',
 ARRAY['cloud', 'architecture']),

-- Security
('https://krebsonsecurity.com/feed/',
 'Krebs on Security',
 'https://krebsonsecurity.com',
 'https://krebsonsecurity.com/favicon.ico',
 ARRAY['security']),

('https://www.schneier.com/feed/atom/',
 'Schneier on Security',
 'https://www.schneier.com',
 'https://www.schneier.com/favicon.ico',
 ARRAY['security']),

-- ML  
('https://openai.com/blog/rss.xml',
 'OpenAI Blog',
 'https://openai.com/blog',
 'https://openai.com/favicon.ico',
 ARRAY['ml']),

('https://www.anthropic.com/news/rss.xml',
 'Anthropic News',
 'https://www.anthropic.com/news',
 'https://www.anthropic.com/favicon.ico',
 ARRAY['ml']),

-- больше хуябра
('https://habr.com/ru/rss/hubs/rust/articles/?fl=ru',
 'Хабр — Rust',
 'https://habr.com/ru/hub/rust/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'rust']),

('https://habr.com/ru/rss/hubs/python/articles/?fl=ru',
 'Хабр — Python',
 'https://habr.com/ru/hub/python/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'python']),

('https://habr.com/ru/rss/hubs/javascript/articles/?fl=ru',
 'Хабр — JavaScript',
 'https://habr.com/ru/hub/javascript/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'javascript', 'webdev']),

('https://habr.com/ru/rss/hubs/typescript/articles/?fl=ru',
 'Хабр — TypeScript',
 'https://habr.com/ru/hub/typescript/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'typescript', 'webdev']),

('https://habr.com/ru/rss/hubs/cpp/articles/?fl=ru',
 'Хабр — C++',
 'https://habr.com/ru/hub/cpp/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'cpp']),

('https://habr.com/ru/rss/hubs/infosecurity/articles/?fl=ru',
 'Хабр — Информационная безопасность',
 'https://habr.com/ru/hub/infosecurity/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'security']),

('https://habr.com/ru/rss/hubs/open_source/articles/?fl=ru',
 'Хабр — Open Source',
 'https://habr.com/ru/hub/open_source/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'opensource']),

('https://habr.com/ru/rss/hubs/system_design/articles/?fl=ru',
 'Хабр — System Design',
 'https://habr.com/ru/hub/system_design/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'architecture', 'distributed']),

('https://habr.com/ru/rss/hubs/ml/articles/?fl=ru',
 'Хабр — Machine Learning',
 'https://habr.com/ru/hub/machine_learning/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'ml']),

('https://habr.com/ru/rss/hubs/network_technologies/articles/?fl=ru',
 'Хабр — Сетевые технологии',
 'https://habr.com/ru/hub/network_technologies/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'networking']),

('https://habr.com/ru/rss/hubs/cloud_computing/articles/?fl=ru',
 'Хабр — Cloud Computing',
 'https://habr.com/ru/hub/cloud_computing/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'cloud']),

('https://habr.com/ru/rss/hubs/linux/articles/?fl=ru',
 'Хабр — Linux',
 'https://habr.com/ru/hub/linux/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'linux']),

('https://habr.com/ru/rss/hubs/docker/articles/?fl=ru',
 'Хабр — Docker',
 'https://habr.com/ru/hub/docker/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'docker', 'devops']),

('https://habr.com/ru/rss/hubs/gamedev/articles/?fl=ru',
 'Хабр — GameDev',
 'https://habr.com/ru/hub/gamedev/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'gamedev']),

('https://habr.com/ru/rss/hubs/development_for_android/articles/?fl=ru',
 'Хабр — Android-разработка',
 'https://habr.com/ru/hub/development_for_android/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'mobile']),

('https://habr.com/ru/rss/hubs/development_for_ios/articles/?fl=ru',
 'Хабр — iOS-разработка',
 'https://habr.com/ru/hub/development_for_ios/',
 'https://habr.com/favicon.ico',
 ARRAY['ru', 'mobile'])

ON CONFLICT (url) DO UPDATE SET
  name        = EXCLUDED.name,
  site_url    = EXCLUDED.site_url,
  favicon_url = EXCLUDED.favicon_url,
  tags        = EXCLUDED.tags;
