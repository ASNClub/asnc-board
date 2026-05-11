import { useState, useEffect, useRef } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { IconSearch, IconBell, IconPencil, IconShieldLock } from '@tabler/icons-react';
import { getNotificationUnreadCount, searchPosts, searchCommunities, searchUsers } from './lib/api';
import { relativeTime, commColor, initials } from './lib/utils';
import { useAuth } from './AuthContext';

const SearchBar = () => {
  const [q, setQ] = useState('');
  const [open, setOpen] = useState(false);
  const [posts, setPosts] = useState([]);
  const [communities, setCommunities] = useState([]);
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(false);
  const wrap = useRef(null);
  const inputRef = useRef(null);
  const navigate = useNavigate();

  useEffect(() => {
    const onClick = (e) => { if (wrap.current && !wrap.current.contains(e.target)) setOpen(false); };
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, []);

  useEffect(() => {
    const onKey = (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        inputRef.current?.focus();
      }
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, []);

  useEffect(() => {
    if (q.trim().length < 2) { setPosts([]); setCommunities([]); setUsers([]); return; }
    let cancelled = false;
    setLoading(true);
    const t = setTimeout(async () => {
      try {
        const [p, c, u] = await Promise.all([
          searchPosts(q.trim()).catch(() => []),
          searchCommunities(q.trim()).catch(() => []),
          searchUsers(q.trim()).catch(() => []),
        ]);
        if (cancelled) return;
        setPosts(p ?? []);
        setCommunities(c ?? []);
        setUsers((u ?? []).filter(x => x && x.username));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }, 250);
    return () => { cancelled = true; clearTimeout(t); };
  }, [q]);

  const go = (path) => { setOpen(false); setQ(''); navigate(path); };
  const hasResults = posts.length > 0 || communities.length > 0 || users.length > 0;

  return (
    <div className="search" ref={wrap} style={{ position: 'relative' }}>
      <IconSearch size={15} stroke={1.8} />
      <input
        ref={inputRef}
        placeholder="найти посты, сообщества, людей…"
        value={q}
        onFocus={() => setOpen(true)}
        onChange={(e) => { setQ(e.target.value); setOpen(true); }}
      />
      {open && q.trim().length >= 2 && (
        <div className="search-dropdown">
          {loading && <div className="search-empty">ищем…</div>}
          {!loading && !hasResults && <div className="search-empty">ничего не нашли</div>}
          {communities.length > 0 && (
            <div className="search-group">
              <div className="search-group-title">сообщества</div>
              {communities.slice(0, 5).map(c => (
                <div key={c.id ?? c.slug} className="search-item" onMouseDown={() => go(`/c/${c.slug}`)}>
                  <div className="search-item-icon" style={{ background: commColor(c.slug), color: 'white' }}>
                    {c.avatarUrl
                      ? <img src={c.avatarUrl} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', borderRadius: '50%' }} />
                      : c.slug?.[0]?.toUpperCase() ?? '?'}
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div className="search-item-title">hg/{c.slug}</div>
                    <div className="search-item-sub">{c.name ?? ''}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
          {users.length > 0 && (
            <div className="search-group">
              <div className="search-group-title">люди</div>
              {users.slice(0, 5).map(u => (
                <div key={u.id ?? u.username} className="search-item" onMouseDown={() => go(`/u/${u.username}`)}>
                  {u.avatarUrl ? (
                    <img src={u.avatarUrl} alt="" className="search-item-icon" style={{ objectFit: 'cover' }} />
                  ) : (
                    <div className="search-item-icon" style={{ background: 'var(--hn-honey-pale)', color: 'var(--hn-honey-dark)' }}>
                      {initials(u.username || '??')}
                    </div>
                  )}
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div className="search-item-title">@{u.username}</div>
                    <div className="search-item-sub">{u.displayName || ''}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
          {posts.length > 0 && (
            <div className="search-group">
              <div className="search-group-title">посты</div>
              {posts.slice(0, 6).map(p => (
                <div key={p.id} className="search-item" onMouseDown={() => go(`/p/${p.id}`)}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div className="search-item-title">{p.title ?? p.content?.slice(0, 60)}</div>
                    <div className="search-item-sub">
                      hg/{p.communitySlug ?? '?'} · {relativeTime(p.createdAt)}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

const TopNav = ({ me }) => {
  const { isAuthenticated, login } = useAuth();
  const isAdmin = me?.isAdmin ?? false;
  const navigate = useNavigate();
  const location = useLocation();
  const [unread, setUnread] = useState(0);
  useEffect(() => {
    if (!isAuthenticated) { setUnread(0); return; }
    let cancelled = false;
    const tick = () => {
      getNotificationUnreadCount()
        .then((d) => { if (!cancelled) setUnread(d?.count ?? 0); })
        .catch(() => {});
    };
    tick();
    const id = setInterval(tick, 30000);
    return () => { cancelled = true; clearInterval(id); };
  }, [isAuthenticated]);

  const isFeedActive = location.pathname === '/';
  const isCommunitiesActive = location.pathname === '/c' || location.pathname.startsWith('/c/');
  const isProfileActive = location.pathname === '/me' || location.pathname.startsWith('/u/') || location.pathname === '/settings';
  const isNotifActive = location.pathname === '/notifications';
  const isRoadmapActive = location.pathname === '/roadmap';

  return (
    <header className="topnav">
      <div className="topnav-inner">
        <Link className="brand" to="/">
          <div className="hex">H</div>
          HoneyGarden
        </Link>
        <SearchBar />
        <div className="nav-spacer" />
        <nav className="nav-links">
          <Link to="/" className={'nav-link' + (isFeedActive ? ' active' : '')}>Лента</Link>
          <Link to="/c" className={'nav-link' + (isCommunitiesActive ? ' active' : '')}>Сообщества</Link>
          {isAuthenticated && (
            <Link to="/submit" className={'nav-link mobile-only' + (location.pathname === '/submit' ? ' active' : '')}>
              <IconPencil size={16} stroke={1.8} />
            </Link>
          )}
          {isAuthenticated && (
            <Link to="/notifications" className={'nav-link mobile-only' + (isNotifActive ? ' active' : '')} style={{ position: 'relative' }}>
              <IconBell size={16} stroke={1.8} />
              {unread > 0 && <span className="dot" />}
            </Link>
          )}
          <Link to="/roadmap" className={'nav-link' + (isRoadmapActive ? ' active' : '')}>Роадмап</Link>
          <Link to="/me" className={'nav-link' + (isProfileActive ? ' active' : '')}>Профиль</Link>
        </nav>
        <div className="topnav-actions">
          {isAuthenticated ? (
            <>
              {isAdmin && (
                <Link className={'icon-btn' + (location.pathname === '/admin' ? ' active' : '')} title="Админка" to="/admin">
                  <IconShieldLock size={17} stroke={1.7} />
                </Link>
              )}
              <Link
                className={'icon-btn' + (isNotifActive ? ' active' : '')}
                title="Уведомления"
                to="/notifications"
              >
                <IconBell size={17} stroke={1.7} />
                {unread > 0 && <span className="dot" />}
              </Link>
              <Link className="btn primary" to="/submit">
                <IconPencil size={13} stroke={2} />
                Новый пост
              </Link>
            </>
          ) : (
            <button className="btn primary" onClick={login} type="button">
              Войти
            </button>
          )}
        </div>
      </div>
    </header>
  );
};

export default TopNav;
