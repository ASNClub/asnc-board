// Конвенции query-ключей для @tanstack/react-query.
// Иерархия: ['domain', id?, sub?, ...args] — invalidation идёт по префиксу.
export const qk = {
  me: () => ['me'],
  feed: (mode) => ['feed', mode],
  trending: (limit = 20) => ['trending', limit],
  online: () => ['online'],
  myCommunities: () => ['my-communities'],
  myCommunity: () => ['my-community'],
  communitiesList: (sort) => ['communities', 'list', sort],

  post: (id) => ['post', id],
  postComments: (id) => ['post', id, 'comments'],

  community: (slug) => ['community', slug],
  communityPosts: (slug, tag = null) => ['community', slug, 'posts', tag],
  communityMods: (slug) => ['community', slug, 'mods'],

  user: (username) => ['user', username],
  userPosts: (username) => ['user', username, 'posts'],
  userActivity: (username) => ['user', username, 'activity'],
  userWakapi: (username) => ['user', username, 'wakapi'],
  userBadges: (username) => ['user', username, 'badges'],
  userPinnedRepos: (username) => ['user', username, 'pinned-repos'],
  userFollowers: (username) => ['user', username, 'followers'],
  userFollowing: (username) => ['user', username, 'following'],

  notifications: () => ['notifications'],
  notificationPrefs: () => ['notifications', 'prefs'],

  badgeDefs: () => ['badge-defs'],
  bookmarks: () => ['bookmarks'],

  searchCommunities: (q) => ['search', 'communities', q],
  searchPosts: (q) => ['search', 'posts', q],
  searchUsers: (q) => ['search', 'users', q],

  integrations: () => ['integrations'],

  myBlocks: () => ['my-blocks'],

  feedback: (sort, status) => ['feedback', sort, status],

  roadmapItems: () => ['roadmap-items'],
  bannedWords: () => ['banned-words'],
};
