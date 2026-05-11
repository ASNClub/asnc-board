import { useState, useEffect, useMemo } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  IconMenu2, IconFlame, IconClock, IconBookmark,
  IconPlus, IconFile, IconHelpCircle, IconMessage,
  IconRefresh,
} from '@tabler/icons-react';
import { useQuery } from '@tanstack/react-query';
import { qk } from '../lib/queryKeys';
import { useAuth } from '../AuthContext';
import { useMe } from '../App';
import PostCard from '../components/PostCard';
import ChatDock from '../components/ChatDock';
import {
  getFeed, getTrending, getBookmarks, getMyFollowedCommunities,
  getOnlineCount, heartbeat,
} from '../lib/api';
import { feedItemToPost, postToCard, commColor, initials, relativeTime } from '../lib/utils';
import { getHiddenTags, getShowRSS } from '../lib/prefs';

const DENSITY_KEY = 'hg.density';
const MODE_KEY    = 'hg.feedMode';

// ─── Left rail ────────────────────────────────────────────────────────────────
const LeftRail = ({ mode, onMode, communities, isAuthenticated, trendingTags, meId }) => {
  const navigate = useNavigate();
  const items = [
    { k: 'foryou',   l: 'Для меня',     icon: IconMenu2 },
    { k: 'trending', l: 'Популярное',   icon: IconFlame },
    { k: 'new',      l: 'Свежее',       icon: IconClock },
    ...(isAuthenticated ? [{ k: 'saved', l: 'Сохранённые', icon: IconBookmark }] : []),
  ];

  return (
    <aside className="rail">
      <div className="rail-section">
        {items.map(it => {
          const Ic = it.icon;
          return (
            <button
              key={it.k}
              type="button"
              className={'rail-link' + (mode === it.k ? ' active' : '')}
              onClick={() => onMode(it.k)}
            >
              <Ic size={15} stroke={1.7} />
              {it.l}
            </button>
          );
        })}
      </div>

      {communities.length > 0 && (
        <div className="rail-section">
          <div className="rail-title">Мои сообщества</div>
          {communities.map(c => (
            <Link
              key={c.slug}
              className="rail-comm"
              to={`/c/${c.slug}`}
            >
              <div className="rail-comm-icon" style={{ background: commColor(c.slug) }}>
                {c.slug?.[0]?.toUpperCase() ?? '?'}
              </div>
              <span className="rail-comm-name">hg/{c.slug}</span>
            </Link>
          ))}
          <Link
            className="rail-link"
            style={{ color: 'var(--text-dim)', fontSize: 12.5, padding: '6px 10px' }}
            to="/c"
          >
            <IconPlus size={14} stroke={1.8} />
            Найти ещё
          </Link>
        </div>
      )}

      <ChatDock meId={meId} />

      {trendingTags.length > 0 && (
        <div className="rail-section">
          <div className="rail-title">Тренды</div>
          <div style={{ padding: '0 8px' }}>
            {trendingTags.map(t => (
              <span key={t} className="rail-tag">
                <span className="hash">#</span>{t}
              </span>
            ))}
          </div>
        </div>
      )}
    </aside>
  );
};

// ─── Composer ─────────────────────────────────────────────────────────────────
const Composer = ({ me, onLogin }) => {
  const navigate = useNavigate();
  if (!me) {
    return (
      <div className="composer">
        <div className="composer-avatar">?</div>
        <button type="button" className="composer-input" onClick={onLogin}>
          войди, чтобы создать пост
        </button>
      </div>
    );
  }
  const goNew = (type) => navigate(type ? `/submit?type=${type}` : '/submit');
  return (
    <div className="composer">
      <div className="composer-avatar">
        {me.avatarUrl
          ? <img src={me.avatarUrl} alt="" style={{ width:'100%', height:'100%', borderRadius:'50%', objectFit:'cover' }} />
          : initials(me.username)}
      </div>
      <button type="button" className="composer-input" onClick={() => goNew()}>
        что посадим сегодня в сад?
      </button>
      <div className="composer-types">
        <button type="button" className="composer-type" onClick={() => goNew('discussion')}>
          <IconMessage size={14} stroke={1.7} />Обсуждение
        </button>
        <button type="button" className="composer-type" onClick={() => goNew('article')}>
          <IconFile size={14} stroke={1.7} />Статья
        </button>
        <button type="button" className="composer-type" onClick={() => goNew('question')}>
          <IconHelpCircle size={14} stroke={1.7} />Вопрос
        </button>
      </div>
    </div>
  );
};

// ─── Trending strip ───────────────────────────────────────────────────────────
const TrendingStrip = ({ posts }) => {
  if (!posts || posts.length === 0) return null;
  const navigate = useNavigate();
  return (
    <div>
      <div className="strip-head">
        <IconFlame className="flame" size={14} stroke={1.7} />
        сейчас обсуждают
      </div>
      <div style={{ height: 6 }} />
      <div className="trending-strip">
        {posts.slice(0, 4).map((p, i) => (
          <div
            key={p.id}
            className="trending-card"
            onClick={() => navigate(`/p/${p.shortId || p.id}`)}
          >
            <div className="trending-rank">#{i + 1}</div>
            <div className="trending-title">{p.title ?? p.content?.slice(0, 80)}</div>
            <div className="trending-meta">
              {p.communitySlug ? `hg/${p.communitySlug}` : ''}
              {p.commentsCount != null ? ` · ${p.commentsCount} комментов` : ''}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// ─── Right rail ───────────────────────────────────────────────────────────────
const RightRail = ({ me, online, trendingTagDeltas, onLogin }) => {
  const navigate = useNavigate();
  return (
    <aside className="right">
      {me ? (
        <div className="widget user-card">
          <div
            className="user-banner"
            style={me.bannerUrl ? { backgroundImage: `url(${me.bannerUrl})`, backgroundSize: 'cover', backgroundPosition: 'center' } : undefined}
          >
            <div className="user-card-avatar">
              {me.avatarUrl
                ? <img src={me.avatarUrl} alt="" style={{ width:'100%', height:'100%', borderRadius:'50%', objectFit:'cover' }} />
                : initials(me.username)}
            </div>
          </div>
          <div className="user-body">
            <div className="user-name">{me.displayName || me.username}</div>
            <div className="user-handle">@{me.username}</div>
            <div className="user-stats">
              <div className="user-stat">
                <div className="v">{(me.reputation ?? 0).toLocaleString('ru')}</div>
                <div className="l">Rep</div>
              </div>
              <div className="user-stat">
                <div className="v">{me.postsCount ?? 0}</div>
                <div className="l">Posts</div>
              </div>
              <div className="user-stat">
                <div className="v">{(me.followersCount ?? 0).toLocaleString('ru')}</div>
                <div className="l">Followers</div>
              </div>
            </div>
            <button
              type="button"
              className="btn"
              style={{ width:'100%', justifyContent:'center' }}
              onClick={() => navigate('/me')}
            >
              Открыть профиль
            </button>
          </div>
        </div>
      ) : (
        <div className="widget user-card">
          <div className="user-banner" />
          <div className="user-body">
            <div className="user-name">Добро пожаловать</div>
            <div className="user-handle">гость</div>
            <p style={{ margin:'10px 0', fontSize:12.5, color:'var(--text-mid)', lineHeight:1.5 }}>
              войди, чтобы подписываться на сообщества, оставлять посты и собирать значки.
            </p>
            <button
              type="button"
              className="btn primary"
              style={{ width:'100%', justifyContent:'center' }}
              onClick={onLogin}
            >
              Войти
            </button>
          </div>
        </div>
      )}

      {online != null && (
        <div className="widget">
          <h4>В саду сейчас</h4>
          <div className="online-count">
            <span className="pulse" />
            <span><b style={{ color:'var(--text)' }}>{online}</b> онлайн</span>
          </div>
        </div>
      )}

      {trendingTagDeltas.length > 0 && (
        <div className="widget">
          <h4>Тренды недели</h4>
          {trendingTagDeltas.map((t, i) => (
            <div key={t.tag} className="tag-row">
              <a className="name">
                <span className="rank">{i + 1}.</span>#{t.tag}
              </a>
              <span className="delta">×{t.count}</span>
            </div>
          ))}
        </div>
      )}

      <div className="widget" style={{ padding:'12px 14px' }}>
        <div className="footer-meta">HoneyGarden · v0.10.0 · grown locally</div>
      </div>
    </aside>
  );
};

// ─── Main screen ──────────────────────────────────────────────────────────────
const FeedScreen = () => {
  const { isAuthenticated, login } = useAuth();
  const { me } = useMe();
  const navigate = useNavigate();

  const [mode, setMode] = useState(() => localStorage.getItem(MODE_KEY) || 'foryou');
  const [density] = useState(() => localStorage.getItem(DENSITY_KEY) || 'cards');
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => { localStorage.setItem(MODE_KEY, mode); }, [mode]);

  // Heartbeat — стучимся в API раз в 4 минуты, чтобы оставаться "онлайн"
  useEffect(() => {
    if (!isAuthenticated) return;
    heartbeat().catch(() => {});
    const id = setInterval(() => heartbeat().catch(() => {}), 4 * 60 * 1000);
    return () => clearInterval(id);
  }, [isAuthenticated]);

  // Список постов в зависимости от источника (mode)
  const feedQuery = useQuery({
    queryKey: qk.feed(mode),
    queryFn: () => {
      if (mode === 'trending') return getTrending(20);
      if (mode === 'saved')    return getBookmarks();
      return getFeed();
    },
  });
  const rawData = feedQuery.data;
  const loading = feedQuery.isLoading;
  const error = feedQuery.error?.message ?? null;
  const refetch = feedQuery.refetch;

  const posts = useMemo(() => {
    const isFeedItem = mode === 'foryou' || mode === 'new';
    const dateOf = (p) => new Date(isFeedItem ? p.publishedAt : p.createdAt).getTime();
    let list = [...(rawData ?? [])];
    if (mode === 'new') list.sort((a, b) => dateOf(b) - dateOf(a));
    const cards = list.map(item => isFeedItem ? feedItemToPost(item) : postToCard(item));
    const hidden = new Set(getHiddenTags());
    const showRSS = getShowRSS();
    return cards.filter(p => {
      if (!showRSS && p.isRSS) return false;
      if (hidden.size && (p.tags ?? []).some(t => hidden.has(String(t).toLowerCase()))) return false;
      return true;
    });
  }, [rawData, mode]);

  // Трендовые посты для верхней полосы и для деривации тегов
  const { data: trendingRaw } = useQuery({
    queryKey: qk.trending(20),
    queryFn: () => getTrending(20),
  });
  const trendingPosts = trendingRaw ?? [];

  // Топ тегов из трендов (не моки — реальные данные)
  const trendingTagDeltas = useMemo(() => {
    const counts = new Map();
    for (const p of trendingPosts) {
      for (const t of (p.tags ?? [])) counts.set(t, (counts.get(t) ?? 0) + 1);
    }
    return [...counts.entries()]
      .sort((a, b) => b[1] - a[1])
      .slice(0, 6)
      .map(([tag, count]) => ({ tag, count }));
  }, [trendingPosts]);

  const trendingRailTags = trendingTagDeltas.map(t => t.tag);

  // Левый рейл — мои сообщества
  const { data: myCommsRaw } = useQuery({
    queryKey: qk.myCommunities(),
    queryFn: getMyFollowedCommunities,
    enabled: isAuthenticated,
  });
  const myComms = (myCommsRaw ?? []).map(c => ({
    slug: c.slug ?? c.Slug,
    name: c.name ?? c.Name,
  }));

  // Онлайн
  const { data: onlineData } = useQuery({
    queryKey: qk.online(),
    queryFn: getOnlineCount,
    staleTime: 5_000,
    refetchInterval: 30_000,
  });
  const onlineCount = onlineData?.count ?? null;

  const openThread = (id) => navigate(`/p/${id}`);

  return (
    <div className="shell">
      <LeftRail
        mode={mode}
        onMode={setMode}
        communities={myComms}
        isAuthenticated={isAuthenticated}
        trendingTags={trendingRailTags}
        meId={me?.id ?? null}
      />

      <main className="feed-col">
        <div className="greeting">
          <div>
            <h1>Сегодня в <span className="amp">саду</span></h1>
            <div className="greeting-sub">
              {posts.length > 0 ? `${posts.length} постов в ленте` : 'свежих постов пока нет'}
            </div>
          </div>
          <button
            type="button"
            className={'btn ghost feed-refresh' + (refreshing ? ' refreshing' : '')}
            style={{ fontFamily:'var(--mono)', fontSize:11 }}
            onClick={async () => {
              setRefreshing(true);
              try { await refetch(); } finally { setTimeout(() => setRefreshing(false), 400); }
            }}
            title="Обновить ленту"
          >
            <IconRefresh size={13} stroke={1.7} className={refreshing ? 'spin' : ''} />обновить
          </button>
        </div>

        <Composer me={me} onLogin={login} />

        {loading && posts.length === 0 ? (
          <div className="empty">грузим ленту…</div>
        ) : error ? (
          <div className="error-banner">не получилось загрузить ленту</div>
        ) : posts.length === 0 ? (
          <div className="empty">
            {mode === 'saved'    ? 'сохранённых постов пока нет'
             : mode === 'trending' ? 'пока ничего не горит'
             : 'пока пусто — подпишись на сообщества, чтобы увидеть ленту'}
          </div>
        ) : (
          <>
            {posts.slice(0, 2).map(p => (
              <PostCard key={p.id} post={p} density={density} onOpen={openThread} />
            ))}
            {trendingPosts.length > 0 && <TrendingStrip posts={trendingPosts} />}
            {posts.slice(2).map(p => (
              <PostCard key={p.id} post={p} density={density} onOpen={openThread} />
            ))}
            <div className="feed-end">— конец ленты —</div>
          </>
        )}
      </main>

      <RightRail
        me={me}
        online={onlineCount}
        trendingTagDeltas={trendingTagDeltas}
        onLogin={login}
      />
    </div>
  );
};

export default FeedScreen;
