import { useState, useEffect, useMemo, useRef } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  IconCamera, IconEdit, IconCheck, IconUserPlus, IconStar, IconStarFilled,
  IconPencil, IconUsers, IconCalendar, IconHash,
  IconLayoutGrid, IconNote, IconHelpCircle, IconBook,
  IconX, IconSettings, IconChevronUp, IconTrash,
} from '@tabler/icons-react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { qk } from '../lib/queryKeys';
import { useJoinCommunity, useStarCommunity } from '../lib/mutations';
import { useAuth } from '../AuthContext';
import { useMe } from '../App';
import PostCard from '../components/PostCard';
import ImageCropModal from '../components/ImageCropModal';
import {
  getCommunity, getCommunityPosts, getCommunityModerators,
  getMyFollowedCommunities,
  uploadFile, updateCommunity,
  getCommunityMembers, getCommunityBans, banCommunityUser, unbanCommunityUser,
  adminDeleteCommunity,
} from '../lib/api';
import { normalizeCommunity, postToCard, initials, commColor, relativeTime } from '../lib/utils';
import { getHiddenTags } from '../lib/prefs';

const AboutCard = ({ community }) => {
  if (!community) return null;
  const created = community.createdAt
    ? new Date(community.createdAt).toLocaleDateString('ru', { day: 'numeric', month: 'long', year: 'numeric' })
    : null;
  return (
    <div className="about-card">
      <h4>о сообществе</h4>
      {community.description && <p className="about-tagline">{community.description}</p>}
      <div className="about-stats">
        <div className="about-stat">
          <div className="v">{(community.members ?? 0).toLocaleString('ru')}</div>
          <div className="l">участников</div>
        </div>
        <div className="about-stat">
          <div className="v">{(community.posts ?? 0).toLocaleString('ru')}</div>
          <div className="l">постов</div>
        </div>
      </div>
      {created && (
        <div className="about-meta">
          <div className="row">
            <span>создано</span>
            <span className="v">{created}</span>
          </div>
          {(community.stars ?? 0) > 0 && (
            <div className="row">
              <span>в избранном у</span>
              <span className="v">{community.stars}</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

// ── Rules ───────────────────────────────────────────────────────────────────
const RulesCard = ({ rules }) => {
  if (!rules || rules.length === 0) return null;
  return (
    <div className="about-card">
      <h4>правила</h4>
      <ol className="rules-list">
        {rules.map((r, i) => {
          const name = typeof r === 'string' ? r : (r.name ?? r.title ?? '');
          const desc = typeof r === 'string' ? null : (r.description ?? r.desc ?? null);
          return (
            <li key={i} className="rule-item">
              <div className="rule-name">{name}</div>
              {desc && <div className="rule-desc">{desc}</div>}
            </li>
          );
        })}
      </ol>
    </div>
  );
};

// ── Mods ────────────────────────────────────────────────────────────────────
const ModsCard = ({ mods, onOpenUser }) => {
  if (!mods || mods.length === 0) return null;
  return (
    <div className="about-card">
      <h4>модерация</h4>
      <div className="mods-list">
        {mods.map((m) => (
          <div key={m.username || m.id} className="mod-row">
            {m.avatarUrl ? (
              <img src={m.avatarUrl} alt="" className="mod-av" style={{ objectFit: 'cover' }} />
            ) : (
              <div className="mod-av">{initials(m.username || '??')}</div>
            )}
            <div>
              <div className="mod-name" onClick={() => onOpenUser?.(m.username)}>
                {m.displayName || m.username}
              </div>
              <div className="mod-handle">@{m.username}</div>
            </div>
            <span className={'mod-role' + (m.role === 'owner' ? ' owner' : '')}>
              {m.role === 'owner' ? 'owner' : 'mod'}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};

// ── Top contributors leaderboard (derived from posts) ──────────────────────
const ContribLeaderboard = ({ posts, onOpenUser }) => {
  const counts = new Map();
  (posts ?? []).forEach((p) => {
    const u = p.author?.username;
    if (!u) return;
    const cur = counts.get(u) ?? { username: u, displayName: p.author?.displayName, avatarUrl: p.author?.avatarUrl, posts: 0, votes: 0 };
    cur.posts += 1;
    cur.votes += (p.votesCount ?? 0);
    counts.set(u, cur);
  });
  const top = [...counts.values()].sort((a, b) => b.votes - a.votes).slice(0, 5);
  if (top.length === 0) return null;
  return (
    <div className="about-card">
      <h4>топ-контрибьюторы</h4>
      <div className="contrib-leaderboard">
        {top.map((u, i) => (
          <div key={u.username} className="contrib-row">
            <span className={'contrib-rank' + (i < 3 ? ' top' : '')}>#{i + 1}</span>
            {u.avatarUrl ? (
              <img className="av" src={u.avatarUrl} alt="" style={{ objectFit: 'cover' }} />
            ) : (
              <div className="av">{initials(u.username)}</div>
            )}
            <div className="info">
              <div className="h" onClick={() => onOpenUser?.(u.username)}>@{u.username}</div>
              <div className="s">{u.posts} {u.posts === 1 ? 'пост' : 'постов'}</div>
            </div>
            <div className="pts"><IconChevronUp size={13} stroke={2} /> {u.votes}</div>
          </div>
        ))}
      </div>
    </div>
  );
};

// ── Members modal (owner-only) ──────────────────────────────────────────────
const MembersModal = ({ slug, onClose, onOpenUser }) => {
  const [members, setMembers] = useState(null);
  const [bans, setBans] = useState(null);
  const [busy, setBusy] = useState(null);
  const reload = async () => {
    const [m, b] = await Promise.all([
      getCommunityMembers(slug).catch(() => []),
      getCommunityBans(slug).catch(() => []),
    ]);
    setMembers(m ?? []);
    setBans(b ?? []);
  };
  useEffect(() => { reload(); }, [slug]);
  const bannedIds = new Set((bans ?? []).map(b => b.userId));

  const ban = async (userId) => {
    setBusy(userId);
    try { await banCommunityUser(slug, userId); await reload(); } catch (e) { alert(e.message ?? 'не удалось'); }
    setBusy(null);
  };
  const unban = async (userId) => {
    setBusy(userId);
    try { await unbanCommunityUser(slug, userId); await reload(); } catch (e) { alert(e.message ?? 'не удалось'); }
    setBusy(null);
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-head">
          <h3>Управление участниками</h3>
          <button type="button" className="icon-btn" onClick={onClose}><IconX size={14} stroke={1.8} /></button>
        </div>
        {members === null ? (
          <div className="empty">грузим…</div>
        ) : (
          <>
            <div className="modal-section-title">Участники ({members.length})</div>
            <div className="member-list">
              {members.map(u => (
                <div key={u.id} className="member-row">
                  {u.avatarUrl ? (
                    <img src={u.avatarUrl} alt="" className="member-av" />
                  ) : (
                    <div className="member-av placeholder">{initials(u.username)}</div>
                  )}
                  <div className="member-id">
                    <button type="button" className="member-name" onClick={() => onOpenUser(u.username)}>
                      {u.displayName || u.username}
                    </button>
                    <div className="member-handle">@{u.username}</div>
                  </div>
                  {bannedIds.has(u.id) ? (
                    <button type="button" className="btn" disabled={busy === u.id} onClick={() => unban(u.id)}>
                      разбанить
                    </button>
                  ) : (
                    <button type="button" className="btn-danger" disabled={busy === u.id} onClick={() => ban(u.id)}>
                      забанить
                    </button>
                  )}
                </div>
              ))}
              {members.length === 0 && <div className="empty">пока нет подписчиков</div>}
            </div>
          </>
        )}
      </div>
    </div>
  );
};

// ── CommunityScreen ─────────────────────────────────────────────────────────
const CommunityScreen = () => {
  const { slug } = useParams();
  const navigate = useNavigate();
  const { me } = useMe();
  const { isAuthenticated, login } = useAuth();

  const [tab, setTab] = useState('all');
  const [tagFilter, setTagFilter] = useState(null);
  const [starredLocal, setStarredLocal] = useState(null);
  const [showMembers, setShowMembers] = useState(false);
  const avatarInput = useRef(null);
  const bannerInput = useRef(null);
  const [commOverride, setCommOverride] = useState(null);
  const [cropState, setCropState] = useState(null);
  const qc = useQueryClient();
  const joinMut = useJoinCommunity();
  const starMut = useStarCommunity();

  const handleAdminDelete = async () => {
    if (!confirm(`Удалить сообщество /${slug}? Все посты будут удалены. Необратимо.`)) return;
    try {
      await adminDeleteCommunity(slug);
      qc.invalidateQueries({ queryKey: ['communities'] });
      navigate('/communities');
    } catch (e) {
      alert(e.message ?? 'не удалось');
    }
  };

  const kindForTab = tab === 'discussions' ? 'discussion'
                   : tab === 'articles'    ? 'article'
                   : tab === 'questions'   ? 'question'
                   : null;

  const { data: commRaw } = useQuery({
    queryKey: qk.community(slug),
    queryFn: () => getCommunity(slug),
    enabled: !!slug,
  });
  const { data: postsRaw, isLoading: postsLoading } = useQuery({
    queryKey: qk.communityPosts(slug, kindForTab),
    queryFn: () => getCommunityPosts(slug, 0, kindForTab),
    enabled: !!slug,
  });
  const { data: modsRaw } = useQuery({
    queryKey: qk.communityMods(slug),
    queryFn: () => getCommunityModerators(slug).catch(() => []),
    enabled: !!slug,
  });
  const { data: followedList } = useQuery({
    queryKey: qk.myCommunities(),
    queryFn: getMyFollowedCommunities,
    enabled: isAuthenticated,
  });

  const commSource = commOverride ?? commRaw;
  const community = commSource ? normalizeCommunity(commSource) : null;
  const isOwner = !!(me && commSource && commSource.ownerId === me.id);
  const mods = modsRaw ?? [];

  const followed = (followedList ?? []).some((c) => c.slug === slug);
  const busy = joinMut.isPending;

  const toggleFollow = () => {
    if (!isAuthenticated) { login(); return; }
    joinMut.mutate({ slug, next: !followed });
  };

  const starred = starredLocal ?? !!commSource?.starred;
  const toggleStar = () => {
    if (!isAuthenticated) { login(); return; }
    const next = !starred;
    setStarredLocal(next);
    starMut.mutate({ slug, next }, { onError: () => setStarredLocal(!next) });
  };

  const pickImage = (kind) => (e) => {
    if (!isOwner) return;
    const f = e.target.files?.[0];
    if (!f) return;
    const src = URL.createObjectURL(f);
    setCropState({ kind, src });
    e.target.value = '';
  };

  const handleCropDone = async (file) => {
    const kind = cropState.kind;
    setCropState(null);
    try {
      const url = await uploadFile(file);
      const updated = await updateCommunity(slug, kind === 'avatar' ? { avatarUrl: url } : { bannerUrl: url });
      setCommOverride(updated);
    } catch (err) {
      alert(err.message ?? 'не удалось загрузить');
    }
  };

  // Posts → cards, tab + tag filters
  const allCards = useMemo(
    () => (postsRaw ?? []).map((p) => ({ raw: p, card: postToCard(p) })),
    [postsRaw],
  );

  const filtered = useMemo(() => {
    let xs = allCards;
    if (tagFilter) xs = xs.filter((x) => (x.card.tags ?? []).includes(tagFilter));
    const hidden = new Set(getHiddenTags());
    if (hidden.size) {
      xs = xs.filter((x) => !(x.card.tags ?? []).some(t => hidden.has(String(t).toLowerCase())));
    }
    return xs;
  }, [allCards, tagFilter]);

  const tabs = [
    { k: 'all',          l: 'Все',         icon: <IconLayoutGrid size={14} stroke={1.7} />, c: allCards.length },
    { k: 'discussions',  l: 'Обсуждения',  icon: <IconNote size={14} stroke={1.7} /> },
    { k: 'articles',     l: 'Статьи',      icon: <IconBook size={14} stroke={1.7} /> },
    { k: 'questions',    l: 'Вопросы',     icon: <IconHelpCircle size={14} stroke={1.7} /> },
  ];

  if (!commRaw && !commOverride) {
    return <div className="comm-shell"><div className="empty" style={{ padding: 60 }}>загрузка сообщества…</div></div>;
  }

  const commColorVal = community?.color ?? commColor(slug);
  const slugLetter = (community?.slug ?? slug)[0]?.toUpperCase() ?? '?';

  const tagsAvailable = community?.tags ?? [];

  return (
    <>
      <div className="comm-hero">
        <div
          className={'comm-banner' + (community?.bannerUrl ? ' has-image' : '')}
          style={community?.bannerUrl
            ? { backgroundImage: `url(${community.bannerUrl})`, backgroundSize: 'cover', backgroundPosition: 'center' }
            : undefined}
        >
          {isOwner && (
            <>
              <button className="comm-banner-edit" type="button" onClick={() => bannerInput.current?.click()}>
                <IconCamera size={11} stroke={1.8} /> обновить баннер
              </button>
              <input ref={bannerInput} type="file" accept="image/*" hidden onChange={pickImage('banner')} />
            </>
          )}
        </div>

        <div className="comm-head">
          <div
            className="comm-icon-big"
            style={{ background: commColorVal, cursor: isOwner ? 'pointer' : 'default' }}
            onClick={() => isOwner && avatarInput.current?.click()}
          >
            {community?.avatarUrl ? (
              <img src={community.avatarUrl} alt="" />
            ) : slugLetter}
            {isOwner && <input ref={avatarInput} type="file" accept="image/*" hidden onChange={pickImage('avatar')} />}
          </div>

          <div className="comm-id">
            <div className="comm-name-row">
              <span className="comm-slug">
                <span className="prefix">hg/</span>{community?.slug ?? slug}
              </span>
              {community?.name && community.name !== community.slug && (
                <span className="comm-fullname">{community.name}</span>
              )}
            </div>
            {community?.description && <p className="comm-tagline">{community.description}</p>}
            <div className="comm-meta">
              <span className="item">
                <IconUsers size={13} stroke={1.7} />
                <span className="v">{(community?.members ?? 0).toLocaleString('ru')}</span> участников
              </span>
              <span className="item">
                <span className="v">{community?.posts ?? 0}</span> постов
              </span>
              {community?.stars > 0 && (
                <span className="item">
                  <IconStar size={13} stroke={1.7} />
                  <span className="v">{community.stars}</span>
                </span>
              )}
              {community?.createdAt && (
                <span className="item">
                  <IconCalendar size={13} stroke={1.7} />
                  с {new Date(community.createdAt).toLocaleDateString('ru', { month: 'long', year: 'numeric' })}
                </span>
              )}
            </div>
          </div>

          <div className="comm-actions">
            {!isOwner && (
              <button
                className={'btn' + (followed ? '' : ' primary')}
                onClick={toggleFollow}
                disabled={busy}
                type="button"
              >
                {followed ? <IconCheck size={13} stroke={1.8} /> : <IconUserPlus size={13} stroke={1.8} />}
                {followed ? ' подписан' : ' подписаться'}
              </button>
            )}
            <button
              className="icon-btn"
              onClick={toggleStar}
              title={starred ? 'убрать из избранного' : 'в избранное'}
              type="button"
            >
              {starred ? <IconStarFilled size={15} stroke={1.7} /> : <IconStar size={15} stroke={1.7} />}
            </button>
            {isOwner && (
              <button
                className="btn"
                onClick={() => setShowMembers(true)}
                title="управление участниками"
                type="button"
              >
                <IconSettings size={13} stroke={1.8} /> участники
              </button>
            )}
            {isOwner && (
              <button
                className="btn primary"
                onClick={() => navigate(`/c/${slug}/submit`)}
                type="button"
              >
                <IconPencil size={13} stroke={2} /> новый пост
              </button>
            )}
            {me?.isAdmin && !isOwner && (
              <button
                className="btn"
                style={{ color: '#7B2D2D', background: '#FFF0F0' }}
                onClick={handleAdminDelete}
                type="button"
                title="удалить сообщество (admin)"
              >
                <IconTrash size={13} stroke={1.8} /> удалить (admin)
              </button>
            )}
          </div>
        </div>
      </div>

      <div className="comm-tabs">
        {tabs.map((t) => (
          <div
            key={t.k}
            className={'comm-tab' + (tab === t.k ? ' active' : '')}
            onClick={() => setTab(t.k)}
          >
            {t.icon}
            {t.l}
            {t.c !== undefined && <span className="cnt">{t.c}</span>}
          </div>
        ))}
      </div>

      {/* main + side */}
      <div className="comm-shell">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16, minWidth: 0 }}>
          {tagsAvailable.length > 0 && (
            <div className="tag-filter-row">
              <span className="label">фильтр:</span>
              {tagsAvailable.map((t) => (
                <span
                  key={t}
                  className={'tag-chip' + (tagFilter === t ? ' active' : '')}
                  onClick={() => setTagFilter(tagFilter === t ? null : t)}
                >
                  <IconHash size={11} stroke={1.8} />
                  {t}
                  {tagFilter === t && <span className="x"><IconX size={10} stroke={2} /></span>}
                </span>
              ))}
            </div>
          )}

          {postsLoading && filtered.length === 0 ? (
            <div className="empty">грузим посты…</div>
          ) : filtered.length === 0 ? (
            <div className="empty">
              {tagFilter ? `нет постов с тегом #${tagFilter}` : 'пока нет постов'}
            </div>
          ) : (
            filtered.map((x) => (
              <PostCard key={x.card.id} post={x.card} onOpen={(id) => navigate(`/p/${id}`)} />
            ))
          )}
        </div>

        <aside style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <AboutCard community={community} />
          <RulesCard rules={community?.rules} />
          <ModsCard mods={mods} onOpenUser={(u) => navigate(`/u/${u}`)} />
          <ContribLeaderboard posts={postsRaw} onOpenUser={(u) => navigate(`/u/${u}`)} />
        </aside>
      </div>

      {showMembers && (
        <MembersModal
          slug={slug}
          onClose={() => setShowMembers(false)}
          onOpenUser={(u) => { setShowMembers(false); navigate(`/u/${u}`); }}
        />
      )}

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
    </>
  );
};

export default CommunityScreen;
