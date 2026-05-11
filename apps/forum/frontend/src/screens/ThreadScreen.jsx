import { useState, useMemo, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  IconArrowLeft, IconArrowUp, IconArrowDown, IconChevronUp, IconChevronDown,
  IconMessage, IconShare3, IconBookmark, IconEyeOff, IconDots,
  IconRss, IconExternalLink, IconPin,
} from '@tabler/icons-react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { qk } from '../lib/queryKeys';
import {
  useVotePost, useBookmarkPost, useVoteComment, useFollowUser,
} from '../lib/mutations';
import { useAuth } from '../AuthContext';
import { useMe } from '../App';
import {
  getPost, getComments,
  createComment, deleteComment,
  deletePost,
  getCommunity, getCommunityPosts,
  getUser, getFollowers,
  adminDeleteComment,
} from '../lib/api';
import {
  normalizePost, normalizeComment, buildCommentTree,
  relativeTime, initials, commColor, postToCard,
} from '../lib/utils';
import { renderMarkdown } from '../lib/markdown';
import { linkTarget } from '../lib/prefs';

const SORT_OPTIONS = [
  { k: 'best', l: 'лучшее' },
  { k: 'new',  l: 'новое' },
  { k: 'old',  l: 'старое' },
];

// ─── Comment ──────────────────────────────────────────────────────────────────
const Comment = ({ c, postId, depth = 0, meId, ownerId, opAuthorId, isAdmin, onAfterReply, onAfterDelete, guestPrompt, onOpenProfile }) => {
  const [collapsed, setCollapsed] = useState(false);
  const [repliesOpen, setRepliesOpen] = useState(depth === 0);
  const [replying, setReplying] = useState(false);
  const [replyText, setReplyText] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [localVoted, setLocalVoted] = useState(c.voted ?? null);
  const [localScore, setLocalScore] = useState(c.votes ?? 0);
  useEffect(() => { setLocalVoted(c.voted ?? null); setLocalScore(c.votes ?? 0); }, [c.id]);
  const voteCommentMut = useVoteComment();
  const voted = localVoted;
  const score = localScore;

  const authorName = c.author?.username ?? c.authorId?.slice?.(0, 8) ?? '?';
  const isOP = opAuthorId && c.authorId === opAuthorId;
  const canModify = meId && (meId === c.authorId || meId === ownerId);

  const handleAdminDelete = async () => {
    setMenuOpen(false);
    if (!confirm('Удалить комментарий (admin)?')) return;
    try {
      await adminDeleteComment(c.id);
      onAfterDelete?.();
    } catch (e) { alert(e.message ?? 'не удалось'); }
  };

  const handleDelete = async () => {
    setMenuOpen(false);
    if (!confirm('Удалить комментарий?')) return;
    try {
      await deleteComment(c.id);
      onAfterDelete?.();
    } catch {}
  };

  const handleVote = () => {
    if (guestPrompt) { guestPrompt(); return; }
    const next = voted !== 'up';
    setLocalVoted(next ? 'up' : null);
    setLocalScore(s => s + (next ? 1 : -1));
    voteCommentMut.mutate({ commentId: c.id, postId, next });
  };

  const handleReplyClick = () => {
    if (guestPrompt) { guestPrompt(); return; }
    setReplying(!replying);
  };

  const handleReplySubmit = async () => {
    if (!replyText.trim() || submitting) return;
    setSubmitting(true);
    try {
      await createComment(postId, { content: replyText.trim(), parentId: c.id });
      setReplyText('');
      setReplying(false);
      onAfterReply?.();
    } catch {}
    setSubmitting(false);
  };

  return (
    <div className={'comment' + (collapsed ? ' collapsed' : '')}>
      <div className="comment-side">
        <div className="comment-avatar" onClick={() => setCollapsed(true)} style={{ cursor: 'pointer' }}>
          {c.author?.avatarUrl
            ? <img src={c.author.avatarUrl} alt="" />
            : initials(authorName)}
        </div>
        <div className="comment-thread-line" onClick={() => setCollapsed(true)} title="свернуть ветку" />
      </div>
      <div className="comment-body">
        <div className="comment-meta">
          <button type="button" className="author" onClick={() => onOpenProfile?.(authorName)}>
            @{authorName}
          </button>
          {isOP && <span className="role op">op</span>}
          <span className="time">{relativeTime(c.createdAt)}</span>
          {collapsed && (
            <button
              type="button"
              className="act"
              onClick={() => setCollapsed(false)}
              style={{ marginLeft: 'auto' }}
            >
              развернуть
            </button>
          )}
        </div>
        <div className="comment-text" dangerouslySetInnerHTML={{ __html: renderMarkdown(c.content) }} />
        <div className="comment-actions">
          <button type="button" className={'vote up' + (voted === 'up' ? ' active' : '')} onClick={handleVote}>
            <IconArrowUp size={13} stroke={1.8} />
            <span className="num">{score}</span>
          </button>
          <button type="button" className="vote down" onClick={handleVote}>
            <IconArrowDown size={13} stroke={1.8} />
          </button>
          <button type="button" className="act" onClick={handleReplyClick}>
            ответить
          </button>
          <button type="button" className="act" onClick={() => {
            navigator.clipboard?.writeText(`${window.location.origin}/p/${postId}#c-${c.id}`).catch(() => {});
          }}>
            поделиться
          </button>
          <div style={{ marginLeft: 'auto', position: 'relative' }}>
            <button type="button" className="act" onClick={() => setMenuOpen(o => !o)}>···</button>
            {menuOpen && (
              <div className="post-menu">
                <button type="button" className="post-menu-item" onClick={() => {
                  navigator.clipboard?.writeText(`${window.location.origin}/p/${postId}#c-${c.id}`).catch(() => {});
                  setMenuOpen(false);
                }}>скопировать ссылку</button>
                {canModify && (
                  <button type="button" className="post-menu-item danger" onClick={handleDelete}>удалить</button>
                )}
                {isAdmin && !canModify && (
                  <button type="button" className="post-menu-item danger" onClick={handleAdminDelete}>удалить (admin)</button>
                )}
              </div>
            )}
          </div>
        </div>

        {replying && (
          <div className="reply-form">
            <textarea
              value={replyText}
              onChange={(e) => setReplyText(e.target.value)}
              placeholder={`ответить @${authorName}…`}
              autoFocus
            />
            <div className="reply-form-foot">
              <button type="button" className="btn ghost" style={{ fontSize: 12 }} onClick={() => setReplying(false)}>
                отмена
              </button>
              <button
                type="button"
                className="btn primary"
                style={{ fontSize: 12 }}
                onClick={handleReplySubmit}
                disabled={submitting || !replyText.trim()}
              >
                {submitting ? 'отправка…' : 'ответить'}
              </button>
            </div>
          </div>
        )}

        {c.replies?.length > 0 && (
          <>
            <button
              type="button"
              className="replies-toggle"
              onClick={() => setRepliesOpen(o => !o)}
            >
              {repliesOpen ? '▾' : '▸'} {repliesOpen ? 'скрыть' : `показать ${c.replies.length}`} {c.replies.length === 1 ? 'ответ' : 'ответов'}
            </button>
            {repliesOpen && (
              <div className="replies">
                {c.replies.map(r => (
                  <Comment
                    key={r.id}
                    c={r}
                    postId={postId}
                    depth={depth + 1}
                    meId={meId}
                    ownerId={ownerId}
                    opAuthorId={opAuthorId}
                    isAdmin={isAdmin}
                    onAfterReply={onAfterReply}
                    onAfterDelete={onAfterDelete}
                    guestPrompt={guestPrompt}
                    onOpenProfile={onOpenProfile}
                  />
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

// ─── Left rail ───────────────────────────────────────────────────────────────
const LeftRail = ({ post, totalComments, onBack }) => {
  return (
    <aside className="rail" style={{ position: 'sticky', top: 96, alignSelf: 'start' }}>
      <button type="button" className="thread-back" onClick={onBack}>
        <IconArrowLeft size={13} stroke={1.8} />
        {post?.communitySlug ? `назад в hg/${post.communitySlug}` : 'назад'}
      </button>
      <div className="rail-section">
        <div className="rail-title">Этот пост</div>
        <button
          type="button"
          className="rail-link"
          onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
        >
          <IconChevronUp size={15} stroke={1.7} />
          К началу
        </button>
        <a className="rail-link" href="#comments">
          <IconMessage size={15} stroke={1.7} />
          Комменты <span className="count">{totalComments}</span>
        </a>
      </div>
    </aside>
  );
};

// ─── Right rail ──────────────────────────────────────────────────────────────
const RightRail = ({ post, author: authorOverride, community, relatedPosts, onOpenCommunity, onOpenPost, onOpenProfile, onFollow, isFollowing }) => {
  const author = authorOverride || post.author;
  const authorName = author?.username ?? post.authorId?.slice?.(0, 8) ?? '?';
  return (
    <aside className="right">
      {author && (
        <div className="widget user-card" style={{ padding: 0 }}>
          <div
            className="user-banner"
            style={author.bannerUrl ? { backgroundImage: `url(${author.bannerUrl})`, backgroundSize: 'cover', backgroundPosition: 'center' } : undefined}
          >
            <div className="user-card-avatar">
              {author.avatarUrl
                ? <img src={author.avatarUrl} alt="" style={{ width: '100%', height: '100%', borderRadius: '50%', objectFit: 'cover' }} />
                : initials(authorName)}
            </div>
          </div>
          <div className="user-body">
            <div className="user-name">{author.displayName || authorName}</div>
            <div className="user-handle">@{authorName} · автор</div>
            {author.bio && (
              <p style={{ margin: '10px 0 12px', fontSize: 12.5, color: 'var(--text-mid)', lineHeight: 1.5 }}>
                {author.bio}
              </p>
            )}
            <div className="user-stats">
              <div className="user-stat">
                <div className="v">{(author.reputation ?? 0).toLocaleString('ru')}</div>
                <div className="l">Rep</div>
              </div>
              <div className="user-stat">
                <div className="v">{author.postsCount ?? 0}</div>
                <div className="l">Posts</div>
              </div>
              <div className="user-stat">
                <div className="v">{(author.followersCount ?? 0).toLocaleString('ru')}</div>
                <div className="l">Followers</div>
              </div>
            </div>
            <div style={{ display: 'flex', gap: 6 }}>
              <button
                type="button"
                className="btn"
                style={{ flex: 1, justifyContent: 'center', fontSize: 12.5 }}
                onClick={() => onOpenProfile?.(authorName)}
              >
                Профиль
              </button>
              {onFollow && (
                <button
                  type="button"
                  className={'btn' + (isFollowing ? '' : ' primary')}
                  style={{ flex: 1, justifyContent: 'center', fontSize: 12.5 }}
                  onClick={() => onFollow(authorName)}
                >
                  {isFollowing ? 'отписаться' : '+ подписаться'}
                </button>
              )}
            </div>
          </div>
        </div>
      )}

      {community && (
        <div className="widget user-card" style={{ padding: 0 }}>
          <div
            className="user-banner"
            style={community.bannerUrl ? { backgroundImage: `url(${community.bannerUrl})`, backgroundSize: 'cover', backgroundPosition: 'center' } : { background: `linear-gradient(135deg, ${commColor(community.slug)}, var(--hn-honey-pale))` }}
          >
            <div
              className="user-card-avatar"
              style={{ background: commColor(community.slug), color: '#fff', fontFamily: 'var(--mono)', fontWeight: 700 }}
            >
              {community.avatarUrl
                ? <img src={community.avatarUrl} alt="" style={{ width: '100%', height: '100%', borderRadius: '50%', objectFit: 'cover' }} />
                : (community.slug?.[0]?.toUpperCase() ?? '?')}
            </div>
          </div>
          <div className="user-body">
            <div className="user-name">{community.name || community.slug}</div>
            <div className="user-handle">hg/{community.slug}</div>
            {community.description && (
              <p style={{ margin: '10px 0 12px', fontSize: 12.5, color: 'var(--text-mid)', lineHeight: 1.5 }}>
                {community.description}
              </p>
            )}
            <div className="user-stats">
              <div className="user-stat">
                <div className="v">{(community.followersCount ?? 0).toLocaleString('ru')}</div>
                <div className="l">Участников</div>
              </div>
              <div className="user-stat">
                <div className="v">{(community.postsCount ?? 0).toLocaleString('ru')}</div>
                <div className="l">Постов</div>
              </div>
            </div>
            <button
              type="button"
              className="btn"
              style={{ width: '100%', justifyContent: 'center', fontSize: 12.5 }}
              onClick={() => onOpenCommunity(community.slug)}
            >
              Открыть сообщество
            </button>
          </div>
        </div>
      )}

      {relatedPosts.length > 0 && (
        <div className="widget">
          <h4>Похожие посты</h4>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {relatedPosts.map(p => (
              <a
                key={p.id}
                className="trending-card"
                style={{ flex: 'auto' }}
                onClick={() => onOpenPost(p.id)}
              >
                <div className="trending-rank">hg/{p.community ?? ''}</div>
                <div className="trending-title">{p.title}</div>
                <div className="trending-meta">
                  {p.author ? `@${p.author} · ` : ''}{p.votes ?? 0} голосов · {p.comments ?? 0} комментов
                </div>
              </a>
            ))}
          </div>
        </div>
      )}
    </aside>
  );
};

// ─── Main screen ─────────────────────────────────────────────────────────────
const ThreadScreen = ({ postId }) => {
  const { isAuthenticated, login } = useAuth();
  const { me } = useMe();
  const navigate = useNavigate();

  const [sort, setSort] = useState('best');
  const [replyText, setReplyText] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');
  const [shareHint, setShareHint] = useState(false);
  const [postMenuOpen, setPostMenuOpen] = useState(false);
  const [optimisticComments, setOptimisticComments] = useState([]);
  const postMenuRef = useRef(null);
  const qc = useQueryClient();
  const votePostMut = useVotePost();
  const bookmarkMut = useBookmarkPost();
  const followMut = useFollowUser();

  const guestPrompt = isAuthenticated ? null : login;

  const postQuery = useQuery({
    queryKey: qk.post(postId),
    queryFn: () => getPost(postId),
    enabled: !!postId,
  });
  const rawPost = postQuery.data;
  const postLoading = postQuery.isLoading;
  const postError = postQuery.error?.message ?? null;

  const commentsQuery = useQuery({
    queryKey: qk.postComments(rawPost?.id),
    queryFn: () => getComments(rawPost.id),
    enabled: !!rawPost?.id,
  });
  const rawComments = commentsQuery.data;
  const reloadComments = () => qc.invalidateQueries({ queryKey: qk.postComments(rawPost?.id) });

  const post = useMemo(() => rawPost ? normalizePost(rawPost) : null, [rawPost]);

  const [votedLocal, setVotedLocal] = useState(null);
  const [votesLocal, setVotesLocal] = useState(null);
  const [bookmarkedLocal, setBookmarkedLocal] = useState(null);
  useEffect(() => { setVotedLocal(null); setVotesLocal(null); setBookmarkedLocal(null); }, [post?.id]);

  const voted = votedLocal !== null ? votedLocal : (post?.voted ?? null);
  const score = votesLocal !== null ? votesLocal : (post?.votes ?? 0);
  const bookmarked = bookmarkedLocal !== null ? bookmarkedLocal : !!post?.bookmarked;

  useEffect(() => {
    const onClick = (e) => {
      if (postMenuRef.current && !postMenuRef.current.contains(e.target)) setPostMenuOpen(false);
    };
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, []);

  const { data: community } = useQuery({
    queryKey: qk.community(post?.communitySlug),
    queryFn: () => getCommunity(post.communitySlug),
    enabled: !!post?.communitySlug,
  });

  const { data: rawRelated } = useQuery({
    queryKey: qk.communityPosts(post?.communitySlug, null),
    queryFn: () => getCommunityPosts(post.communitySlug),
    enabled: !!post?.communitySlug,
  });

  const authorUsername = post?.author?.username ?? null;
  const { data: fullAuthor } = useQuery({
    queryKey: qk.user(authorUsername),
    queryFn: () => getUser(authorUsername).catch(() => null),
    enabled: !!authorUsername,
  });

  const { data: authorFollowers } = useQuery({
    queryKey: qk.userFollowers(authorUsername),
    queryFn: () => getFollowers(authorUsername).catch(() => []),
    enabled: !!authorUsername && !!me?.id,
  });
  const isAuthorFollowing = (authorFollowers ?? []).some(u => u.id === me?.id);
  const [followedLocal, setFollowedLocal] = useState(null);
  const isFollowingAuthor = followedLocal ?? isAuthorFollowing;
  const followBusy = followMut.isPending;

  const handleAuthorFollow = () => {
    if (!isAuthenticated) { login(); return; }
    if (followBusy || !authorUsername) return;
    const next = !isFollowingAuthor;
    setFollowedLocal(next);
    followMut.mutate({ username: authorUsername, next }, {
      onError: (e) => {
        setFollowedLocal(!next);
        alert(e.message ?? 'не удалось');
      },
    });
  };

  const relatedPosts = useMemo(() => {
    if (!rawRelated || !post) return [];
    return rawRelated
      .filter(p => p.id !== post.id)
      .slice(0, 3)
      .map(postToCard);
  }, [rawRelated, post?.id]);

  const allComments = useMemo(() => {
    const flat = (rawComments ?? []).map(normalizeComment);
    const realIds = new Set(flat.map(c => c.id));
    const optimistic = optimisticComments.filter(c => !realIds.has(c.id));
    return [...flat, ...optimistic];
  }, [rawComments, optimisticComments]);

  useEffect(() => {
    if (rawComments) {
      const realIds = new Set(rawComments.map(c => c.id));
      setOptimisticComments(prev => prev.filter(c => !realIds.has(c.id)));
    }
  }, [rawComments]);

  const tree = useMemo(() => {
    const sorted = [...allComments];
    if (sort === 'new')      sorted.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
    else if (sort === 'old') sorted.sort((a, b) => new Date(a.createdAt) - new Date(b.createdAt));
    else                     sorted.sort((a, b) => (b.votes ?? 0) - (a.votes ?? 0));
    return buildCommentTree(sorted);
  }, [allComments, sort]);

  const totalComments = allComments.length;

  if (!postId) {
    return (
      <div className="thread-shell">
        <div />
        <main className="thread-col">
          <div className="empty">пост не выбран</div>
        </main>
      </div>
    );
  }

  if (postLoading) {
    return (
      <div className="thread-shell">
        <div />
        <main className="thread-col">
          <div className="empty">грузим пост…</div>
        </main>
      </div>
    );
  }

  if (postError || !post) {
    return (
      <div className="thread-shell">
        <div />
        <main className="thread-col">
          <div className="error-banner">{postError ?? 'пост не найден'}</div>
        </main>
      </div>
    );
  }

  const authorName = post.author?.username ?? post.authorId?.slice?.(0, 8) ?? '?';

  const handlePostVote = () => {
    if (guestPrompt) { guestPrompt(); return; }
    const next = voted !== 'up';
    setVotedLocal(next ? 'up' : null);
    setVotesLocal((score ?? 0) + (next ? 1 : -1));
    votePostMut.mutate({ postId: post.id, next }, {
      onError: () => { setVotedLocal(null); setVotesLocal(null); },
    });
  };

  const handleBookmark = () => {
    if (guestPrompt) { guestPrompt(); return; }
    const next = !bookmarked;
    setBookmarkedLocal(next);
    bookmarkMut.mutate({ postId: post.id, next }, {
      onError: () => setBookmarkedLocal(null),
    });
  };

  const handleShare = async () => {
    try {
      await navigator.clipboard.writeText(`${window.location.origin}/p/${postId}`);
      setShareHint(true);
      setTimeout(() => setShareHint(false), 1500);
    } catch {}
  };

  const handleSubmitTopComment = async () => {
    if (!replyText.trim() || submitting) return;
    setSubmitting(true);
    setSubmitError('');
    const text = replyText.trim();
    try {
      const created = await createComment(post.id, { content: text });
      const norm = normalizeComment(created);
      setOptimisticComments(prev => [...prev, norm]);
      setReplyText('');
      reloadComments();
    } catch (e) {
      setSubmitError(e.message ?? 'не удалось отправить комментарий');
    }
    setSubmitting(false);
  };

  const onBack = () => post.communitySlug ? navigate(`/c/${post.communitySlug}`) : navigate('/');
  const onOpenProfile = (username) => navigate(`/u/${username}`);
  const onOpenCommunity = (slug) => navigate(`/c/${slug}`);
  const onOpenPost = (id) => navigate(`/p/${id}`);

  let extHost = '';
  if (post.externalUrl) {
    try { extHost = new URL(post.externalUrl).hostname.replace(/^www\./, ''); } catch {}
  }

  return (
    <div className="thread-shell">
      <LeftRail post={post} totalComments={totalComments} onBack={onBack} />

      <main className="thread-col">
        {/* POST */}
        <article className="thread-post">
          <div className="thread-post-head">
            {post.communitySlug && (
              <button
                type="button"
                className="source-pill"
                onClick={() => onOpenCommunity(post.communitySlug)}
              >
                <span className="ico" style={{ background: commColor(post.communitySlug) }}>
                  {post.communitySlug[0]?.toUpperCase()}
                </span>
                hg/{post.communitySlug}
              </button>
            )}
            {post.isRSS && (
              <span className="source-pill" style={{ background: '#EAF1F6', borderColor: '#D6E4ED' }}>
                <IconRss size={12} stroke={1.8} />
                {post.source?.name ?? 'RSS'}
              </span>
            )}
            <span className="dot">·</span>
            {post.isRSS ? (
              <span>{post.source?.name ?? 'внешняя статья'}</span>
            ) : (
              <>
                <button type="button" className="author" onClick={() => onOpenProfile(authorName)}>
                  @{authorName}
                </button>
                <span className="role">автор</span>
              </>
            )}
            <span className="dot">·</span>
            <span>{relativeTime(post.createdAt)}</span>
            <span className="dot">·</span>
            <span>{score} голосов</span>
            <span className="dot">·</span>
            <span>{post.views} просмотров</span>
            {post.isPinned && (
              <>
                <span className="dot">·</span>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4, color: 'var(--hn-honey-dark)' }}>
                  <IconPin size={12} stroke={1.8} />закреплено
                </span>
              </>
            )}
            {post.externalUrl && (
              <a
                className="ext-link"
                href={post.externalUrl}
                target={linkTarget()}
                rel="noopener noreferrer"
                style={{ marginLeft: 'auto' }}
              >
                {extHost} <IconExternalLink size={11} stroke={1.8} />
              </a>
            )}
            <div style={{ marginLeft: post.externalUrl ? 0 : 'auto', position: 'relative' }} ref={postMenuRef}>
              <button
                type="button"
                className="icon-btn"
                title="ещё"
                style={{ width: 30, height: 30 }}
                onClick={() => setPostMenuOpen(o => !o)}
              >
                <IconDots size={14} stroke={1.8} />
              </button>
              {postMenuOpen && (
                <div className="post-menu">
                  <button type="button" className="post-menu-item" onClick={() => {
                    navigator.clipboard?.writeText(`${window.location.origin}/p/${postId}`).catch(() => {});
                    setPostMenuOpen(false);
                  }}>скопировать ссылку</button>
                  {me?.id === post.authorId && !post.isRSS && (
                    <button type="button" className="post-menu-item" onClick={() => {
                      setPostMenuOpen(false);
                      navigate(`/p/${postId}/edit`);
                    }}>редактировать</button>
                  )}
                  {me?.id === post.authorId && !post.isRSS && (
                    <button type="button" className="post-menu-item danger" onClick={async () => {
                      setPostMenuOpen(false);
                      if (!confirm('Удалить пост безвозвратно?')) return;
                      try {
                        await deletePost(post.id);
                        navigate(post.communitySlug ? `/c/${post.communitySlug}` : '/');
                      } catch {}
                    }}>удалить</button>
                  )}
                </div>
              )}
            </div>
          </div>

          <h1 className="thread-post-title">{post.title ?? post.content.slice(0, 80)}</h1>

          {post.tags?.length > 0 && (
            <div className="thread-post-tags">
              {post.tags.map(t => (
                <span key={t} className="tag"><span className="hash">#</span>{t}</span>
              ))}
            </div>
          )}

          {post.isRSS ? (
            <div className="thread-post-body">
              {post.coverImageUrl && (
                <img
                  src={post.coverImageUrl}
                  alt=""
                  style={{ width: '100%', borderRadius: 8, marginBottom: 16, objectFit: 'cover', maxHeight: 360 }}
                />
              )}
              {post.externalUrl && (
                <a
                  href={post.externalUrl}
                  target={linkTarget()}
                  rel="noopener noreferrer"
                  className="btn primary"
                  style={{ display: 'inline-flex', gap: 8, alignItems: 'center' }}
                >
                  <IconExternalLink size={14} stroke={1.8} />
                  Читать оригинал на {(() => { try { return new URL(post.externalUrl).hostname.replace(/^www\./, ''); } catch { return post.source?.name ?? 'источнике'; } })()}
                </a>
              )}
              <p style={{ marginTop: 14, fontSize: 13, color: 'var(--text-mid)', fontFamily: 'var(--sans)' }}>
                Обсуждение в комментариях ниже.
              </p>
            </div>
          ) : (
            <div
              className="thread-post-body"
              dangerouslySetInnerHTML={{ __html: renderMarkdown(post.content) }}
            />
          )}

          <div className="thread-post-foot">
            <div className={'thread-vote' + (voted === 'up' ? ' upvoted' : '')}>
              <button type="button" className={'arrow up' + (voted === 'up' ? ' active' : '')} onClick={handlePostVote}>
                <IconArrowUp size={14} stroke={1.8} />
              </button>
              <span className="num">{score}</span>
              <button type="button" className="arrow down" onClick={handlePostVote}>
                <IconArrowDown size={14} stroke={1.8} />
              </button>
            </div>
            <a href="#comments" className="thread-action">
              <IconMessage size={14} stroke={1.8} />
              {totalComments} комментов
            </a>
            <button type="button" className="thread-action" onClick={handleShare}>
              <IconShare3 size={14} stroke={1.8} />
              {shareHint ? 'скопировано' : 'поделиться'}
            </button>
            <div className="thread-foot-spacer" />
            <button type="button" className={'thread-action' + (bookmarked ? ' active' : '')} onClick={handleBookmark}>
              <IconBookmark size={14} stroke={1.8} fill={bookmarked ? 'currentColor' : 'none'} />
              {bookmarked ? 'сохранено' : 'сохранить'}
            </button>
            <button type="button" className="thread-action">
              <IconEyeOff size={14} stroke={1.8} />
              скрыть
            </button>
          </div>
        </article>

        {/* COMMENT COMPOSER */}
        {guestPrompt ? (
          <div className="comment-composer">
            <div className="comment-composer-avatar">??</div>
            <div className="comment-composer-body" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div style={{ fontSize: 13, color: 'var(--text-mid)' }}>
                войди, чтобы оставить комментарий
              </div>
              <button type="button" className="btn primary" style={{ fontSize: 12.5 }} onClick={guestPrompt}>
                Войти
              </button>
            </div>
          </div>
        ) : (
          <div className="comment-composer">
            <div className="comment-composer-avatar">
              {me?.avatarUrl
                ? <img src={me.avatarUrl} alt="" />
                : initials(me?.username ?? '?')}
            </div>
            <div className="comment-composer-body">
              <textarea
                className="comment-composer-input"
                value={replyText}
                onChange={(e) => setReplyText(e.target.value)}
                placeholder="оставь свой комментарий…"
                onKeyDown={(e) => {
                  if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
                    e.preventDefault();
                    handleSubmitTopComment();
                  }
                }}
              />
              {submitError && (
                <div className="error-banner" style={{ margin: '6px 0', fontSize: 12.5, padding: '6px 10px' }}>
                  {submitError}
                </div>
              )}
              <div className="comment-composer-foot">
                <span className="hint">markdown ок · <kbd>⌘ ↵</kbd> отправить</span>
                <button
                  type="button"
                  className="btn ghost"
                  style={{ fontSize: 12.5 }}
                  onClick={() => { setReplyText(''); setSubmitError(''); }}
                  disabled={!replyText}
                >
                  очистить
                </button>
                <button
                  type="button"
                  className="btn primary"
                  style={{ fontSize: 12.5 }}
                  onClick={handleSubmitTopComment}
                  disabled={submitting || !replyText.trim()}
                >
                  {submitting ? 'отправка…' : 'отправить'}
                </button>
              </div>
            </div>
          </div>
        )}

        <div className="comments-head" id="comments">
          <div className="count">
            <span className="num">{totalComments}</span>{' '}
            {totalComments === 1 ? 'коммент' : 'комментов'}
          </div>
          <div className="comments-sort">
            {SORT_OPTIONS.map(opt => (
              <button
                key={opt.k}
                type="button"
                className={'tab' + (sort === opt.k ? ' active' : '')}
                onClick={() => setSort(opt.k)}
              >
                {opt.l}
              </button>
            ))}
          </div>
        </div>

        {commentsQuery.isLoading ? (
          <div className="empty">грузим комменты…</div>
        ) : tree.length === 0 ? (
          <div className="empty">пока нет комментов — будь первым</div>
        ) : (
          <div className="comments">
            {tree.map(c => (
              <Comment
                key={c.id}
                c={c}
                postId={post.id}
                depth={0}
                meId={me?.id}
                ownerId={post.authorId}
                opAuthorId={post.authorId}
                isAdmin={!!me?.isAdmin}
                onAfterReply={reloadComments}
                onAfterDelete={reloadComments}
                guestPrompt={guestPrompt}
                onOpenProfile={onOpenProfile}
              />
            ))}
          </div>
        )}
      </main>

      <RightRail
        post={post}
        author={fullAuthor || post.author}
        community={community}
        relatedPosts={relatedPosts}
        onOpenCommunity={onOpenCommunity}
        onOpenPost={onOpenPost}
        onOpenProfile={onOpenProfile}
        onFollow={isAuthenticated && me?.id !== post.authorId ? handleAuthorFollow : null}
        isFollowing={isFollowingAuthor}
      />
    </div>
  );
};

export default ThreadScreen;
