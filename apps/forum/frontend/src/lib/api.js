const BASE = '/api/v1';
const DEV_AUTH_ID = import.meta.env.VITE_DEV_AUTH_ID;

let _token = null;
let _onAuthExpired = null;
export function setAccessToken(token) { _token = token; }
export function setOnAuthExpired(fn)  { _onAuthExpired = fn; }

function handleUnauthorized() {
  _token = null;
  if (_onAuthExpired) _onAuthExpired();
}

async function request(path, init = {}) {
  const headers = { 'Content-Type': 'application/json', ...init.headers };
  if (DEV_AUTH_ID)  headers['X-Dev-Auth-ID']  = DEV_AUTH_ID;
  else if (_token)  headers['Authorization']   = `Bearer ${_token}`;

  const res = await fetch(`${BASE}${path}`, { ...init, headers });

  if (res.status === 401) {
    handleUnauthorized();
    throw new Error('unauthorized');
  }

  const text = await res.text();
  if (!text || res.status === 204) {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return null;
  }
  let json;
  try { json = JSON.parse(text); } catch { throw new Error(`HTTP ${res.status}: bad response`); }
  if (!res.ok || json.error) throw new Error(json?.error?.message ?? json?.error ?? `HTTP ${res.status}`);
  return json.data ?? json;
}

// ── Users ─────────────────────────────────────────────────────────────────────
// Response fields (json tags): id, authId, username, displayName, avatarUrl,
//   bannerUrl, bio, reputation, privacy, onboardingDone, tags, platforms, createdAt

export const getMe = () =>
  request('/users/me');

export const updateMe = (data) =>
  request('/users/me', { method: 'PUT', body: JSON.stringify(data) });

export const getUser = (username) =>
  request(`/users/${username}`);

export const getUserPosts = (username, offset = 0) =>
  request(`/users/${username}/posts?limit=20&offset=${offset}`);

export const getUserActivity = (username, offset = 0) =>
  request(`/users/${username}/activity?limit=20&offset=${offset}`);

export const followUser = (username) =>
  request(`/users/${username}/follow`, { method: 'POST' });

export const unfollowUser = (username) =>
  request(`/users/${username}/follow`, { method: 'DELETE' });

export const blockUser = (username) =>
  request(`/users/${username}/block`, { method: 'POST' });

export const unblockUser = (username) =>
  request(`/users/${username}/block`, { method: 'DELETE' });

export const getMyBlocks = () =>
  request('/users/me/blocks');

export const getMyFollowedCommunities = () =>
  request('/users/me/communities');

export const setUserTags = (tags) =>
  request('/users/me/tags', { method: 'PUT', body: JSON.stringify({ tags }) });

// ── Communities ───────────────────────────────────────────────────────────────
// Response fields (NO json tags → PascalCase): ID, Slug, Name, Description,
//   AvatarURL, BannerURL, FollowersCount, PostsCount, StarsCount, Tags, CreatedAt

export const getCommunity = (slug) =>
  request(`/communities/${slug}`);

export const getMyCommunity = () =>
  request('/communities/me');

export const starCommunity = (slug) =>
  request(`/communities/${slug}/star`, { method: 'POST' });

export const unstarCommunity = (slug) =>
  request(`/communities/${slug}/star`, { method: 'DELETE' });

export const joinCommunity = (slug) =>
  request(`/communities/${slug}/follow`, { method: 'POST' });

export const leaveCommunity = (slug) =>
  request(`/communities/${slug}/follow`, { method: 'DELETE' });

export const createCommunity = (data) =>
  request('/communities', { method: 'POST', body: JSON.stringify(data) });

export const listCommunities = (sort = 'popular', limit = 50, offset = 0) =>
  request(`/communities?sort=${sort}&limit=${limit}&offset=${offset}`);

// ── Feedback / Roadmap ────────────────────────────────────────────────────────
export const listFeedback = (sort = 'top', status = '', limit = 50, offset = 0) => {
  const q = new URLSearchParams({ sort, limit, offset });
  if (status) q.set('status', status);
  return request(`/feedback?${q}`);
};
export const createFeedback = (data) =>
  request('/feedback', { method: 'POST', body: JSON.stringify(data) });
export const voteFeedback = (id) =>
  request(`/feedback/${id}/vote`, { method: 'POST' });
export const unvoteFeedback = (id) =>
  request(`/feedback/${id}/vote`, { method: 'DELETE' });
export const updateFeedbackStatus = (id, status) =>
  request(`/admin/feedback/${id}`, { method: 'PATCH', body: JSON.stringify({ status }) });
export const deleteFeedback = (id) =>
  request(`/admin/feedback/${id}`, { method: 'DELETE' });

// ── Roadmap (admin) ─────────────────────────────────────────────────────────
export const listRoadmapItems = () =>
  request('/roadmap');
export const createRoadmapItem = (data) =>
  request('/admin/roadmap', { method: 'POST', body: JSON.stringify(data) });
export const updateRoadmapItem = (id, data) =>
  request(`/admin/roadmap/${id}`, { method: 'PUT', body: JSON.stringify(data) });
export const deleteRoadmapItem = (id) =>
  request(`/admin/roadmap/${id}`, { method: 'DELETE' });

// ── Banned words (admin) ────────────────────────────────────────────────────
export const listBannedWords = () =>
  request('/admin/banned-words');
export const createBannedWord = (data) =>
  request('/admin/banned-words', { method: 'POST', body: JSON.stringify(data) });
export const deleteBannedWord = (id) =>
  request(`/admin/banned-words/${id}`, { method: 'DELETE' });

// ── Posts ─────────────────────────────────────────────────────────────────────
// Response fields (NO json tags → PascalCase): ID, CommunityID, AuthorID,
//   Title, Content, Media, ViewsCount, VotesCount, IsPinned, CreatedAt

export const getCommunityPosts = (slug, offset = 0, kind = null) => {
  const q = new URLSearchParams({ limit: 20, offset });
  if (kind) q.set('kind', kind);
  return request(`/communities/${slug}/posts?${q}`);
};

export const getCommunityMembers = (slug, offset = 0) =>
  request(`/communities/${slug}/members?limit=50&offset=${offset}`);

export const getCommunityBans = (slug) =>
  request(`/communities/${slug}/bans`);

export const banCommunityUser = (slug, userId, { type = 'ban', reason = null, expiresAt = null } = {}) =>
  request(`/communities/${slug}/bans`, {
    method: 'POST',
    body: JSON.stringify({ user_id: userId, type, reason, expires_at: expiresAt }),
  });

export const unbanCommunityUser = (slug, userId) =>
  request(`/communities/${slug}/bans/${userId}`, { method: 'DELETE' });

export const getPost = (id) =>
  request(`/posts/${id}`);

// POST /api/v1/communities/:slug/posts
export const createPost = (slug, data) =>
  request(`/communities/${slug}/posts`, { method: 'POST', body: JSON.stringify(data) });

export const updatePost = (id, data) =>
  request(`/posts/${id}`, { method: 'PUT', body: JSON.stringify(data) });

export const deletePost = (id) =>
  request(`/posts/${id}`, { method: 'DELETE' });

export const votePost = (id) =>
  request(`/posts/${id}/vote`, { method: 'POST' });

export const unvotePost = (id) =>
  request(`/posts/${id}/vote`, { method: 'DELETE' });

// ── Comments ──────────────────────────────────────────────────────────────────
// Response fields (NO json tags → PascalCase): ID, PostID, AuthorID, ParentID,
//   Content, VotesCount, CreatedAt

export const getComments = (postId) =>
  request(`/posts/${postId}/comments`);

export const createComment = (postId, data) =>
  request(`/posts/${postId}/comments`, { method: 'POST', body: JSON.stringify(data) });

export const voteComment = (id) =>
  request(`/comments/${id}/vote`, { method: 'POST' });

export const unvoteComment = (id) =>
  request(`/comments/${id}/vote`, { method: 'DELETE' });

export const deleteComment = (id) =>
  request(`/comments/${id}`, { method: 'DELETE' });

// ── Feed ──────────────────────────────────────────────────────────────────────
// Response fields (json tags): type, id, url, summary, source, post_id,
//   community_slug, author, comments_count, title, cover_image, tags,
//   published_at, upvotes

export const getFeed = (cursor = '', limit = 20) => {
  const q = new URLSearchParams({ limit });
  if (cursor) q.set('cursor', cursor);
  return request(`/feed?${q}`);
};

// ── Notifications ─────────────────────────────────────────────────────────────
export const getNotifications = (offset = 0) =>
  request(`/notifications?limit=20&offset=${offset}`);

export const markNotificationRead = (id) =>
  request(`/notifications/${id}/read`, { method: 'PUT' });

export const markAllNotificationsRead = () =>
  request('/notifications/read-all', { method: 'POST' });

export const getNotificationUnreadCount = () =>
  request('/notifications/unread-count');

// ── Users (additional) ───────────────────────────────────────────────────────
export const getFollowers = (username, offset = 0) =>
  request(`/users/${username}/followers?limit=20&offset=${offset}`);

export const getFollowing = (username, offset = 0) =>
  request(`/users/${username}/following?limit=20&offset=${offset}`);

export const getFriends = () =>
  request('/users/me/friends');

// ── Bookmarks ────────────────────────────────────────────────────────────────
export const bookmarkPost = (id) =>
  request(`/posts/${id}/bookmark`, { method: 'POST' });

export const unbookmarkPost = (id) =>
  request(`/posts/${id}/bookmark`, { method: 'DELETE' });

export const getBookmarks = (offset = 0) =>
  request(`/users/me/bookmarks?limit=20&offset=${offset}`);

export const pinPost = (id) =>
  request(`/posts/${id}/pin`, { method: 'POST' });
export const unpinPost = (id) =>
  request(`/posts/${id}/pin`, { method: 'DELETE' });

// ── Badges ───────────────────────────────────────────────────────────────────
export const getBadgeDefinitions = () =>
  request('/badges');

export const getUserBadges = (username) =>
  request(`/users/${username}/badges`);

// ── Community moderators ─────────────────────────────────────────────────────
export const getCommunityModerators = (slug) =>
  request(`/communities/${slug}/moderators`);

export const addCommunityModerator = (slug, username) =>
  request(`/communities/${slug}/moderators/${username}`, { method: 'POST' });

export const removeCommunityModerator = (slug, username) =>
  request(`/communities/${slug}/moderators/${username}`, { method: 'DELETE' });

export const updateCommunity = (slug, data) =>
  request(`/communities/${slug}`, { method: 'PUT', body: JSON.stringify(data) });

// ── Notification preferences ─────────────────────────────────────────────────
export const getNotificationPreferences = () =>
  request('/users/me/notification-preferences');

export const setNotificationPreference = (type, enabled) =>
  request('/users/me/notification-preferences', { method: 'PUT', body: JSON.stringify({ type, enabled }) });

// ── Online ───────────────────────────────────────────────────────────────────
export const heartbeat = () =>
  request('/users/me/heartbeat', { method: 'POST' });

export const getOnlineCount = () =>
  request('/online/count');

export const isUserOnline = (username) =>
  request(`/users/${username}/online`);

// ── Trending ─────────────────────────────────────────────────────────────────
export const getTrending = (limit = 5) =>
  request(`/trending?limit=${limit}`);

// ── Search ───────────────────────────────────────────────────────────────────
export const searchPosts = (q, offset = 0) =>
  request(`/search/posts?q=${encodeURIComponent(q)}&limit=20&offset=${offset}`)
    .then(r => r?.hits ?? []);

export const searchCommunities = (q, offset = 0) =>
  request(`/search/communities?q=${encodeURIComponent(q)}&limit=20&offset=${offset}`)
    .then(r => r?.hits ?? []);

export const searchUsers = (q) =>
  request(`/search/users?q=${encodeURIComponent(q)}&limit=20`);

// ── Upload ───────────────────────────────────────────────────────────────────
export async function uploadFile(file) {
  const fd = new FormData();
  fd.append('file', file);
  const headers = {};
  if (DEV_AUTH_ID) headers['X-Dev-Auth-ID'] = DEV_AUTH_ID;
  else if (_token) headers['Authorization'] = `Bearer ${_token}`;
  const res = await fetch(`${BASE}/upload`, { method: 'POST', body: fd, headers });
  if (res.status === 401) { handleUnauthorized(); throw new Error('unauthorized'); }
  const text = await res.text();
  const json = text ? JSON.parse(text) : {};
  if (!res.ok || json.error) throw new Error(json?.error?.message ?? json?.error ?? `HTTP ${res.status}`);
  return (json.data ?? json).url;
}

// ── Wakapi ───────────────────────────────────────────────────────────────────
export const connectWakapi = (data) =>
  request('/users/me/wakapi', { method: 'POST', body: JSON.stringify(data) });

export const disconnectWakapi = () =>
  request('/users/me/wakapi', { method: 'DELETE' });

export const getUserWakapi = (username) =>
  request(`/users/${username}/wakapi`);

// ── Git-интеграции (GitHub / GitLab / Codeberg) ──────────────────────────────
// GitAccount: { id, userId, provider, username, instanceUrl, createdAt, updatedAt }
// PinnedRepo: { id, userId, gitAccountId, externalId, name, description, url,
//               language, starsCount, forksCount, isFork, topics, sortOrder, syncedAt }

export const listIntegrations = () =>
  request('/me/integrations');

// returns { authUrl } — нужно сделать window.location.assign(authUrl)
export const beginIntegrationConnect = (provider) =>
  request(`/integrations/${provider}/connect`, { method: 'POST' });

export const disconnectIntegration = (provider) =>
  request(`/me/integrations/${provider}`, { method: 'DELETE' });

// Подтянуть свежий список репов с провайдера (только подключённого).
export const listIntegrationRepos = (provider) =>
  request(`/integrations/${provider}/repos`);

// Полностью перезаписать набор пинов: pins = [{ provider, externalId }, ...]
export const setPinnedRepos = (pins) =>
  request('/me/pinned-repos', { method: 'PUT', body: JSON.stringify({ pins }) });

// Публичные пины пользователя для отображения на профиле.
export const getUserPinnedRepos = (username) =>
  request(`/users/${username}/repos`);

// ── Chat ──────────────────────────────────────────────────────────────────────
// ChatMessage: { id, authorId, author: { id, username, displayName, avatarUrl, reputation }, content, createdAt }
export const getChatMessages = (limit = 50) =>
  request(`/chat/messages?limit=${limit}`);

export const sendChatMessage = (content) =>
  request('/chat/messages', { method: 'POST', body: JSON.stringify({ content }) });

export function openChatSocket() {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const url = `${proto}//${window.location.host}/api/v1/chat/ws`;
  return _token ? new WebSocket(url, ['bearer', _token]) : new WebSocket(url);
}

// Admin
export const adminBanUser = (username) =>
  request(`/admin/users/${username}/ban`, { method: 'POST' });
export const adminUnbanUser = (username) =>
  request(`/admin/users/${username}/ban`, { method: 'DELETE' });
export const adminDeletePost = (id) =>
  request(`/admin/posts/${id}`, { method: 'DELETE' });
export const adminDeleteComment = (id) =>
  request(`/admin/comments/${id}`, { method: 'DELETE' });
export const adminDeleteCommunity = (slug) =>
  request(`/admin/communities/${slug}`, { method: 'DELETE' });
