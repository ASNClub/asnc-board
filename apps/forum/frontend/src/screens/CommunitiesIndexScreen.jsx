import { useState, useEffect, useMemo } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  IconSearch, IconUsers, IconMessage, IconClock, IconPlus,
} from '@tabler/icons-react';
import { useQuery } from '@tanstack/react-query';
import { qk } from '../lib/queryKeys';
import { useJoinCommunity } from '../lib/mutations';
import { useAuth } from '../AuthContext';
import {
  getMyFollowedCommunities, searchCommunities, getMyCommunity, listCommunities,
} from '../lib/api';
import { commColor } from '../lib/utils';

const SORTS = [
  { k: 'popular', l: 'Популярные' },
  { k: 'active',  l: 'Активные' },
  { k: 'new',     l: 'Новые' },
];

const formatNum = (n) => (n ?? 0).toLocaleString('ru');

const CommCard = ({ c, followed, busy, onToggle, isOwn }) => (
  <Link className="cidx-card" to={`/c/${c.slug}`} style={{ textDecoration: 'none', color: 'inherit' }}>
    <div className="cidx-icon" style={{ background: commColor(c.slug) }}>
      {c.avatarUrl
        ? <img src={c.avatarUrl} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
        : (c.slug || '?').slice(0, 2).toLowerCase()}
    </div>
    <div className="body">
      <div className="name-row">
        <span className="name">{c.name || c.slug}</span>
        <span className="slug"><span className="prefix">hg/</span>{c.slug}</span>
      </div>
      {c.description && <div className="tagline">{c.description}</div>}
      <div className="stats">
        <span className="it">
          <IconUsers size={11} stroke={2} />
          <span className="v">{formatNum(c.followersCount)}</span> участников
        </span>
        <span className="it">
          <IconMessage size={11} stroke={2} />
          <span className="v">{formatNum(c.postsCount)}</span> постов
        </span>
        {c.createdAt && (
          <span className="it">
            <IconClock size={11} stroke={2} />
            с {new Date(c.createdAt).getFullYear()}
          </span>
        )}
      </div>
    </div>
    <div className="actions" onClick={(e) => e.stopPropagation()}>
      {isOwn ? (
        <span className="btn" style={{ opacity: .7, cursor: 'default' }}>Моё</span>
      ) : (
        <button
          className={'btn' + (followed ? ' primary' : '')}
          onClick={(e) => { e.preventDefault(); onToggle(c.slug, !followed); }}
          disabled={busy}
        >
          {followed ? 'Подписан' : '+ Подписаться'}
        </button>
      )}
    </div>
  </Link>
);

const CommunitiesIndexScreen = () => {
  const navigate = useNavigate();
  const { isAuthenticated, login } = useAuth();
  const [q, setQ]               = useState('');
  const [searchRes, setSearchRes] = useState([]);
  const [searching, setSearching] = useState(false);
  const [sort, setSort]         = useState('popular');
  const joinMut = useJoinCommunity();

  const { data: followedRaw } = useQuery({
    queryKey: qk.myCommunities(),
    queryFn: getMyFollowedCommunities,
    enabled: isAuthenticated,
  });
  const followed = followedRaw ?? [];
  const followedSet = useMemo(() => new Set(followed.map(c => c.slug)), [followed]);

  const { data: ownCommunity } = useQuery({
    queryKey: qk.myCommunity(),
    queryFn: () => getMyCommunity().catch(() => null),
    enabled: isAuthenticated,
  });

  const { data: discoverList } = useQuery({
    queryKey: qk.communitiesList(sort),
    queryFn: () => listCommunities(sort, 50, 0),
  });

  // Debounced search
  useEffect(() => {
    if (q.trim().length < 2) { setSearchRes([]); setSearching(false); return; }
    let cancelled = false;
    setSearching(true);
    const t = setTimeout(async () => {
      try {
        const res = await searchCommunities(q.trim());
        const list = Array.isArray(res?.hits) ? res.hits : (Array.isArray(res) ? res : []);
        if (!cancelled) setSearchRes(list);
      } finally {
        if (!cancelled) setSearching(false);
      }
    }, 250);
    return () => { cancelled = true; clearTimeout(t); };
  }, [q]);

  const list = q.trim().length >= 2
    ? searchRes
    : (discoverList ?? []);
  const sorted = list;

  const toggle = (slug, wantFollow) => {
    if (!isAuthenticated) { login(); return; }
    joinMut.mutate({ slug, next: wantFollow });
  };
  const busyFor = (slug) => joinMut.isPending && joinMut.variables?.slug === slug;

  return (
    <div className="cidx-shell">
      <div className="cidx-head">
        <div>
          <h1>Сад <span className="amp">сообществ</span></h1>
          <div className="sub">Подпишись на интересное или создай своё</div>
        </div>
        <div className="actions">
          {ownCommunity ? (
            <button className="btn primary" onClick={() => navigate(`/c/${ownCommunity.slug}`)}>
              Открыть моё сообщество
            </button>
          ) : (
            <button className="btn primary" onClick={() => navigate('/c/new')}>
              <IconPlus size={13} stroke={2} />
              Создать сообщество
            </button>
          )}
        </div>
      </div>

      <div className="cidx-toolbar">
        <div className="cidx-search">
          <IconSearch size={14} stroke={2} />
          <input
            placeholder="искать по названию или тегу…"
            value={q}
            onChange={e => setQ(e.target.value)}
          />
        </div>
        <div className="cidx-sort-tabs">
          {SORTS.map(s => (
            <button
              key={s.k}
              className={'cidx-sort-tab' + (sort === s.k ? ' active' : '')}
              onClick={() => setSort(s.k)}
            >
              {s.l}
            </button>
          ))}
        </div>
      </div>

      {searching && <div className="cidx-empty">ищем…</div>}

      {!searching && sorted.length === 0 && (
        <div className="cidx-empty">
          {q.trim().length >= 2 ? 'ничего не нашли' : 'пока сообществ нет'}
        </div>
      )}

      {sorted.length > 0 && (
        <div className="cidx-grid">
          {sorted.map(c => (
            <CommCard
              key={c.slug}
              c={c}
              followed={followedSet.has(c.slug)}
              busy={busyFor(c.slug)}
              onToggle={toggle}
              isOwn={ownCommunity?.slug === c.slug}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default CommunitiesIndexScreen;
