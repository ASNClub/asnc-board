// Шаренные мутации с optimistic-апдейтом и cache invalidation.
// Все хуки расчитаны на единый QueryClient (см. main.jsx).
import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  votePost, unvotePost,
  voteComment, unvoteComment,
  bookmarkPost, unbookmarkPost,
  followUser, unfollowUser,
  blockUser, unblockUser,
  joinCommunity, leaveCommunity,
  starCommunity, unstarCommunity,
  markNotificationRead, markAllNotificationsRead,
  voteFeedback, unvoteFeedback, createFeedback,
} from './api';
import { qk } from './queryKeys';

// Патчер: пробегает по всем кешам в QueryClient и применяет updater к постам с нужным id.
// Обновляет любую структуру: массив постов, объект с .items[], одиночный пост, объект {posts:[]}.
function patchPostInCache(qc, postId, patch) {
  qc.getQueryCache().getAll().forEach((q) => {
    qc.setQueryData(q.queryKey, (old) => mapPosts(old, postId, patch));
  });
}

function mapPosts(node, postId, patch) {
  if (!node) return node;
  if (Array.isArray(node)) {
    let changed = false;
    const next = node.map((item) => {
      const m = mapPosts(item, postId, patch);
      if (m !== item) changed = true;
      return m;
    });
    return changed ? next : node;
  }
  if (typeof node === 'object') {
    if (node.id === postId && ('votes' in node || 'title' in node)) {
      return { ...node, ...patch(node) };
    }
    let changed = false;
    const next = {};
    for (const k of Object.keys(node)) {
      const v = node[k];
      const nv = (Array.isArray(v) || (v && typeof v === 'object' && !(v instanceof Date)))
        ? mapPosts(v, postId, patch) : v;
      if (nv !== v) changed = true;
      next[k] = nv;
    }
    return changed ? next : node;
  }
  return node;
}

// ─── Vote on post ────────────────────────────────────────────────────────
export function useVotePost() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ postId, next }) => {
      if (next) await votePost(postId);
      else      await unvotePost(postId);
    },
    onMutate: async ({ postId, next }) => {
      const prev = snapshot(qc);
      patchPostInCache(qc, postId, (p) => ({
        voted: next ? 'up' : null,
        votes: (p.votes ?? 0) + (next ? (p.voted === 'up' ? 0 : 1) : (p.voted === 'up' ? -1 : 0)),
      }));
      return { prev };
    },
    onError: (_e, _v, ctx) => restore(qc, ctx?.prev),
    onSettled: (_d, _e, { postId }) => {
      qc.invalidateQueries({ queryKey: qk.post(postId) });
    },
  });
}

// ─── Bookmark on post ────────────────────────────────────────────────────
export function useBookmarkPost() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ postId, next }) => {
      if (next) await bookmarkPost(postId);
      else      await unbookmarkPost(postId);
    },
    onMutate: async ({ postId, next }) => {
      const prev = snapshot(qc);
      patchPostInCache(qc, postId, () => ({ bookmarked: next }));
      return { prev };
    },
    onError: (_e, _v, ctx) => restore(qc, ctx?.prev),
    onSettled: () => qc.invalidateQueries({ queryKey: qk.bookmarks() }),
  });
}

// ─── Vote on comment ─────────────────────────────────────────────────────
export function useVoteComment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ commentId, next }) => {
      if (next) await voteComment(commentId);
      else      await unvoteComment(commentId);
    },
    onSettled: (_d, _e, { postId }) => {
      if (postId) qc.invalidateQueries({ queryKey: qk.postComments(postId) });
    },
  });
}

// ─── Follow / unfollow user ──────────────────────────────────────────────
export function useFollowUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ username, next }) => {
      if (next) await followUser(username);
      else      await unfollowUser(username);
    },
    onSettled: (_d, _e, { username }) => {
      qc.invalidateQueries({ queryKey: qk.user(username) });
      qc.invalidateQueries({ queryKey: qk.userFollowers(username) });
    },
  });
}

// ─── Block / unblock user ────────────────────────────────────────────────
export function useBlockUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ username, next }) => {
      if (next) await blockUser(username);
      else      await unblockUser(username);
    },
    onSettled: (_d, _e, { username }) => {
      qc.invalidateQueries({ queryKey: qk.user(username) });
      qc.invalidateQueries({ queryKey: qk.myBlocks() });
      qc.invalidateQueries({ queryKey: ['feed'] });
      qc.invalidateQueries({ queryKey: ['trending'] });
      qc.invalidateQueries({ queryKey: ['post'] });
    },
  });
}

// ─── Join / leave community ──────────────────────────────────────────────
export function useJoinCommunity() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ slug, next }) => {
      if (next) await joinCommunity(slug);
      else      await leaveCommunity(slug);
    },
    onSettled: (_d, _e, { slug }) => {
      qc.invalidateQueries({ queryKey: qk.community(slug) });
      qc.invalidateQueries({ queryKey: qk.myCommunities() });
    },
  });
}

// ─── Star / unstar community ─────────────────────────────────────────────
export function useStarCommunity() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ slug, next }) => {
      if (next) await starCommunity(slug);
      else      await unstarCommunity(slug);
    },
    onSettled: (_d, _e, { slug }) => {
      qc.invalidateQueries({ queryKey: qk.community(slug) });
    },
  });
}

// ─── Mark notification read ──────────────────────────────────────────────
export function useMarkNotificationRead() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id }) => markNotificationRead(id),
    onSettled: () => qc.invalidateQueries({ queryKey: qk.notifications() }),
  });
}

export function useMarkAllNotificationsRead() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => markAllNotificationsRead(),
    onSettled: () => qc.invalidateQueries({ queryKey: qk.notifications() }),
  });
}

// ─── Feedback / Roadmap ──────────────────────────────────────────────────
export function useVoteFeedback() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, next }) => {
      if (next) await voteFeedback(id);
      else      await unvoteFeedback(id);
    },
    onMutate: async ({ id, next }) => {
      const all = qc.getQueryCache().getAll().filter((q) => q.queryKey[0] === 'feedback');
      const prev = all.map((q) => [q.queryKey, q.state.data]);
      all.forEach((q) => {
        qc.setQueryData(q.queryKey, (old) => {
          if (!Array.isArray(old)) return old;
          return old.map((f) => f.id === id
            ? { ...f, isVoted: next, votesCount: f.votesCount + (next ? 1 : -1) }
            : f);
        });
      });
      return { prev };
    },
    onError: (_err, _vars, ctx) => restore(qc, ctx?.prev),
    onSettled: () => qc.invalidateQueries({ queryKey: ['feedback'] }),
  });
}

export function useCreateFeedback() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data) => createFeedback(data),
    onSettled: () => qc.invalidateQueries({ queryKey: ['feedback'] }),
  });
}

// ─── helpers ─────────────────────────────────────────────────────────────
function snapshot(qc) {
  return qc.getQueryCache().getAll().map((q) => [q.queryKey, q.state.data]);
}
function restore(qc, prev) {
  if (!prev) return;
  prev.forEach(([key, data]) => qc.setQueryData(key, data));
}
