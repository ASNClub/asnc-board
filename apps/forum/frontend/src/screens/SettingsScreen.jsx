import { useState, useEffect, useRef } from 'react';
import {
  IconUser, IconSettings, IconPalette, IconBell, IconSearch,
  IconLock, IconWorld, IconTrash, IconCheck, IconX, IconLoader2,
  IconBrandGithub, IconStar,
} from '@tabler/icons-react';
import { useMe } from '../App';

const PROVIDER_LABELS = { github: 'GitHub' };
import {
  getMe, updateMe, uploadFile, setUserTags,
  getNotificationPreferences, setNotificationPreference,
  listIntegrations, beginIntegrationConnect, disconnectIntegration,
  listIntegrationRepos, setPinnedRepos, getUserPinnedRepos,
  connectWakapi, disconnectWakapi, getUserWakapi,
} from '../lib/api';
import { applyAppearance } from '../lib/appearance';
import { INTEREST_TAGS } from '../data/tags';

const SECTIONS = [
  { id: 'profile',       label: 'Профиль',      Icon: IconUser },
  { id: 'account',       label: 'Аккаунт',      Icon: IconSettings },
  { id: 'appearance',    label: 'Внешний вид',  Icon: IconPalette },
  { id: 'notifications', label: 'Уведомления',  Icon: IconBell },
  { id: 'feed',          label: 'Лента',        Icon: IconSearch },
  { id: 'privacy',       label: 'Приватность',  Icon: IconLock },
  { id: 'connections',   label: 'Подключения',  Icon: IconWorld },
];

const Toggle = ({ checked, onChange }) => (
  <label className="toggle">
    <input type="checkbox" checked={!!checked} onChange={e => onChange?.(e.target.checked)} />
    <span className="slider" />
  </label>
);

const Segmented = ({ options, value, onChange }) => (
  <div className="segmented">
    {options.map(o => (
      <button
        key={o.value}
        type="button"
        className={'seg-btn' + (value === o.value ? ' active' : '')}
        onClick={() => onChange(o.value)}
      >
        {o.label}
      </button>
    ))}
  </div>
);

const NavLink = ({ id, label, Icon, active, onClick, danger }) => (
  <a
    className={'settings-nav-link' + (active ? ' active' : '')}
    onClick={() => onClick(id)}
    style={danger ? { color: 'var(--hn-red)' } : undefined}
  >
    <Icon size={15} stroke={1.8} />
    {label}
  </a>
);

const ProfileCard = ({ me, onMeUpdate }) => {
  const [name, setName] = useState('');
  const [bio,  setBio]  = useState('');
  const [saving, setSaving] = useState(false);
  const [saved,  setSaved]  = useState(false);
  const [error,  setError]  = useState(null);
  const fileRef = useRef(null);

  useEffect(() => {
    if (!me) return;
    setName(me.displayName || '');
    setBio(me.bio || '');
  }, [me]);

  const dirty = me && (name !== (me.displayName || '') || bio !== (me.bio || ''));

  const save = async () => {
    setSaving(true); setError(null);
    try {
      const updated = await updateMe({
        displayName: name.trim() || undefined,
        bio: bio.trim(),
      });
      onMeUpdate?.(updated);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (e) {
      setError(e.message || 'не удалось сохранить');
    } finally {
      setSaving(false);
    }
  };

  const pickAvatar = async (e) => {
    const f = e.target.files?.[0];
    e.target.value = '';
    if (!f) return;
    try {
      const url = await uploadFile(f);
      const updated = await updateMe({ avatarUrl: url });
      onMeUpdate?.(updated);
    } catch (er) {
      setError(er.message || 'не удалось загрузить');
    }
  };

  const removeAvatar = async () => {
    try {
      const updated = await updateMe({ avatarUrl: '' });
      onMeUpdate?.(updated);
    } catch (er) {
      setError(er.message || 'не удалось убрать аватар');
    }
  };

  const handle = me?.username ?? '';
  const avatarText = (me?.username ?? '??').slice(0, 2).toUpperCase();

  return (
    <div className="settings-card" id="profile">
      <div className="settings-card-head">
        <h2>Профиль</h2>
        <p className="sub">как тебя видят другие садоводы</p>
      </div>

      <div className="settings-row">
        <div className="key">Имя <span className="hint">отображается в постах</span></div>
        <div className="ctl">
          <input className="input-text" value={name} onChange={e => setName(e.target.value)} />
        </div>
        <div className="meta"></div>
      </div>

      <div className="settings-row">
        <div className="key">Логин <span className="hint">@-handle, виден всем</span></div>
        <div className="ctl">
          <span style={{ fontFamily: 'var(--mono)', color: 'var(--text-dim)', fontSize: 13 }}>@</span>
          <input className="input-text mono" value={handle} disabled style={{ opacity: 0.6 }} />
        </div>
        <div className="meta">honeygarden.space/u/{handle}</div>
      </div>

      <div className="settings-row">
        <div className="key">Био <span className="hint">до 200 символов</span></div>
        <div className="ctl">
          <textarea className="input-text" value={bio} onChange={e => setBio(e.target.value)} maxLength={200} />
        </div>
        <div className="meta">{bio.length}/200</div>
      </div>

      <div className="settings-row">
        <div className="key">Аватар</div>
        <div className="ctl">
          <div style={{
            width: 48, height: 48, borderRadius: 12,
            background: me?.avatarUrl ? `center/cover url(${me.avatarUrl})` : 'linear-gradient(135deg,var(--hn-honey-pale),var(--hn-honey-light))',
            color: 'var(--hn-honey-dark)', display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontFamily: 'var(--mono)', fontWeight: 700, fontSize: 15, overflow: 'hidden',
          }}>
            {!me?.avatarUrl && avatarText}
          </div>
          <button className="btn" style={{ fontSize: 12 }} onClick={() => fileRef.current?.click()}>загрузить</button>
          <input ref={fileRef} type="file" accept="image/*" hidden onChange={pickAvatar} />
          {me?.avatarUrl && (
            <button className="btn ghost" style={{ fontSize: 12, color: 'var(--text-dim)' }} onClick={removeAvatar}>убрать</button>
          )}
        </div>
        <div className="meta">PNG / JPG, до 5 МБ</div>
      </div>

      <div className="settings-row">
        <div className="key"></div>
        <div className="ctl"></div>
        <div className="meta" style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          {error && <span style={{ color: 'var(--hn-red)' }}>{error}</span>}
          {saved && <span style={{ color: 'var(--hn-green)', display: 'inline-flex', alignItems: 'center', gap: 4 }}><IconCheck size={12} /> сохранено</span>}
          <button className="btn primary" style={{ fontSize: 12 }} onClick={save} disabled={!dirty || saving}>
            {saving ? 'сохраняем…' : 'сохранить'}
          </button>
        </div>
      </div>
    </div>
  );
};

const AccountCard = () => (
  <div className="settings-card" id="account">
    <div className="settings-card-head">
      <h2>Аккаунт</h2>
      <p className="sub">пароль и двухфакторка управляются через Zitadel</p>
    </div>
    <div className="settings-row">
      <div className="key">Пароль и 2FA</div>
      <div className="ctl" style={{ fontSize: 12.5, color: 'var(--text-mid)' }}>
        управление паролем, 2FA и сессиями — в кабинете Zitadel
      </div>
      <div className="meta">
        <a className="btn" style={{ fontSize: 12, textDecoration: 'none' }}
           href="https://auth.honeygarden.space" target="_blank" rel="noreferrer">
          открыть Zitadel ↗
        </a>
      </div>
    </div>
    <div className="settings-row">
      <div className="key">Email</div>
      <div className="ctl" style={{ fontSize: 12.5, color: 'var(--text-mid)' }}>
        почта используется для уведомлений и восстановления
      </div>
      <div className="meta">в Zitadel</div>
    </div>
  </div>
);

const AppearanceCard = () => {
  const [theme, setTheme] = useState(() => localStorage.getItem('hg.theme') || 'light');
  const [accent, setAccent] = useState(() => localStorage.getItem('hg.accent') || 'honey');
  const [density, setDensity] = useState(() => localStorage.getItem('hg.density') || 'cards');
  const [fontSize, setFontSize] = useState(() => localStorage.getItem('hg.fontSize') || 'm');
  const [serif, setSerif] = useState(() => localStorage.getItem('hg.serif') !== '0');
  const [animations, setAnimations] = useState(() => localStorage.getItem('hg.anim') !== '0');

  const persist = (k, v) => { localStorage.setItem(k, v); applyAppearance(); };

  const accents = [
    { k: 'honey',     color: 'var(--hn-honey-bright)' },
    { k: 'green',     color: '#A3BE8C' },
    { k: 'indigo',    color: '#81A1C1' },
    { k: 'lavender',  color: '#B48EAD' },
    { k: 'terracota', color: '#D08770' },
  ];

  return (
    <div className="settings-card" id="appearance">
      <div className="settings-card-head">
        <h2>Внешний вид</h2>
        <p className="sub">тема, плотность, типографика</p>
      </div>

      <div className="settings-row">
        <div className="key">Тема</div>
        <div className="ctl">
          <div className="theme-wrap">
            {[
              { k: 'light', label: 'light' },
              { k: 'dark',  label: 'dark' },
              { k: 'auto',  label: 'авто' },
            ].map(t => (
              <div key={t.k} className="theme-cell">
                <div
                  className={`theme-swatch ${t.k}` + (theme === t.k ? ' active' : '')}
                  onClick={() => { setTheme(t.k); persist('hg.theme', t.k); }}
                >
                  <div className="bar"></div><div className="bar"></div><div className="bar"></div>
                </div>
                <div className="name">{t.label}</div>
              </div>
            ))}
          </div>
        </div>
        <div className="meta">honey palette</div>
      </div>

      <div className="settings-row">
        <div className="key">Акцент</div>
        <div className="ctl" style={{ gap: 6 }}>
          {accents.map(a => (
            <button
              key={a.k}
              type="button"
              title={a.k}
              onClick={() => { setAccent(a.k); persist('hg.accent', a.k); }}
              style={{
                width: 24, height: 24, borderRadius: '50%', cursor: 'pointer',
                border: `2px solid ${accent === a.k ? a.color : 'var(--bg-card)'}`,
                background: a.color,
                boxShadow: accent === a.k ? `0 0 0 2px var(--bg-card), 0 0 0 4px ${a.color}` : undefined,
              }}
            />
          ))}
        </div>
        <div className="meta">оттенок CTA-кнопок</div>
      </div>

      <div className="settings-row">
        <div className="key">Плотность ленты</div>
        <div className="ctl">
          <Segmented
            value={density}
            onChange={(v) => { setDensity(v); persist('hg.density', v); }}
            options={[
              { value: 'cards',   label: 'Карточки' },
              { value: 'compact', label: 'Компакт' },
            ]}
          />
        </div>
        <div className="meta"></div>
      </div>

      <div className="settings-row">
        <div className="key">Размер шрифта</div>
        <div className="ctl">
          <Segmented
            value={fontSize}
            onChange={(v) => { setFontSize(v); persist('hg.fontSize', v); }}
            options={[
              { value: 's', label: 'S' },
              { value: 'm', label: 'M' },
              { value: 'l', label: 'L' },
            ]}
          />
        </div>
        <div className="meta">текущий: {fontSize === 's' ? '13px' : fontSize === 'l' ? '16px' : '14.5px'}</div>
      </div>

      <div className="settings-row">
        <div className="key">Серифы в заголовках <span className="hint">Source Serif 4</span></div>
        <div className="ctl" style={{ justifyContent: 'flex-end' }}>
          <Toggle checked={serif} onChange={(v) => { setSerif(v); persist('hg.serif', v ? '1' : '0'); }} />
        </div>
        <div className="meta">убери, если режет глаз</div>
      </div>

      <div className="settings-row">
        <div className="key">Анимации</div>
        <div className="ctl" style={{ justifyContent: 'flex-end' }}>
          <Toggle checked={animations} onChange={(v) => { setAnimations(v); persist('hg.anim', v ? '1' : '0'); }} />
        </div>
        <div className="meta">пульс, hover-эффекты</div>
      </div>
    </div>
  );
};

const NOTIF_ROWS = [
  { type: 'comment.created',    label: 'Ответ на мой пост',    desc: 'кто-то отвечает в треде, который ты завёл' },
  { type: 'comment.reply',      label: 'Ответ на коммент',     desc: 'кто-то отвечает на твой комментарий' },
  { type: 'mention',            label: 'Упоминания @',         desc: 'тебя зовут в обсуждение' },
  { type: 'post.voted',         label: 'Голос за пост',        desc: 'апвоут на твой пост' },
  { type: 'user.followed',      label: 'Новый подписчик',      desc: 'кто-то подписался на тебя' },
  { type: 'community.followed', label: 'Подписки на сообщества', desc: 'новый участник в моём сообществе' },
];

const NotificationsCard = () => {
  const [prefs, setPrefs] = useState({});
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    let cancelled = false;
    getNotificationPreferences().then(list => {
      if (cancelled) return;
      const m = {};
      (list ?? []).forEach(p => { m[p.type] = p.enabled; });
      setPrefs(m);
      setLoaded(true);
    }).catch(() => setLoaded(true));
    return () => { cancelled = true; };
  }, []);

  const toggle = (type) => {
    const next = !(prefs[type] ?? true);
    setPrefs(p => ({ ...p, [type]: next }));
    setNotificationPreference(type, next).catch(() => {});
  };

  return (
    <div className="settings-card" id="notifications">
      <div className="settings-card-head">
        <h2>Уведомления</h2>
        <p className="sub">что присылать в колокольчик</p>
      </div>
      {!loaded ? (
        <div className="settings-row"><div className="key" style={{ color: 'var(--text-dim)' }}>загрузка…</div><div></div><div></div></div>
      ) : NOTIF_ROWS.map(r => (
        <div key={r.type} className="settings-row">
          <div className="key">{r.label}</div>
          <div className="ctl" style={{ fontSize: 12.5, color: 'var(--text-dim)' }}>{r.desc}</div>
          <div className="meta" style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Toggle checked={prefs[r.type] ?? true} onChange={() => toggle(r.type)} />
          </div>
        </div>
      ))}
    </div>
  );
};

const FeedCard = ({ me, onMeUpdate }) => {
  const [interests, setInterests] = useState(() => new Set(me?.tags ?? []));
  const [savingInt, setSavingInt] = useState(false);
  const [intSaved, setIntSaved] = useState(false);
  const [intError, setIntError] = useState(null);

  useEffect(() => {
    if (me?.tags) setInterests(new Set(me.tags));
  }, [me?.tags?.join(',')]);

  const toggleInterest = (id) => {
    setInterests(prev => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  };

  const saveInterests = async () => {
    setSavingInt(true);
    setIntError(null);
    try {
      const tags = [...interests];
      await setUserTags(tags);
      onMeUpdate?.({ ...(me ?? {}), tags });
      setIntSaved(true);
      setTimeout(() => setIntSaved(false), 2000);
    } catch (e) {
      setIntError(e.message ?? 'не удалось сохранить');
    } finally {
      setSavingInt(false);
    }
  };

  const interestsDirty = me && (
    interests.size !== (me.tags?.length ?? 0) ||
    [...interests].some(t => !(me.tags ?? []).includes(t))
  );

  const [sort, setSort] = useState(() => localStorage.getItem('hg.feedMode') || 'forme');
  const [showRSS, setShowRSS] = useState(() => localStorage.getItem('hg.showRSS') !== '0');
  const [linkTarget, setLinkTarget] = useState(() => localStorage.getItem('hg.linkTarget') || 'blank');
  const [hiddenTags, setHiddenTags] = useState(() => {
    try { return JSON.parse(localStorage.getItem('hg.hiddenTags') || '[]'); } catch { return []; }
  });
  const [draft, setDraft] = useState('');

  const persistTags = (next) => {
    setHiddenTags(next);
    localStorage.setItem('hg.hiddenTags', JSON.stringify(next));
  };

  const addTag = () => {
    const t = draft.trim().replace(/^#/, '').toLowerCase();
    if (!t || hiddenTags.includes(t)) { setDraft(''); return; }
    persistTags([...hiddenTags, t]);
    setDraft('');
  };

  const removeTag = (t) => persistTags(hiddenTags.filter(x => x !== t));

  return (
    <div className="settings-card" id="feed">
      <div className="settings-card-head">
        <h2>Лента</h2>
        <p className="sub">что попадает в «для меня»</p>
      </div>

      <div className="settings-row">
        <div className="key">Сортировка по умолчанию</div>
        <div className="ctl">
          <select
            className="input-select"
            value={sort}
            onChange={(e) => { setSort(e.target.value); localStorage.setItem('hg.feedMode', e.target.value); }}
          >
            <option value="forme">Для меня</option>
            <option value="trending">Популярное</option>
            <option value="new">Свежее</option>
            <option value="bookmarks">Сохранённые</option>
          </select>
        </div>
        <div className="meta"></div>
      </div>

      <div className="settings-row">
        <div className="key">Показывать RSS-посты</div>
        <div className="ctl" style={{ justifyContent: 'flex-end' }}>
          <Toggle checked={showRSS} onChange={(v) => { setShowRSS(v); localStorage.setItem('hg.showRSS', v ? '1' : '0'); }} />
        </div>
        <div className="meta">статьи из подключённых лент</div>
      </div>

      <div className="settings-row">
        <div className="key">Открывать ссылки</div>
        <div className="ctl">
          <Segmented
            value={linkTarget}
            onChange={(v) => { setLinkTarget(v); localStorage.setItem('hg.linkTarget', v); }}
            options={[
              { value: 'blank', label: 'в новой вкладке' },
              { value: 'self',  label: 'в этой' },
            ]}
          />
        </div>
        <div className="meta"></div>
      </div>

      <div className="settings-row">
        <div className="key">Интересы <span className="hint">определяют «Для меня» и тренды</span></div>
        <div className="ctl" style={{ flexWrap: 'wrap', gap: 6, alignItems: 'flex-start' }}>
          {INTEREST_TAGS.map(t => {
            const on = interests.has(t.id);
            return (
              <button
                key={t.id}
                type="button"
                onClick={() => toggleInterest(t.id)}
                className={'tag-pill' + (on ? ' active' : '')}
                style={{
                  cursor: 'pointer', userSelect: 'none',
                  background: on ? 'var(--hn-honey-pale)' : 'var(--bg-sunk)',
                  color: on ? 'var(--hn-honey-dark)' : 'var(--text-mid)',
                  borderColor: on ? 'var(--hn-honey-light)' : 'var(--border)',
                }}
              >
                #{t.id}
              </button>
            );
          })}
        </div>
        <div className="meta" style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 6 }}>
          {intError && <span style={{ color: 'var(--hn-red)', fontSize: 11.5 }}>{intError}</span>}
          {intSaved && <span style={{ color: 'var(--hn-green)', fontSize: 11.5, display: 'inline-flex', alignItems: 'center', gap: 4 }}><IconCheck size={12} /> сохранено</span>}
          <button className="btn primary" style={{ fontSize: 12 }} onClick={saveInterests} disabled={!interestsDirty || savingInt}>
            {savingInt ? 'сохраняем…' : 'сохранить'}
          </button>
        </div>
      </div>

      <div className="settings-row">
        <div className="key">Скрытые теги <span className="hint">не покажутся в ленте</span></div>
        <div className="ctl" style={{ flexWrap: 'wrap', gap: 4 }}>
          {hiddenTags.map(t => (
            <span key={t} className="tag-pill">
              #{t} <span className="x" onClick={() => removeTag(t)}>×</span>
            </span>
          ))}
          <input
            className="tag-input"
            placeholder="добавить тег…"
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ',' || e.key === ' ') { e.preventDefault(); addTag(); } }}
            onBlur={addTag}
            style={{ flex: 1, minWidth: 120 }}
          />
        </div>
        <div className="meta"></div>
      </div>
    </div>
  );
};

const PrivacyCard = ({ me, onMeUpdate }) => {
  const [privacy, setPrivacy] = useState(me?.privacy || 'public');
  const [showActivity, setShowActivity] = useState(me?.showActivity ?? true);

  useEffect(() => { if (me) { setPrivacy(me.privacy || 'public'); setShowActivity(me.showActivity ?? true); } }, [me]);

  const updatePrivacy = async (v) => {
    setPrivacy(v);
    try {
      const updated = await updateMe({ privacy: v });
      onMeUpdate?.(updated);
    } catch {}
  };

  return (
    <div className="settings-card" id="privacy">
      <div className="settings-card-head">
        <h2>Приватность</h2>
        <p className="sub">кто видит твой профиль и активность</p>
      </div>

      <div className="settings-row">
        <div className="key">Видимость профиля</div>
        <div className="ctl">
          <Segmented
            value={privacy}
            onChange={updatePrivacy}
            options={[
              { value: 'public',  label: 'Публичный' },
              { value: 'private', label: 'Приватный' },
            ]}
          />
        </div>
        <div className="meta">приватный — только подписчики</div>
      </div>

      <div className="settings-row">
        <div className="key">Активность на профиле</div>
        <div className="ctl" style={{ justifyContent: 'flex-end' }}>
          <Toggle
            checked={showActivity}
            onChange={async (v) => { setShowActivity(v); try { const u = await updateMe({ showActivity: v }); onMeUpdate?.(u); } catch {} }}
          />
        </div>
        <div className="meta">heatmap и список действий</div>
      </div>
    </div>
  );
};

const PROVIDERS = ['github'];

const PinnedReposManager = () => {
  const [me, setMeUser] = useState(null);
  const [repos, setRepos] = useState(null);
  const [pinned, setPinned] = useState(new Set());
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState(null);
  const [okMsg, setOkMsg] = useState(null);

  const reload = async () => {
    setLoading(true);
    setErr(null);
    try {
      const meData = await getMe();
      setMeUser(meData);
      const [list, currentPins] = await Promise.all([
        listIntegrationRepos('github'),
        getUserPinnedRepos(meData.username).catch(() => []),
      ]);
      setRepos(list ?? []);
      setPinned(new Set((currentPins ?? []).map(r => r.externalId)));
    } catch (e) {
      setErr(e.message ?? 'не удалось загрузить репы');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { reload(); }, []);

  const toggle = (externalId) => {
    setPinned(prev => {
      const next = new Set(prev);
      if (next.has(externalId)) next.delete(externalId);
      else if (next.size < 6) next.add(externalId);
      return next;
    });
  };

  const save = async () => {
    setSaving(true);
    setErr(null);
    setOkMsg(null);
    try {
      const pins = [...pinned].map(externalId => ({ provider: 'github', externalId }));
      await setPinnedRepos(pins);
      setOkMsg('сохранено');
      setTimeout(() => setOkMsg(null), 2500);
    } catch (e) {
      setErr(e.message ?? 'не удалось сохранить');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="pinned-repos-block">
      <div className="pinned-repos-head">
        <div className="ttl">Закреплённые репозитории</div>
        <div className="sub">выбери до 6 репозиториев — они появятся на твоём профиле</div>
      </div>

      {err && <div className="pinned-repos-err">{err}</div>}

      {loading || !repos ? (
        <div className="pinned-repos-empty">грузим репы…</div>
      ) : repos.length === 0 ? (
        <div className="pinned-repos-empty">нет публичных репов</div>
      ) : (
        <>
          <div className="pinned-repos-list">
            {repos.map(r => {
              const checked = pinned.has(r.externalId);
              const disabled = !checked && pinned.size >= 6;
              return (
                <label
                  key={r.externalId}
                  className={'pinned-repo-row' + (checked ? ' checked' : '') + (disabled ? ' disabled' : '')}
                >
                  <input
                    type="checkbox"
                    checked={checked}
                    disabled={disabled}
                    onChange={() => toggle(r.externalId)}
                  />
                  <div className="info">
                    <div className="name">{r.name}</div>
                    {r.description && <div className="desc">{r.description}</div>}
                  </div>
                  <div className="meta">
                    {r.language && <span>{r.language} · </span>}
                    <IconStar size={13} stroke={1.8} style={{ verticalAlign: 'text-bottom' }} /> {r.starsCount ?? 0}
                  </div>
                </label>
              );
            })}
          </div>
          <div className="pinned-repos-foot">
            <button className="btn primary" onClick={save} disabled={saving}>
              {saving ? <IconLoader2 size={12} className="spin" /> : 'сохранить'}
            </button>
            <span className="counter">{pinned.size}/6 выбрано</span>
            {okMsg && <span className="ok">{okMsg}</span>}
          </div>
        </>
      )}
    </div>
  );
};

const ConnectionsCard = () => {
  const { me } = useMe();
  const [accounts, setAccounts] = useState([]);
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState(null);

  // Wakapi
  const [wakaUrl,  setWakaUrl]  = useState('https://waka.honeygarden.space');
  const [wakaUser, setWakaUser] = useState('');
  const [wakaKey,  setWakaKey]  = useState('');
  const [wakaConnected, setWakaConnected] = useState(false);
  const [wakaSaving, setWakaSaving] = useState(false);
  const [wakaForm,   setWakaForm]   = useState(false);
  const [wakaError,  setWakaError]  = useState(null);

  const load = async () => {
    try {
      const accs = await listIntegrations();
      setAccounts(accs ?? []);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoaded(true);
    }
    if (me?.username) {
      try {
        const w = await getUserWakapi(me.username);
        setWakaConnected(!!w?.connected);
      } catch {}
    }
  };

  useEffect(() => { load(); }, [me?.username]);

  useEffect(() => {
    const url = new URL(window.location.href);
    const status = url.searchParams.get('status');
    const integration = url.searchParams.get('integration');
    if (status && integration) {
      url.searchParams.delete('status');
      url.searchParams.delete('integration');
      url.searchParams.delete('msg');
      window.history.replaceState({}, '', url);
      if (status === 'ok') load();
    }
  }, []);

  const accountFor = (p) => accounts.find(a => a.provider === p);

  const connect = async (provider) => {
    try {
      const { authUrl } = await beginIntegrationConnect(provider);
      window.location.assign(authUrl);
    } catch (e) {
      setError(e.message);
    }
  };

  const disconnect = async (provider) => {
    if (!confirm(`Отвязать ${PROVIDER_LABELS[provider]}?`)) return;
    try {
      await disconnectIntegration(provider);
      setAccounts(a => a.filter(x => x.provider !== provider));
    } catch (e) {
      setError(e.message);
    }
  };

  const wakaConnect = async () => {
    if (!wakaKey.trim() || !wakaUser.trim() || !wakaUrl.trim()) return;
    setWakaSaving(true);
    setWakaError(null);
    try {
      await connectWakapi({ instanceUrl: wakaUrl.trim(), apiKey: wakaKey.trim(), username: wakaUser.trim() });
      setWakaConnected(true);
      setWakaForm(false);
      setWakaKey('');
    } catch (e) {
      setWakaError(e.message ?? 'не удалось подключить (проверь URL/username/key)');
    } finally {
      setWakaSaving(false);
    }
  };

  const wakaDisconnect = async () => {
    if (!confirm('Отвязать Wakapi?')) return;
    try {
      await disconnectWakapi();
      setWakaConnected(false);
      setWakaUser(''); setWakaKey('');
    } catch (e) {
      setWakaError(e.message);
    }
  };

  return (
    <div className="settings-card" id="connections">
      <div className="settings-card-head">
        <h2>Подключения</h2>
        <p className="sub">привязанные аккаунты git-провайдеров</p>
      </div>

      {error && (
        <div className="settings-row">
          <div className="key" style={{ color: 'var(--hn-red)' }}>ошибка</div>
          <div className="ctl" style={{ color: 'var(--hn-red)', fontSize: 12.5 }}>{error}</div>
          <div></div>
        </div>
      )}

      {PROVIDERS.map(p => {
        const acc = accountFor(p);
        return (
          <div key={p} className="conn-row">
            <div className="conn-icon"><IconBrandGithub size={18} stroke={1.7} /></div>
            <div className="conn-info">
              <div className="name">{PROVIDER_LABELS[p]}</div>
              <div className="handle">
                {acc ? `@${acc.username}` : 'привяжи аккаунт для закреплённых репозиториев'}
              </div>
            </div>
            {acc ? (
              <button className="btn" style={{ fontSize: 12 }} onClick={() => disconnect(p)}>
                отвязать
              </button>
            ) : (
              <button className="btn primary" style={{ fontSize: 12 }} onClick={() => connect(p)}>
                подключить
              </button>
            )}
          </div>
        );
      })}

      {accountFor('github') && <PinnedReposManager />}

      <div className="conn-row">
        <div className="conn-icon" style={{ background: '#1F1A12', color: 'var(--hn-honey-bright)', fontFamily: 'var(--mono)', fontSize: 11, fontWeight: 700 }}>WK</div>
        <div className="conn-info">
          <div className="name">Wakapi</div>
          <div className="handle">{wakaConnected ? 'подключено' : 'статистика кодинга в профиле'}</div>
        </div>
        {wakaConnected ? (
          <button className="btn" style={{ fontSize: 12 }} onClick={wakaDisconnect}>отвязать</button>
        ) : (
          <button
            className={'btn' + (wakaForm ? '' : ' primary')}
            style={{ fontSize: 12 }}
            onClick={() => setWakaForm(v => !v)}
          >
            {wakaForm ? 'отмена' : 'подключить'}
          </button>
        )}
      </div>

      {wakaForm && !wakaConnected && (
        <div className="pinned-repos-block">
          <div className="pinned-repos-head">
            <div className="ttl">Подключение Wakapi</div>
            <div className="sub">URL инстанса (с /api или без — без разницы), твой username на Wakapi, API-ключ (UUID из Wakapi → Settings → API key)</div>
          </div>
          {wakaError && <div className="pinned-repos-err">{wakaError}</div>}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            <input
              className="input-text mono"
              value={wakaUrl}
              onChange={e => setWakaUrl(e.target.value)}
              placeholder="https://waka.honeygarden.space"
            />
            <input
              className="input-text mono"
              value={wakaUser}
              onChange={e => setWakaUser(e.target.value)}
              placeholder="username (твой логин в Wakapi)"
            />
            <input
              className="input-text mono"
              type="password"
              value={wakaKey}
              onChange={e => setWakaKey(e.target.value)}
              placeholder="API key (UUID)"
            />
          </div>
          <div className="pinned-repos-foot">
            <button
              className="btn primary"
              onClick={wakaConnect}
              disabled={wakaSaving || !wakaKey.trim() || !wakaUser.trim() || !wakaUrl.trim()}
            >
              {wakaSaving ? <IconLoader2 size={12} className="spin" /> : 'подключить'}
            </button>
          </div>
        </div>
      )}

      {!loaded && (
        <div className="settings-row">
          <div className="key" style={{ color: 'var(--text-dim)' }}>загрузка…</div>
          <div></div><div></div>
        </div>
      )}
    </div>
  );
};

const DangerCard = () => (
  <div className="settings-card" id="danger" style={{ borderColor: 'rgba(191,97,106,.25)' }}>
    <div className="settings-card-head danger">
      <h2>Опасная зона</h2>
      <p className="sub">действия здесь нельзя откатить</p>
    </div>
    <div className="settings-row">
      <div className="key">Экспорт данных</div>
      <div className="ctl" style={{ fontSize: 12.5, color: 'var(--text-mid)' }}>
        скачать архив с твоими постами, комментариями и подписками
      </div>
      <div><button className="btn" style={{ fontSize: 12 }} disabled>скоро</button></div>
    </div>
    <div className="settings-row">
      <div className="key">Удалить аккаунт</div>
      <div className="ctl" style={{ fontSize: 12.5, color: 'var(--text-mid)' }}>
        управляется через Zitadel — удаление аккаунта там удалит и профиль HoneyGarden
      </div>
      <div>
        <a className="btn btn-outline-danger" style={{ fontSize: 12, textDecoration: 'none' }}
           href="https://auth.honeygarden.space" target="_blank" rel="noreferrer">
          в Zitadel ↗
        </a>
      </div>
    </div>
  </div>
);

const SettingsScreen = () => {
  const { me, setMe } = useMe();
  const [active, setActive] = useState('profile');
  const onMeUpdate = (next) => setMe?.(next);

  const scrollTo = (id) => {
    setActive(id === 'danger' ? 'danger' : id);
    const el = document.getElementById(id);
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
  };

  useEffect(() => {
    const sectionIds = [...SECTIONS.map(s => s.id), 'danger'];
    const onScroll = () => {
      let current = active;
      for (const id of sectionIds) {
        const el = document.getElementById(id);
        if (!el) continue;
        const { top } = el.getBoundingClientRect();
        if (top <= 140) current = id;
      }
      if (current !== active) setActive(current);
    };
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, [active]);

  return (
    <div className="settings-shell">
      <aside className="settings-nav">
        <h2 className="settings-nav-title">Настройки</h2>
        {SECTIONS.map(s => (
          <NavLink
            key={s.id}
            id={s.id}
            label={s.label}
            Icon={s.Icon}
            active={active === s.id}
            onClick={scrollTo}
          />
        ))}
        <div style={{ height: 8 }}></div>
        <NavLink
          id="danger"
          label="Опасная зона"
          Icon={IconTrash}
          active={active === 'danger'}
          onClick={scrollTo}
          danger
        />
      </aside>

      <main className="settings-main">
        <ProfileCard me={me} onMeUpdate={onMeUpdate} />
        <AccountCard />
        <AppearanceCard />
        <NotificationsCard />
        <FeedCard me={me} onMeUpdate={onMeUpdate} />
        <PrivacyCard me={me} onMeUpdate={onMeUpdate} />
        <ConnectionsCard />
        <DangerCard />
      </main>
    </div>
  );
};

export default SettingsScreen;
