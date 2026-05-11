// Доступные пользователю теги интересов. Совпадают с тегами на RSS-источниках
// в platform/deploy/postgres/rss_sources_seed.sql — лента отбирает посты по
// пересечению.
export const INTEREST_TAGS = [
  // Языки
  { id: 'go',         l: 'Go' },
  { id: 'rust',       l: 'Rust' },
  { id: 'python',     l: 'Python' },
  { id: 'javascript', l: 'JavaScript' },
  { id: 'typescript', l: 'TypeScript' },
  { id: 'cpp',        l: 'C / C++' },

  // Бэкенд / архитектура
  { id: 'backend',       l: 'Backend' },
  { id: 'architecture',  l: 'Архитектура' },
  { id: 'microservices', l: 'Микросервисы' },
  { id: 'distributed',   l: 'Distributed' },
  { id: 'performance',   l: 'Performance' },
  { id: 'observability', l: 'Observability' },

  // Данные
  { id: 'databases', l: 'Базы данных' },
  { id: 'sql',       l: 'SQL' },
  { id: 'postgres',  l: 'PostgreSQL' },
  { id: 'nosql',     l: 'NoSQL' },
  { id: 'redis',     l: 'Redis / Cache' },
  { id: 'kafka',     l: 'Kafka / Streaming' },

  // Инфра / ops
  { id: 'devops',     l: 'DevOps' },
  { id: 'kubernetes', l: 'Kubernetes' },
  { id: 'docker',     l: 'Docker' },
  { id: 'linux',      l: 'Linux' },
  { id: 'cloud',      l: 'Cloud' },
  { id: 'networking', l: 'Networking' },

  // Прочее
  { id: 'security',   l: 'Security' },
  { id: 'ml',         l: 'ML / AI' },
  { id: 'webdev',     l: 'Web / Frontend' },
  { id: 'mobile',     l: 'Mobile' },
  { id: 'gamedev',    l: 'GameDev' },
  { id: 'opensource', l: 'Open Source' },
  { id: 'ru',         l: 'Русское' },
];
