import { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  IconCheck, IconArrowUp, IconMessage, IconUserPlus,
  IconAt, IconAward, IconBell, IconSettings,
} from '@tabler/icons-react';
import { useQuery } from '@tanstack/react-query';
import { qk } from '../lib/queryKeys';
import { useMarkNotificationRead, useMarkAllNotificationsRead } from '../lib/mutations';
import { getNotifications } from '../lib/api';

const TABS = [
  { k: 'all',     l: 'Все' },
  { k: 'mention', l: 'Упоминания' },
  { k: 'vote',    l: 'Голоса' },
  { k: 'comment', l: 'Комментарии' },
  { k: 'follow',  l: 'Подписки' },
  { k: 'badge',   l: 'Бейджи' },
];

const TYPE_MAP = {
  'user.followed':        { kind: 'follow' },
  'friendship.requested': { kind: 'follow' },
  'friendship.accepted':  { kind: 'follow' },
  'community.followed':   { kind: 'follow' },
  'community.starred':    { kind: 'vote'   },
  'comment.created':      { kind: 'comment' },
  'post.voted':           { kind: 'vote'   },
  'post.created':         { kind: 'comment' },
  'mention':              { kind: 'mention' },
  'mention.created':      { kind: 'mention' },
  'badge.awarded':        { kind: 'badge'  },
};

const ICON_BY_KIND = {
  vote:    IconArrowUp,
  comment: IconMessage,
  follow:  IconUserPlus,
  mention: IconAt,
  badge:   IconAward,
};

const parsePayload = (p) => {
  if (!p) return {};
  try { return typeof p === 'string' ? JSON.parse(p) : p; } catch { return {}; }
};

const relativeTime = (iso) => {
  if (!iso) return '';
  const d = new Date(iso);
  const sec = Math.floor((Date.now() - d.getTime()) / 1000);
  if (sec < 60) return `${sec} с`;
  if (sec < 3600) return `${Math.floor(sec / 60)} мин`;
  if (sec < 86400) return `${Math.floor(sec / 3600)} ч`;
  if (sec < 7 * 86400) return d.toLocaleDateString('ru', { weekday: 'short' });
  if (sec < 30 * 86400) return `${Math.floor(sec / 86400)} д`;
  return `${Math.floor(sec / (7 * 86400))} нед`;
};

const groupByDate = (items) => {
  const today = []; const week = []; const older = [];
  const now = Date.now();
  for (const n of items) {
    const d = new Date(n.createdAt).getTime();
    const dayDiff = Math.floor((now - d) / 86400000);
    if (dayDiff < 1) today.push(n);
    else if (dayDiff < 7) week.push(n);
    else older.push(n);
  }
  return [
    { label: 'сегодня', items: today },
    { label: 'на этой неделе', items: week },
    { label: 'раньше', items: older },
  ].filter(g => g.items.length > 0);
};

const renderTitle = (n) => {
  const actor = n.actor;
  const handle = actor?.username ?? '?';
  const aBtn = (
    <span className="author">@{handle}</span>
  );
  const post = n.post;
  const comm = n.community;
  const postTitle = post?.title ? <span className="target">«{post.title}»</span> : null;
  const commTitle = comm?.slug ? <span className="target">hg/{comm.slug}</span> : null;

  switch (n.type) {
    case 'user.followed':
      return <>{aBtn} подписался на тебя</>;
    case 'friendship.requested':
      return <>{aBtn} отправил запрос на дружбу</>;
    case 'friendship.accepted':
      return <>{aBtn} принял запрос на дружбу</>;
    case 'community.followed':
      return <>{aBtn} подписался на {commTitle ?? 'твоё сообщество'}</>;
    case 'community.starred':
      return <>{aBtn} добавил {commTitle ?? 'твоё сообщество'} в избранное</>;
    case 'comment.created':
      return <>{aBtn} прокомментировал {postTitle ?? 'твой пост'}{commTitle ? <> в {commTitle}</> : null}</>;
    case 'post.voted':
      return <>{aBtn} поддержал {postTitle ?? 'твой пост'}</>;
    case 'post.created':
      return <>{aBtn} опубликовал {postTitle ?? 'пост'}{commTitle ? <> в {commTitle}</> : null}</>;
    case 'mention':
    case 'mention.created':
      return <>{aBtn} упомянул тебя{postTitle ? <> в {postTitle}</> : null}{commTitle ? <> ({commTitle})</> : null}</>;
    case 'badge.awarded':
      return <>Получен новый значок</>;
    default:
      return <>{n.type}</>;
  }
};

const NotifRow = ({ n, onClick }) => {
  const map = TYPE_MAP[n.type] ?? { kind: 'comment' };
  const Icon = ICON_BY_KIND[map.kind] ?? IconBell;
  const payload = parsePayload(n.payload);
  const actor = n.actor;

  return (
    <div className={'notif-row' + (n.isRead ? '' : ' unread')} onClick={() => onClick(n, payload)}>
      <div className="notif-avatar-wrap">
        {actor?.avatarUrl
          ? <img className="notif-avatar" src={actor.avatarUrl} alt="" />
          : <div className={'notif-icon ' + map.kind}><Icon size={18} stroke={2} /></div>}
        {actor?.avatarUrl && (
          <span className={'notif-icon-badge ' + map.kind}>
            <Icon size={11} stroke={2.2} />
          </span>
        )}
      </div>
      <div className="notif-body">
        <div className="notif-title">{renderTitle(n)}</div>
        {n.snippet && <div className="notif-snippet">{n.snippet}</div>}
      </div>
      <span className="notif-time">{relativeTime(n.createdAt)}</span>
    </div>
  );
};

const NotificationsScreen = () => {
  const navigate = useNavigate();
  const [tab, setTab] = useState('all');
  const [items, setItems] = useState([]);
  const notifQuery = useQuery({
    queryKey: qk.notifications(),
    queryFn: getNotifications,
  });
  const raw = notifQuery.data;
  const loading = notifQuery.isLoading;
  const error = notifQuery.error?.message ?? null;
  const markOneMut = useMarkNotificationRead();
  const markAllMut = useMarkAllNotificationsRead();

  useEffect(() => {
    if (raw) setItems(Array.isArray(raw) ? raw : []);
  }, [raw]);

  const handleMarkAll = () => {
    setItems(prev => prev.map(n => ({ ...n, isRead: true })));
    markAllMut.mutate();
  };

  const handleClick = (n, payload) => {
    if (!n.isRead) {
      setItems(prev => prev.map(x => x.id === n.id ? { ...x, isRead: true } : x));
      markOneMut.mutate({ id: n.id });
    }
    if (n.post?.id) navigate(`/p/${n.post.id}`);
    else if (n.community?.slug) navigate(`/c/${n.community.slug}`);
    else if (n.actor?.username) navigate(`/u/${n.actor.username}`);
    else if (payload.post_id) navigate(`/p/${payload.post_id}`);
    else if (payload.community_slug) navigate(`/c/${payload.community_slug}`);
    else if (payload.username) navigate(`/u/${payload.username}`);
  };

  const counts = useMemo(() => {
    const c = { all: 0 };
    for (const n of items) {
      if (n.isRead) continue;
      c.all++;
      const kind = TYPE_MAP[n.type]?.kind ?? 'comment';
      c[kind] = (c[kind] ?? 0) + 1;
    }
    return c;
  }, [items]);

  const filtered = tab === 'all'
    ? items
    : items.filter(n => (TYPE_MAP[n.type]?.kind ?? 'comment') === tab);

  const groups = groupByDate(filtered);

  return (
    <div className="notif-shell">
      <div className="notif-head">
        <h1>Уведомления{counts.all > 0 && <span className="cnt">{counts.all}</span>}</h1>
        <div className="actions">
          {counts.all > 0 && (
            <button className="btn ghost" onClick={handleMarkAll}>
              <IconCheck size={13} stroke={2} /> Прочитать все
            </button>
          )}
          <button className="icon-btn" title="Настройки" onClick={() => navigate('/settings#notifications')}>
            <IconSettings size={16} stroke={2} />
          </button>
        </div>
      </div>

      <div className="notif-tabs">
        {TABS.map(t => (
          <button key={t.k} className={'notif-tab' + (tab === t.k ? ' active' : '')} onClick={() => setTab(t.k)}>
            {t.l}
            {counts[t.k] > 0 && <span className="cnt">{counts[t.k]}</span>}
          </button>
        ))}
      </div>

      {loading && <div style={{ padding: 30, textAlign: 'center', color: 'var(--text-dim)', fontFamily: 'var(--mono)', fontSize: 13 }}>загрузка…</div>}
      {error && <div className="error-banner">{error.message ?? 'ошибка'}</div>}

      {!loading && filtered.length === 0 && (
        <div className="notif-empty">
          <div className="ico"><IconBell size={24} stroke={2} /></div>
          <div className="title">Тут тихо</div>
          <div className="sub">когда начнут отвечать или подписываться — увидишь здесь</div>
        </div>
      )}

      {groups.map(g => (
        <div className="notif-group" key={g.label}>
          <div className="notif-group-head">{g.label}</div>
          <div className="notif-card">
            {g.items.map(n => <NotifRow key={n.id} n={n} onClick={handleClick} />)}
          </div>
        </div>
      ))}
    </div>
  );
};

export default NotificationsScreen;
