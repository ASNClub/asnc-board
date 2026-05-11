// Relative time (ru)
export function relativeTime(iso) {
  const diff = Math.floor((Date.now() - new Date(iso)) / 1000);
  if (diff < 60)    return 'только что';
  if (diff < 3600)  return `${Math.floor(diff / 60)} мин`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} ч`;
  const d = Math.floor(diff / 86400);
  if (d < 30)       return `${d} д`;
  if (d < 365)      return `${Math.floor(d / 30)} мес`;
  return `${Math.floor(d / 365)} г`;
}

// Инициалы из username / displayName
export function initials(name = '') {
  return name.slice(0, 2).toUpperCase() || '??';
}

// Цвет сообщества по slug
export const COMM_COLORS = {
  'rust-lang':    '#D08770',
  'systems':      '#81A1C1',
  'frontend':     '#B48EAD',
  'architecture': '#A3BE8C',
  'showcase':     '#E09832',
  'devtools':     '#8FBCBB',
  'reading':      '#EBCB8B',
  'golang':       '#00ADD8',
  'python':       '#3572A5',
  'linux':        '#F5C55A',
};
export const commColor = (slug) => COMM_COLORS[slug] ?? '#8FBCBB';

// Prefer shortId (8-char base62) over UUID for shareable URLs.
export const postUrl = (post) => `/p/${post?.shortId || post?.id}`;

// FeedItem (camelCase from backend) → формат PostCard
export function feedItemToPost(item) {
  const isRSS = item.type === 'rss';
  return {
    id:           item.postId ?? item.id,
    shortId:      item.shortId ?? null,
    community:    item.communitySlug ?? null,
    commColor:    commColor(item.communitySlug),
    avatar:       initials(item.author?.username ?? '?'),
    author:       item.author?.username ?? null,
    authorAvatar: item.author?.avatarUrl ?? null,
    authorRep:    null,
    time:         relativeTime(item.publishedAt),
    title:        item.title,
    excerpt:      item.summary ?? '',
    tags:         item.tags ?? [],
    comments:     item.commentsCount ?? 0,
    votes:        item.upvotes ?? 0,
    thumb:        item.coverImage ?? null,
    type:         item.type,
    isRSS,
    sourceName:   isRSS ? item.source?.name ?? null : null,
    sourceFavicon: isRSS ? item.source?.faviconUrl ?? null : null,
    externalUrl:  isRSS ? item.url ?? null : null,
    voted:        item.isVoted ? 'up' : null,
    bookmarked:   !!item.isBookmarked,
    kind:         item.kind ?? null,
  };
}

// PostView (camelCase, enriched with author + communitySlug) → формат PostCard
export function postToCard(post) {
  const isRSS = !!post.sourceId;
  const authorName = post.author?.username ?? null;
  const slug = post.communitySlug ?? null;
  return {
    id:           post.id,
    shortId:      post.shortId ?? null,
    community:    slug,
    commColor:    commColor(slug),
    avatar:       initials(authorName ?? '?'),
    author:       authorName,
    authorAvatar: post.author?.avatarUrl ?? null,
    authorRep:    post.author?.reputation ?? null,
    time:         relativeTime(post.createdAt),
    title:        post.title ?? post.content?.slice(0, 80),
    excerpt:      post.content ?? '',
    tags:         post.tags ?? [],
    comments:     0,
    votes:        post.votesCount ?? 0,
    thumb:        post.coverImageUrl || null,
    type:         isRSS ? 'rss' : 'post',
    isRSS,
    sourceName:   isRSS ? post.source?.name ?? null : null,
    sourceFavicon: isRSS ? post.source?.faviconUrl ?? null : null,
    externalUrl:  isRSS ? post.externalUrl ?? null : null,
    voted:        post.isVoted ? 'up' : null,
    bookmarked:   !!post.isBookmarked,
    kind:         post.kind ?? null,
  };
}

// Excerpt: strip MD images / inline code, truncate by words.
export function excerpt(text = '', maxChars = 240) {
  if (!text) return '';
  let s = text
    .replace(/<style[\s\S]*?<\/style>/gi, '')
    .replace(/<script[\s\S]*?<\/script>/gi, '')
    .replace(/<!--[\s\S]*?-->/g, '')
    .replace(/<[^>]+>/g, ' ')
    .replace(/&nbsp;/g, ' ')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&[a-z]+;/gi, '')
    .replace(/!\[[^\]]*\]\([^)]*\)/g, '')
    .replace(/```[\s\S]*?```/g, '')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
    .replace(/^#{1,6}\s+/gm, '')
    .replace(/^>\s?/gm, '')
    .replace(/^\s*[-*+]\s+/gm, '')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\*([^*]+)\*/g, '$1')
    .replace(/_([^_]+)_/g, '$1')
    .replace(/\n{2,}/g, ' ')
    .replace(/\n/g, ' ')
    .replace(/\s+/g, ' ')
    .replace(/[<>\/]+\s*$/g, '')
    .trim();
  if (s.length <= maxChars) return s;
  const cut = s.slice(0, maxChars);
  const lastSpace = cut.lastIndexOf(' ');
  let out = lastSpace > maxChars * 0.6 ? cut.slice(0, lastSpace) : cut;
  out = out.replace(/[<>\/\s]+$/g, '');
  return out + '…';
}

// PostView (camelCase) → нормализованный объект
export function normalizePost(p) {
  const isRSS = !!p.sourceId;
  return {
    id:            p.id,
    shortId:       p.shortId ?? null,
    communityId:   p.communityId ?? null,
    communitySlug: p.communitySlug ?? null,
    authorId:      p.authorId ?? null,
    author:        p.author ?? null,
    title:         p.title ?? null,
    content:       p.content ?? '',
    media:         p.media ?? [],
    views:         p.viewsCount ?? 0,
    votes:         p.votesCount ?? 0,
    isPinned:      p.isPinned ?? false,
    createdAt:     p.createdAt,
    isRSS,
    sourceId:      p.sourceId ?? null,
    source:        p.source ?? null,
    externalUrl:   p.externalUrl ?? null,
    coverImageUrl: p.coverImageUrl ?? null,
    tags:          p.tags ?? [],
    voted:         p.isVoted ? 'up' : null,
    bookmarked:    !!p.isBookmarked,
    kind:          p.kind ?? null,
  };
}

// CommentView (camelCase) → нормализованный объект
export function normalizeComment(c) {
  return {
    id:        c.id,
    postId:    c.postId,
    authorId:  c.authorId,
    author:    c.author ?? null,
    parentId:  c.parentId ?? null,
    content:   c.content ?? '',
    votes:     c.votesCount ?? 0,
    createdAt: c.createdAt,
  };
}

// Flat comments → nested tree
export function buildCommentTree(comments) {
  const map = {};
  const roots = [];
  for (const c of comments) {
    map[c.id] = { ...c, replies: [] };
  }
  for (const c of comments) {
    if (c.parentId && map[c.parentId]) {
      map[c.parentId].replies.push(map[c.id]);
    } else {
      roots.push(map[c.id]);
    }
  }
  return roots;
}

// Notification (camelCase) → нормализованный объект
export function normalizeNotification(n) {
  let payload = {};
  if (n.payload) {
    try { payload = typeof n.payload === 'string' ? JSON.parse(n.payload) : n.payload; } catch {}
  }
  const typeMap = {
    'user.followed':       'follow',
    'friendship.requested':'follow',
    'friendship.accepted': 'follow',
    'community.followed':  'follow',
    'community.starred':   'vote',
    'comment.created':     'comment',
    'post.voted':          'vote',
    'post.created':        'comment',
  };
  const textMap = {
    'user.followed':       'подписался на тебя',
    'friendship.requested':'отправил запрос на дружбу',
    'friendship.accepted': 'принял запрос на дружбу',
    'community.followed':  'подписался на сообщество',
    'community.starred':   'добавил в избранное',
    'comment.created':     'прокомментировал твой пост',
    'post.voted':          'проголосовал за твой пост',
    'post.created':        'опубликовал пост',
  };
  return {
    id:    n.id,
    type:  typeMap[n.type] ?? 'comment',
    read:  n.isRead ?? false,
    time:  relativeTime(n.createdAt),
    actor: payload.username ? { a: initials(payload.username), n: payload.username } : null,
    text:  textMap[n.type] ?? n.type,
    ref:   payload.post_title ?? payload.community_slug ?? null,
    refId: payload.post_id ?? null,
  };
}

// Community (camelCase) → нормализованный объект
export function normalizeCommunity(c) {
  return {
    id:          c.id,
    slug:        c.slug,
    name:        c.name,
    description: c.description ?? '',
    avatarUrl:   c.avatarUrl ?? null,
    bannerUrl:   c.bannerUrl ?? null,
    members:     c.followersCount ?? 0,
    posts:       c.postsCount ?? 0,
    stars:       c.starsCount ?? 0,
    tags:        c.tags ?? [],
    rules:       c.rules ?? [],
    color:       commColor(c.slug),
  };
}
