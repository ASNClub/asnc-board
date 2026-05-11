import { useMemo, useRef, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  IconCamera, IconEdit, IconCheck, IconUserPlus, IconPencil, IconSettings,
  IconHexagon, IconPlus, IconLink, IconUsers, IconMapPin, IconCalendar,
  IconLayoutGrid, IconNote, IconCode, IconTrophy, IconActivity, IconPin,
  IconMessageCircle, IconChevronUp, IconBookmark, IconArrowsExchange,
  IconBrandGithub, IconGitBranch, IconBan, IconStar, IconGitFork,
} from '@tabler/icons-react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { qk } from '../lib/queryKeys';
import { useFollowUser, useBlockUser } from '../lib/mutations';
import { useAuth } from '../AuthContext';
import { useMe } from '../App';
import {
  getUser, getFollowers, getFollowing,
  getUserBadges, getUserActivity, getUserPosts, getUserWakapi, getBadgeDefinitions,
  getMyCommunity, uploadFile, updateMe, getUserPinnedRepos, getMyBlocks,
  pinPost, unpinPost, adminBanUser, adminUnbanUser,
} from '../lib/api';
import { relativeTime, initials, commColor } from '../lib/utils';
import ImageCropModal from '../components/ImageCropModal';
import { linkTarget } from '../lib/prefs';

// ── helpers ─────────────────────────────────────────────────────────────────
const providerIcon = (prov, size = 13) =>
  prov === 'github'
    ? <IconBrandGithub size={size} stroke={1.7} />
    : <IconGitBranch size={size} stroke={1.7} />;

const pinProvider = (p) => {
  try {
    const h = new URL(p.url).host.toLowerCase();
    if (h.includes('github')) return 'github';
    return null;
  } catch { return null; }
};

const providerProfileUrl = (pin) => {
  try {
    const u = new URL(pin.url);
    const seg = u.pathname.split('/').filter(Boolean);
    return `${u.protocol}//${u.host}/${seg[0] ?? ''}`;
  } catch { return '#'; }
};

const ACTIVITY_META = {
  post:    { color: '#E09832', label: 'опубликовал пост',    text: 'Опубликовал пост',    icon: <IconPencil size={9} stroke={2.2} /> },
  comment: { color: '#8FBCBB', label: 'прокомментировал',    text: 'Прокомментировал',    icon: <IconMessageCircle size={9} stroke={2.2} /> },
  vote:    { color: '#A3BE8C', label: 'проголосовал',        text: 'Поддержал',           icon: <IconChevronUp size={9} stroke={2.2} /> },
};

const formatDay = (iso) => {
  const d = new Date(iso);
  const today = new Date();
  const yest  = new Date(today); yest.setDate(today.getDate() - 1);
  if (d.toDateString() === today.toDateString()) return 'Сегодня';
  if (d.toDateString() === yest.toDateString())  return 'Вчера';
  return d.toLocaleDateString('ru', { day: 'numeric', month: 'long' });
};

const formatTime = (iso) =>
  new Date(iso).toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' });

// ── contribution heatmap (from real activity timestamps) ────────────────────
const buildContribCells = (activity) => {
  const counts = new Map();
  (activity ?? []).forEach((a) => {
    const d = new Date(a.createdAt);
    const key = d.toISOString().slice(0, 10);
    counts.set(key, (counts.get(key) ?? 0) + 1);
  });
  // 53 weeks × 7 days = 371 cells, ending today
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const cells = [];
  for (let i = 53 * 7 - 1; i >= 0; i--) {
    const d = new Date(today);
    d.setDate(today.getDate() - i);
    const key = d.toISOString().slice(0, 10);
    const n = counts.get(key) ?? 0;
    let l = 0;
    if (n >= 1) l = 1;
    if (n >= 3) l = 2;
    if (n >= 6) l = 3;
    if (n >= 10) l = 4;
    cells.push({ key, n, l });
  }
  // organise as columns: 53 columns × 7 rows. Grid auto-flow column.
  return cells;
};

// ── ContribHeatmap ──────────────────────────────────────────────────────────
const ContribHeatmap = ({ activity }) => {
  const cells = useMemo(() => buildContribCells(activity), [activity]);
  const total = cells.reduce((s, c) => s + c.n, 0);
  return (
    <div className="contrib">
      <div className="contrib-head">
        <h3>Активность за год</h3>
        <span className="sub">{total} {total === 1 ? 'событие' : 'событий'}</span>
      </div>
      <div className="contrib-grid" style={{ gridAutoFlow: 'column', gridTemplateRows: 'repeat(7, 11px)', gridTemplateColumns: 'repeat(53, 11px)' }}>
        {cells.map((c) => (
          <div
            key={c.key}
            className={'contrib-cell' + (c.l ? ` l${c.l}` : '')}
            title={`${c.key}: ${c.n}`}
          />
        ))}
      </div>
      <div className="contrib-legend">
        <span>меньше</span>
        <div className="scale">
          <span className="contrib-cell" />
          <span className="contrib-cell l1" />
          <span className="contrib-cell l2" />
          <span className="contrib-cell l3" />
          <span className="contrib-cell l4" />
        </div>
        <span>больше</span>
      </div>
    </div>
  );
};

// ── Activity feed ───────────────────────────────────────────────────────────
const ActivityCard = ({ activity, onOpenPost }) => {
  const [filter, setFilter] = useState('all');
  const items = (activity ?? []).filter((a) => filter === 'all' || a.type === filter);

  // group by day
  const groups = [];
  items.forEach((a) => {
    const day = formatDay(a.createdAt);
    const last = groups[groups.length - 1];
    if (last && last.day === day) last.items.push(a);
    else groups.push({ day, items: [a] });
  });

  const filters = [
    { k: 'all',     l: 'Всё' },
    { k: 'post',    l: 'Посты' },
    { k: 'comment', l: 'Комментарии' },
  ];

  return (
    <div className="activity">
      <div className="activity-head">
        <h3>Активность</h3>
        <div className="activity-filters">
          {filters.map((f) => (
            <button
              key={f.k}
              type="button"
              className={'activity-filter' + (filter === f.k ? ' active' : '')}
              onClick={() => setFilter(f.k)}
            >
              {f.l}
            </button>
          ))}
        </div>
      </div>

      {groups.length === 0 ? (
        <div className="empty">пока нет событий</div>
      ) : groups.map((g) => (
        <div key={g.day}>
          <div className="activity-day">{g.day}</div>
          {g.items.map((a, i) => {
            const meta = ACTIVITY_META[a.type] ?? ACTIVITY_META.comment;
            const ref = a.title || a.content?.slice(0, 80) || '';
            return (
              <div key={a.id ?? `${g.day}-${i}`} className="act-row">
                <div className="act-dot-col">
                  <div className="act-dot" style={{ '--ac': meta.color }}>
                    {meta.icon}
                  </div>
                  <div className="act-line" />
                </div>
                <div className="act-text">
                  <span className="who">{meta.text} </span>
                  {ref && (
                    <span
                      className="ref"
                      onClick={() => a.postId && onOpenPost?.(a.postId)}
                    >{ref}</span>
                  )}
                  {a.communitySlug && (
                    <div className="meta">
                      <span>hg/{a.communitySlug}</span>
                    </div>
                  )}
                </div>
                <div className="act-time">{formatTime(a.createdAt)}</div>
              </div>
            );
          })}
        </div>
      ))}
    </div>
  );
};

// ── Wakapi widget ───────────────────────────────────────────────────────────
const WakapiCard = ({ wakapi }) => {
  if (!wakapi?.connected || !wakapi.stats) return null;
  const stats = wakapi.stats;
  const totalH = Math.floor((stats.totalSeconds ?? 0) / 3600);
  const totalM = Math.floor(((stats.totalSeconds ?? 0) % 3600) / 60);
  const langs = (stats.languages ?? []).slice(0, 5);
  return (
    <div className="waka">
      <h4>
        активность за неделю
        <span className="live">live</span>
      </h4>
      <div className="waka-total">
        {totalH}ч {totalM}м
        <span className="sub">кода за 7 дней</span>
      </div>
      {langs.length > 0 && (
        <>
          <div className="waka-bar">
            {langs.map((l, i) => (
              <span key={i} style={{ background: l.color, width: `${l.percent}%` }} />
            ))}
          </div>
          <ul className="waka-langs">
            {langs.map((l) => (
              <li key={l.name}>
                <span className="dot" style={{ background: l.color }} />
                {l.name}
                <span className="pct">{Math.round(l.percent)}%</span>
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  );
};

// ── Friends widget (VK-style) ───────────────────────────────────────────────
const ONLINE_THRESHOLD = 5 * 60 * 1000;
const isOnline = (u) => u.lastSeenAt && (Date.now() - new Date(u.lastSeenAt).getTime()) < ONLINE_THRESHOLD;
const lastSeenLabel = (u) => {
  if (!u.lastSeenAt) return null;
  const ago = Date.now() - new Date(u.lastSeenAt).getTime();
  if (ago < ONLINE_THRESHOLD) return 'онлайн';
  const m = Math.floor(ago / 60000);
  if (m < 60) return `был ${m}м назад`;
  const h = Math.floor(m / 60);
  if (h < 24) return `был ${h}ч назад`;
  const d = Math.floor(h / 24);
  return `был ${d}д назад`;
};

const FriendsWidget = ({ friends }) => {
  if (!friends || friends.length === 0) return null;
  const sorted = [...friends].sort((a, b) => (isOnline(b) ? 1 : 0) - (isOnline(a) ? 1 : 0));
  const onlineCount = sorted.filter(isOnline).length;
  return (
    <div className="about-card">
      <h4 style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        Друзья
        <span style={{ fontSize: 10.5 }}>
          {friends.length}
          {onlineCount > 0 && <> · <span style={{ color: 'var(--hn-green)' }}>{onlineCount} онлайн</span></>}
        </span>
      </h4>
      <div className="friends-flow">
        {sorted.slice(0, 12).map((f) => (
          <Link key={f.username} className="friend-chip" to={`/u/${f.username}`} title={lastSeenLabel(f) || `@${f.username}`}>
            <div className="friend-av">
              {f.avatarUrl ? <img src={f.avatarUrl} alt="" /> : initials(f.username)}
              {isOnline(f) && <span className="on" />}
            </div>
            <span className="friend-name">{f.displayName || f.username}</span>
          </Link>
        ))}
      </div>
    </div>
  );
};

// ── Top tags widget (from user's posts) ─────────────────────────────────────
const TopTagsWidget = ({ posts }) => {
  const counts = {};
  (posts ?? []).forEach((p) => (p.tags ?? []).forEach((t) => { counts[t] = (counts[t] ?? 0) + 1; }));
  const top = Object.entries(counts).sort((a, b) => b[1] - a[1]).slice(0, 5);
  if (top.length === 0) return null;
  const max = top[0][1];
  return (
    <div className="about-card">
      <h4>топ-теги</h4>
      <div className="tags-bar-list">
        {top.map(([tag, n]) => (
          <div key={tag} className="row">
            <span className="t">#{tag}</span>
            <div className="bar"><div style={{ width: `${(n / max) * 100}%` }} /></div>
            <span className="n">{n}</span>
          </div>
        ))}
      </div>
    </div>
  );
};

// ── Badges ──────────────────────────────────────────────────────────────────
const RARITY_ORDER = ['legendary', 'epic', 'rare', 'common'];
const RARITY_COLOR = { legendary: '#F5C55A', epic: '#B48EAD', rare: '#81A1C1', common: '#8FBCBB' };

const BadgeDetailModal = ({ badge, onClose }) => (
  <div className="modal-overlay" onClick={onClose}>
    <div className="modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 420, textAlign: 'center' }}>
      <div className="badge-hex" style={{
        '--bc': badge.color || RARITY_COLOR[badge.rarity] || 'var(--hn-honey)',
        width: 96, height: 110, fontSize: 36, margin: '8px auto 14px',
      }}>{badge.glyph}</div>
      <h3 style={{ margin: '0 0 4px', fontFamily: 'var(--serif)', fontSize: 20 }}>{badge.nameRu || badge.name}</h3>
      <span className={`badge-rarity-dot ${badge.rarity}`}>{badge.rarity}</span>
      {badge.descriptionRu || badge.description ? (
        <p style={{ margin: '14px 0', fontSize: 13.5, color: 'var(--text-mid)', lineHeight: 1.55 }}>
          {badge.descriptionRu || badge.description}
        </p>
      ) : null}
      <div style={{ fontFamily: 'var(--mono)', fontSize: 11.5, color: 'var(--text-dim)' }}>
        {badge.earnedAt ? `получен ${relativeTime(badge.earnedAt)}` : 'ещё не получен'}
      </div>
      <button type="button" className="btn" style={{ marginTop: 16, justifyContent: 'center' }} onClick={onClose}>закрыть</button>
    </div>
  </div>
);

const BadgesWidget = ({ userBadges, defs, total, onAll }) => {
  const [selected, setSelected] = useState(null);
  const earned = (userBadges ?? [])
    .map((b) => {
      const def = (defs ?? []).find((d) => d.id === b.badge?.id) ?? b.badge;
      return def ? { ...def, name: def.nameRu || def.name, rarity: def.rarity || 'common', earnedAt: b.earnedAt } : null;
    })
    .filter(Boolean)
    .sort((a, b) => {
      const ai = RARITY_ORDER.indexOf(a.rarity);
      const bi = RARITY_ORDER.indexOf(b.rarity);
      return (ai === -1 ? 99 : ai) - (bi === -1 ? 99 : bi);
    });
  if (earned.length === 0) return null;
  return (
    <div className="about-card">
      <h4 style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        Витрина
        <span style={{ fontSize: 10.5 }}>{earned.length}{total ? ` / ${total}` : ''}</span>
      </h4>
      <div className="badges-showcase">
        {earned.slice(0, 6).map((b) => (
          <div key={b.id} className={`badge-card ${b.rarity}`} style={{ '--bc': b.color || 'var(--hn-honey)' }} onClick={() => setSelected(b)}>
            <div className="badge-hex" style={{ '--bc': b.color || 'var(--hn-honey)' }}>{b.glyph}</div>
            <div className="badge-name">{b.name}</div>
            <span className={`badge-rarity-dot ${b.rarity}`}>{b.rarity}</span>
          </div>
        ))}
      </div>
      <div style={{ marginTop: 10, textAlign: 'center' }}>
        <a onClick={onAll} style={{ cursor: 'pointer', fontFamily: 'var(--mono)', fontSize: 11, color: 'var(--hn-honey-dark)' }}>все значки →</a>
      </div>
      {selected && <BadgeDetailModal badge={selected} onClose={() => setSelected(null)} />}
    </div>
  );
};

// ── Posts tab list ──────────────────────────────────────────────────────────
const PostsList = ({ posts, isOwn, onPin }) => {
  if (!posts || posts.length === 0) {
    return <div className="empty">постов пока нет</div>;
  }
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      {posts.map((p) => (
        <div key={p.id} className="post-tile" style={{ display: 'flex', alignItems: 'center' }}>
          <Link
            to={`/p/${p.shortId || p.id}`}
            style={{ display: 'contents', textDecoration: 'none', color: 'inherit' }}
          >
            <div className="post-tile-comm" style={{ background: commColor(p.communitySlug) }}>
              {p.communitySlug?.[0]?.toUpperCase() ?? '?'}
            </div>
            <div className="post-tile-body">
              <div className="post-tile-meta">
                hg/{p.communitySlug ?? '?'}
                <span> · {relativeTime(p.createdAt)}</span>
                {p.isPinned && <span style={{ color: 'var(--hn-honey-dark)' }}> · 📌</span>}
              </div>
              <div className="post-tile-title">{p.title || p.content?.slice(0, 80)}</div>
            </div>
            <div className="post-tile-stats">
              <span className="stat"><IconChevronUp size={13} stroke={2} /> {p.votesCount ?? 0}</span>
              {p.commentsCount > 0 && (
                <span className="stat"><IconMessageCircle size={13} stroke={1.8} /> {p.commentsCount}</span>
              )}
            </div>
          </Link>
          {isOwn && onPin && (
            <button
              className="icon-btn"
              onClick={() => onPin(p.id, !p.isPinned)}
              title={p.isPinned ? 'Открепить' : 'Закрепить'}
              style={{ width: 28, height: 28, minWidth: 28, marginLeft: 6, color: p.isPinned ? 'var(--hn-honey-dark)' : 'var(--text-dim)' }}
            >
              <IconPin size={14} stroke={1.8} />
            </button>
          )}
        </div>
      ))}
    </div>
  );
};

// ── Repos tab grid ──────────────────────────────────────────────────────────
const ReposGrid = ({ pins, isOther, onOpenSettings }) => {
  if (!pins || pins.length === 0) {
    return (
      <div className="empty" style={{ padding: '40px 24px' }}>
        <div style={{ fontFamily: 'var(--serif)', fontSize: 16, color: 'var(--text)', marginBottom: 8 }}>
          нет закреплённых репозиториев
        </div>
        <div style={{ marginBottom: 14 }}>
          {isOther
            ? 'этот пользователь пока ничего не закрепил'
            : 'подключи GitHub в настройках'}
        </div>
        {!isOther && (
          <button className="btn primary" onClick={onOpenSettings}>
            <IconSettings size={13} stroke={1.8} /> Открыть интеграции
          </button>
        )}
      </div>
    );
  }
  return (
    <div className="repos-grid">
      {pins.map((p) => {
        const prov = pinProvider(p);
        return (
          <a key={p.id} className="repo-pin" href={p.url} target={linkTarget()} rel="noreferrer">
            <div className="repo-pin-head">
              {providerIcon(prov, 18)}
              <span className="repo-pin-name">{p.name}</span>
            </div>
            {p.description && <div className="repo-pin-desc">{p.description}</div>}
            <div className="repo-pin-foot">
              {p.language && <span className="lang">{p.language}</span>}
              <span className="stat"><IconStar size={14} stroke={1.8} /> {p.starsCount ?? 0}</span>
              {(p.forksCount ?? 0) > 0 && (
                <span className="stat"><IconGitFork size={14} stroke={1.8} /> {p.forksCount}</span>
              )}
            </div>
          </a>
        );
      })}
    </div>
  );
};

// ── Badges tab grid ─────────────────────────────────────────────────────────
const BadgesGrid = ({ userBadges, defs }) => {
  const [selected, setSelected] = useState(null);
  const earnedMap = new Map();
  (userBadges ?? []).forEach((b) => { earnedMap.set(b.badge?.id, b.earnedAt); });

  const all = (defs ?? []).map((d) => ({
    ...d,
    earnedAt: earnedMap.get(d.id) ?? null,
  }));
  all.sort((a, b) => {
    const ai = RARITY_ORDER.indexOf(a.rarity);
    const bi = RARITY_ORDER.indexOf(b.rarity);
    return ai - bi;
  });

  if (all.length === 0) return <div className="empty">значков пока нет</div>;

  return (
    <>
      <div className="badges-grid">
        {all.map((b) => (
          <div
            key={b.id}
            className={'badge-card' + (b.earnedAt ? ` ${b.rarity}` : ' locked')}
            onClick={() => setSelected(b)}
            style={{ '--bc': b.color || RARITY_COLOR[b.rarity] || 'var(--hn-honey)', cursor: 'pointer' }}
          >
            <div className="badge-hex" style={{ '--bc': b.color || RARITY_COLOR[b.rarity] || 'var(--hn-honey)' }}>
              {b.glyph}
            </div>
            <div className="badge-name">{b.nameRu || b.name}</div>
            <span className={`badge-rarity-dot ${b.rarity}`}>{b.rarity}</span>
            <div className="badge-earned">
              {b.earnedAt ? `получен ${relativeTime(b.earnedAt)}` : 'не получен'}
            </div>
          </div>
        ))}
      </div>
      {selected && <BadgeDetailModal badge={selected} onClose={() => setSelected(null)} />}
    </>
  );
};

// ── Guest fallback ──────────────────────────────────────────────────────────
const GuestProfile = ({ onLogin }) => (
  <div className="profile-shell">
    <div style={{ textAlign: 'center', padding: '80px 24px', maxWidth: 420, margin: '0 auto' }}>
      <h2 style={{ fontFamily: 'var(--serif)', fontSize: 22, margin: '0 0 8px', fontWeight: 700 }}>
        Профиль — для участников
      </h2>
      <p style={{ fontFamily: 'var(--mono)', fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.6, margin: '0 0 20px' }}>
        войди через Zitadel, чтобы создать свой профиль, собирать значки, подписываться и публиковать посты.
      </p>
      <button className="btn primary" onClick={onLogin} style={{ width: '100%', justifyContent: 'center', padding: '11px 16px' }}>
        Войти / Зарегистрироваться
      </button>
    </div>
  </div>
);

// ── ProfileScreen ───────────────────────────────────────────────────────────
const ProfileScreen = () => {
  const { username: routeUsername } = useParams();
  const [tab, setTab] = useState('overview');
  const [banBusy, setBanBusy] = useState(false);
  const [cropState, setCropState] = useState(null);
  const { me: meProp, setMe: onMeUpdate } = useMe();
  const [meLocal, setMeLocal] = useState(null);
  const me = meLocal || meProp;
  const navigate = useNavigate();
  const { login } = useAuth();
  const avatarInput = useRef(null);
  const bannerInput = useRef(null);

  const qc = useQueryClient();
  const isOther = !!routeUsername && routeUsername !== me?.username;
  const username = isOther ? routeUsername : me?.username;

  const handlePin = async (postId, pin) => {
    try {
      if (pin) await pinPost(postId); else await unpinPost(postId);
      qc.invalidateQueries({ queryKey: qk.userPosts(username) });
    } catch {}
  };

  const otherUserQuery = useQuery({
    queryKey: qk.user(routeUsername),
    queryFn: () => getUser(routeUsername),
    enabled: isOther,
  });
  const otherUser = otherUserQuery.data;
  const otherUserLoading = otherUserQuery.isLoading;
  const otherUserError = otherUserQuery.error?.message ?? null;
  const profileUser = isOther ? otherUser : me;

  const [followingState, setFollowingState] = useState(null);
  const followMut = useFollowUser();
  const followBusy = followMut.isPending;
  const blockMut = useBlockUser();
  const { data: myBlocksList } = useQuery({
    queryKey: qk.myBlocks(),
    queryFn: getMyBlocks,
    enabled: !!me && isOther,
  });
  const isBlocked = (myBlocksList ?? []).some(u => u.username === routeUsername);

  const { data: apiFollowers } = useQuery({
    queryKey: qk.userFollowers(username),
    queryFn: () => getFollowers(username),
    enabled: !!username,
  });
  const { data: apiFollowing } = useQuery({
    queryKey: qk.userFollowing(username),
    queryFn: () => getFollowing(username),
    enabled: !!username,
  });
  const apiFriends = useMemo(() => {
    if (!apiFollowers || !apiFollowing) return null;
    const followingIds = new Set(apiFollowing.map(u => u.id));
    return apiFollowers.filter(u => followingIds.has(u.id));
  }, [apiFollowers, apiFollowing]);
  const { data: apiUserPosts } = useQuery({
    queryKey: qk.userPosts(username),
    queryFn: () => getUserPosts(username),
    enabled: !!username,
  });
  const { data: apiActivity } = useQuery({
    queryKey: qk.userActivity(username),
    queryFn: () => getUserActivity(username),
    enabled: !!username,
  });
  const { data: apiWakapi } = useQuery({
    queryKey: qk.userWakapi(username),
    queryFn: () => getUserWakapi(username).catch(() => null),
    enabled: !!username,
  });
  const { data: apiUserBadges } = useQuery({
    queryKey: qk.userBadges(username),
    queryFn: () => getUserBadges(username),
    enabled: !!username,
  });
  const { data: apiBadgeDefs } = useQuery({
    queryKey: qk.badgeDefs(),
    queryFn: getBadgeDefinitions,
  });
  const { data: myCommunity } = useQuery({
    queryKey: qk.myCommunity(),
    queryFn: () => getMyCommunity().catch(() => null),
    enabled: !isOther && !!me,
  });
  const { data: apiPinnedRepos } = useQuery({
    queryKey: qk.userPinnedRepos(username),
    queryFn: () => getUserPinnedRepos(username).catch(() => []),
    enabled: !!username,
  });

  // group pinned-repo providers for icon row — declared BEFORE early returns
  // so hook order stays stable across renders.
  const providerGroups = useMemo(() => {
    const m = new Map();
    (apiPinnedRepos ?? []).forEach((p) => {
      const prov = pinProvider(p);
      if (!prov) return;
      if (!m.has(prov)) m.set(prov, []);
      m.get(prov).push(p);
    });
    return m;
  }, [apiPinnedRepos]);

  if (!me && !isOther) return <GuestProfile onLogin={login} />;
  if (isOther && otherUserLoading) {
    return <div className="profile-shell"><div className="empty" style={{ padding: 60 }}>загрузка профиля…</div></div>;
  }
  if (isOther && (otherUserError || !profileUser)) {
    return (
      <div className="profile-shell">
        <div style={{ textAlign: 'center', padding: '80px 24px' }}>
          <h2 style={{ fontFamily: 'var(--serif)', fontSize: 22, margin: '0 0 8px' }}>Пользователь не найден</h2>
          <div style={{ fontFamily: 'var(--mono)', fontSize: 12, color: 'var(--text-dim)', marginBottom: 18 }}>
            @{routeUsername} не существует или удалён
          </div>
          <button className="btn primary" onClick={() => navigate('/')}>На ленту</button>
        </div>
      </div>
    );
  }
  if (isOther && profileUser?.privacy === 'private' && profileUser.id !== me?.id) {
    return (
      <div className="profile-shell">
        <div style={{ textAlign: 'center', padding: '80px 24px', maxWidth: 420, margin: '0 auto' }}>
          <div className="profile-avatar-big" style={{ margin: '0 auto 18px' }}>
            {profileUser.avatarUrl ? (
              <img src={profileUser.avatarUrl} alt="" style={{ width: '100%', height: '100%', borderRadius: 'inherit', objectFit: 'cover' }} />
            ) : (
              initials(profileUser.username)
            )}
          </div>
          <h2 style={{ fontFamily: 'var(--serif)', fontSize: 22, margin: '0 0 4px', fontWeight: 700 }}>
            @{profileUser.username}
          </h2>
          <div style={{ fontFamily: 'var(--mono)', fontSize: 12, color: 'var(--text-dim)', marginBottom: 18 }}>
            аккаунт скрыт
          </div>
          <p style={{ fontSize: 13.5, color: 'var(--text-mid)', lineHeight: 1.55, margin: 0 }}>
            пользователь сделал свой профиль приватным — посты, активность и подписки скрыты.
          </p>
        </div>
      </div>
    );
  }

  const isFollowing = followingState ?? (apiFollowers ?? []).some((f) => f.id === me?.id);
  const followsMe   = (apiFollowing ?? []).some((f) => f.id === me?.id);
  const isFriends   = isFollowing && followsMe;

  const toggleFollow = () => {
    if (followBusy) return;
    const next = !isFollowing;
    setFollowingState(next);
    followMut.mutate({ username: routeUsername, next }, {
      onError: (e) => {
        setFollowingState(!next);
        alert(e.message ?? 'не удалось');
      },
    });
  };

  const toggleBlock = () => {
    if (blockMut.isPending) return;
    const next = !isBlocked;
    if (next && !confirm(`Заблокировать @${routeUsername}? Его посты и комменты исчезнут из ленты.`)) return;
    blockMut.mutate({ username: routeUsername, next }, {
      onError: (e) => alert(e.message ?? 'не удалось'),
    });
  };

  const isBanned = !!profileUser?.bannedAt;

  const toggleBan = async () => {
    if (banBusy) return;
    const action = isBanned ? 'Разбанить' : 'Забанить';
    if (!confirm(`${action} @${routeUsername}? ${isBanned ? 'Пользователь снова получит доступ.' : 'Его посты и аккаунт исчезнут для всех.'}`)) return;
    setBanBusy(true);
    try {
      if (isBanned) await adminUnbanUser(routeUsername);
      else await adminBanUser(routeUsername);
      qc.invalidateQueries({ queryKey: qk.user(routeUsername) });
    } catch (e) {
      alert(e.message ?? 'не удалось');
    } finally {
      setBanBusy(false);
    }
  };

  const pickImage = (kind) => (e) => {
    if (isOther) return;
    const f = e.target.files?.[0];
    if (!f) return;
    const url = URL.createObjectURL(f);
    setCropState({ kind, src: url });
    e.target.value = '';
  };

  const handleCropDone = async (file) => {
    const kind = cropState.kind;
    setCropState(null);
    try {
      const url = await uploadFile(file);
      const updated = await updateMe(kind === 'avatar' ? { avatarUrl: url } : { bannerUrl: url });
      setMeLocal(updated);
      onMeUpdate?.(updated);
    } catch (err) {
      alert(err.message ?? 'не удалось загрузить');
    }
  };

  const displayName = profileUser.displayName || profileUser.username;
  const followers = apiFollowers?.length ?? 0;
  const following = apiFollowing?.length ?? 0;
  const postsCount = apiUserPosts?.length ?? 0;
  const reposCount = apiPinnedRepos?.length ?? 0;
  const badgesCount = apiUserBadges?.length ?? 0;
  const totalBadges = apiBadgeDefs?.length ?? 0;
  const joined = profileUser.createdAt
    ? new Date(profileUser.createdAt).toLocaleDateString('ru', { month: 'long', year: 'numeric' })
    : null;

  const pinned = (apiUserPosts ?? []).find((p) => p.isPinned) ?? null;

  const tabs = [
    { k: 'overview',  l: 'Обзор',       icon: <IconLayoutGrid size={14} stroke={1.7} /> },
    { k: 'posts',     l: 'Посты',       icon: <IconNote      size={14} stroke={1.7} />, c: postsCount },
    { k: 'repos',     l: 'Репозитории', icon: <IconCode      size={14} stroke={1.7} />, c: reposCount },
    { k: 'badges',    l: 'Значки',      icon: <IconTrophy    size={14} stroke={1.7} />, c: badgesCount },
    { k: 'activity',  l: 'Активность',  icon: <IconActivity  size={14} stroke={1.7} /> },
  ];

  return (
    <div className="profile-shell">
      {/* hero */}
      <div className="profile-hero">
        <div
          className="profile-hero-banner"
          style={profileUser.bannerUrl ? { backgroundImage: `url(${profileUser.bannerUrl})`, backgroundSize: 'cover', backgroundPosition: 'center' } : undefined}
        >
          {!isOther && (
            <>
              <button className="profile-hero-banner-edit" onClick={() => bannerInput.current?.click()} type="button">
                <IconCamera size={11} stroke={1.8} /> обновить баннер <span style={{ opacity: 0.7, fontSize: 10, fontFamily: 'var(--mono)' }}>1200×300</span>
              </button>
              <input ref={bannerInput} type="file" accept="image/*" hidden onChange={pickImage('banner')} />
            </>
          )}
        </div>
        <div className="profile-hero-body">
          <div className="profile-avatar-big">
            {profileUser.avatarUrl ? (
              <img src={profileUser.avatarUrl} alt="" style={{ width: '100%', height: '100%', borderRadius: 'inherit', objectFit: 'cover' }} />
            ) : (
              initials(profileUser.username)
            )}
            {!isOther && (
              <>
                <div className="edit-overlay" onClick={() => avatarInput.current?.click()} title="сменить аватар (200×200)">
                  <IconEdit size={13} stroke={1.8} />
                </div>
                <input ref={avatarInput} type="file" accept="image/*" hidden onChange={pickImage('avatar')} />
              </>
            )}
          </div>

          <div className="profile-id-block">
            <div className="profile-id-row">
              <h1 className="profile-name-big">{displayName}</h1>
              <span className="profile-handle-big">@{profileUser.username}</span>
            </div>
            {profileUser.bio && <p className="profile-bio-big">{profileUser.bio}</p>}
            <div className="profile-meta-row">
              <span className="item">
                <IconUsers size={13} stroke={1.7} />
                <span className="v">{followers}</span> подписчиков
              </span>
              <span className="item">
                <span className="v">{following}</span> подписок
              </span>
              {profileUser.location && (
                <span className="item">
                  <IconMapPin size={13} stroke={1.7} />
                  {profileUser.location}
                </span>
              )}
              {profileUser.website && (
                <span className="item">
                  <IconLink size={13} stroke={1.7} />
                  <a href={profileUser.website.startsWith('http') ? profileUser.website : `https://${profileUser.website}`} target={linkTarget()} rel="noreferrer">
                    {profileUser.website.replace(/^https?:\/\//, '')}
                  </a>
                </span>
              )}
              {joined && (
                <span className="item">
                  <IconCalendar size={13} stroke={1.7} />
                  с {joined}
                </span>
              )}
              {providerGroups.size > 0 && (
                <span className="profile-providers">
                  {Array.from(providerGroups.entries()).map(([prov, pins]) => (
                    <a
                      key={prov}
                      className="provider-icon-btn"
                      href={providerProfileUrl(pins[0])}
                      target={linkTarget()}
                      rel="noreferrer"
                      title={`${prov} · ${pins.length}`}
                    >
                      {providerIcon(prov, 13)}
                    </a>
                  ))}
                </span>
              )}
            </div>
          </div>

          <div className="profile-actions">
            {isOther ? (
              <button
                className={'btn' + (isFriends ? '' : ' primary')}
                style={{ justifyContent: 'center' }}
                disabled={followBusy}
                onClick={toggleFollow}
                type="button"
                title={
                  isFriends ? 'вы подписаны друг на друга — друзья'
                  : followsMe && !isFollowing ? 'этот пользователь подписан на тебя — подпишись в ответ, чтобы стать друзьями'
                  : isFollowing ? 'ты подписан, ждём ответной подписки'
                  : 'подписаться'
                }
              >
                {isFriends      ? <><IconCheck size={13} stroke={1.8} /> друзья</>
                : followsMe && !isFollowing ? <><IconUserPlus size={13} stroke={1.8} /> в друзья</>
                : isFollowing  ? <><IconCheck size={13} stroke={1.8} /> подписан</>
                : <><IconUserPlus size={13} stroke={1.8} /> подписаться</>}
              </button>
            ) : null}
            {isOther && me && (
              <button
                className="btn"
                style={{ justifyContent: 'center', color: isBlocked ? 'var(--text-dim)' : '#B23A48' }}
                disabled={blockMut.isPending}
                onClick={toggleBlock}
                type="button"
                title={isBlocked ? 'разблокировать' : 'заблокировать'}
              >
                <IconBan size={13} stroke={1.8} /> {isBlocked ? 'разблок.' : 'блок'}
              </button>
            )}
            {me?.isAdmin && isOther && (
              <button
                className="btn"
                style={{ justifyContent: 'center', color: isBanned ? 'var(--text-dim)' : '#7B2D2D', background: isBanned ? undefined : '#FFF0F0' }}
                disabled={banBusy}
                onClick={toggleBan}
                type="button"
                title={isBanned ? 'снять бан' : 'забанить (глобально)'}
              >
                <IconBan size={13} stroke={2} /> {isBanned ? 'снять бан' : 'забанить'}
              </button>
            )}
            {!isOther && (
              <>
                {myCommunity && (
                  <button className="btn primary" style={{ justifyContent: 'center' }} onClick={() => navigate(`/c/${myCommunity.slug}/submit`)} type="button">
                    <IconPencil size={13} stroke={2} /> новый пост
                  </button>
                )}
                <button className="btn" style={{ justifyContent: 'center' }} onClick={() => navigate('/settings')} type="button">
                  <IconSettings size={13} stroke={1.8} /> настройки
                </button>
                {myCommunity ? (
                  <button className="btn" style={{ justifyContent: 'center' }} onClick={() => navigate(`/c/${myCommunity.slug}`)} type="button">
                    <IconHexagon size={13} stroke={1.8} /> моё сообщество
                  </button>
                ) : (
                  <button className="btn" style={{ justifyContent: 'center' }} onClick={() => navigate('/c/new')} type="button">
                    <IconPlus size={13} stroke={1.8} /> создать сообщество
                  </button>
                )}
              </>
            )}
          </div>
        </div>
      </div>

      {/* tabs */}
      <div className="profile-tabs">
        {tabs.map((t) => (
          <div
            key={t.k}
            className={'profile-tab' + (tab === t.k ? ' active' : '')}
            onClick={() => setTab(t.k)}
          >
            {t.icon}
            {t.l}
            {t.c !== undefined && <span className="cnt">{t.c}</span>}
          </div>
        ))}
      </div>

      {/* tab-specific full-width content */}
      {tab === 'badges' && (
        <div style={{ marginTop: 20 }}>
          <BadgesGrid userBadges={apiUserBadges} defs={apiBadgeDefs} />
        </div>
      )}
      {tab === 'repos' && (
        <div style={{ marginTop: 20 }}>
          <ReposGrid pins={apiPinnedRepos} isOther={isOther} onOpenSettings={() => navigate('/settings')} />
        </div>
      )}
      {tab === 'posts' && (
        <div style={{ marginTop: 20 }}>
          <PostsList posts={apiUserPosts} isOwn={!isOther} onPin={!isOther ? handlePin : null} />
        </div>
      )}

      {/* overview / activity layout (1fr + 320px) */}
      {(tab === 'overview' || tab === 'activity') && (
        <div className="profile-content">
          <div className="profile-main">
            {tab === 'overview' && (profileUser?.showActivity !== false) && <ContribHeatmap activity={apiActivity} />}

            {pinned && tab === 'overview' && (
              <>
                <div className="pinned-row">
                  <IconPin size={13} stroke={1.8} /> закреплённый пост
                </div>
                <PostsList posts={[pinned]} onOpen={(id) => navigate(`/p/${id}`)} />
              </>
            )}

            {(profileUser?.showActivity !== false) && <ActivityCard activity={apiActivity} onOpenPost={(id) => navigate(`/p/${id}`)} />}
          </div>

          <aside className="profile-side">
            <BadgesWidget
              userBadges={apiUserBadges}
              defs={apiBadgeDefs}
              total={totalBadges}
              onAll={() => setTab('badges')}
            />
            <FriendsWidget friends={apiFriends} />
            <WakapiCard wakapi={apiWakapi} />
            <TopTagsWidget posts={apiUserPosts} />
          </aside>
        </div>
      )}

      <footer className="profile-footer">
        <div className="footer-meta">HoneyGarden · v0.10.0 · grown locally</div>
      </footer>
      {cropState && (
        <ImageCropModal
          image={cropState.src}
          aspect={cropState.kind === 'avatar' ? 1 : 7.2}
          minWidth={cropState.kind === 'avatar' ? 400 : 1200}
          hint={cropState.kind === 'avatar' ? '200×200' : '1200×160'}
          onDone={handleCropDone}
          onCancel={() => setCropState(null)}
        />
      )}
    </div>
  );
};

export default ProfileScreen;
