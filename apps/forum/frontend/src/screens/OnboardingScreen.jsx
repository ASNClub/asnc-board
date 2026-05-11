import { useState, useEffect, useRef } from 'react';
import {
  IconCheck, IconCamera,
  IconChevronRight, IconChevronLeft, IconPlus,
} from '@tabler/icons-react';
import { updateMe, setUserTags, uploadFile, searchCommunities, joinCommunity } from '../lib/api';
import { commColor } from '../lib/utils';
import { INTEREST_TAGS } from '../data/tags';

const STEPS = [
  { k: 'profile',     l: 'Профиль' },
  { k: 'interests',   l: 'Интересы' },
  { k: 'communities', l: 'Сообщества' },
];

const initials = (s) => (s || '?').slice(0, 2).toUpperCase();

const OnboardingScreen = ({ displayNameFromIdp = '', onDone }) => {
  const [step,        setStep]        = useState(0);
  const [username,    setUsername]    = useState('');
  const [displayName, setDisplayName] = useState(displayNameFromIdp);
  const [bio,         setBio]         = useState('');
  const [avatarUrl,   setAvatarUrl]   = useState(null);
  const [tags,        setTags]        = useState([]);
  const [joined,      setJoined]      = useState(new Set());
  const [suggested,   setSuggested]   = useState([]);
  const [error,       setError]       = useState('');
  const [loading,     setLoading]     = useState(false);
  const [uploading,   setUploading]   = useState(false);
  const fileRef = useRef(null);

  const toggleTag = (id) =>
    setTags(prev => prev.includes(id) ? prev.filter(t => t !== id) : [...prev, id]);

  const toggleJoin = (slug) =>
    setJoined(prev => {
      const next = new Set(prev);
      next.has(slug) ? next.delete(slug) : next.add(slug);
      return next;
    });

  const onAvatarFile = async (e) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    try {
      const url = await uploadFile(file);
      setAvatarUrl(url);
    } catch {
      setError('не получилось загрузить аватар');
    } finally {
      setUploading(false);
    }
  };

  // Подгружаем сообщества по выбранным тегам, когда переходим на шаг 3
  useEffect(() => {
    if (step !== 2 || tags.length === 0) return;
    let cancelled = false;
    (async () => {
      const seen = new Map();
      for (const tag of tags.slice(0, 5)) {
        try {
          const res = await searchCommunities(tag);
          const list = Array.isArray(res?.hits) ? res.hits : (Array.isArray(res) ? res : []);
          for (const c of list) {
            if (!seen.has(c.slug)) seen.set(c.slug, c);
          }
        } catch { /* ignore */ }
      }
      if (!cancelled) setSuggested([...seen.values()].slice(0, 8));
    })();
    return () => { cancelled = true };
  }, [step, tags]);

  const validProfile =
    /^[a-z0-9][a-z0-9_-]{1,31}$/.test(username) &&
    displayName.trim().length >= 1;

  const handleNext = () => {
    setError('');
    if (step === 0 && !validProfile) {
      setError('username минимум 2 символа, латиница / цифры / _ / -');
      return;
    }
    setStep(s => Math.min(s + 1, STEPS.length - 1));
  };
  const handleBack = () => { setError(''); setStep(s => Math.max(0, s - 1)); };

  const handleFinish = async () => {
    setError('');
    setLoading(true);
    try {
      await updateMe({
        username:       username.trim(),
        displayName:    displayName.trim(),
        bio:            bio.trim() || null,
        avatarUrl:      avatarUrl,
        onboardingDone: true,
      });
      if (tags.length > 0) await setUserTags(tags);
      for (const slug of joined) {
        await joinCommunity(slug).catch(() => {});
      }
      onDone?.();
    } catch (e) {
      setError(e.message ?? 'ошибка сохранения');
    } finally {
      setLoading(false);
    }
  };

  const progressPct = Math.round(((step + 1) / STEPS.length) * 100);

  return (
    <div className="onb-shell">
      <aside className="onb-side">
        <h1>Настроим <span className="amp">сад</span></h1>
        <ul className="onb-steps">
          {STEPS.map((s, i) => {
            const cls = i < step ? 'done' : i === step ? 'active' : '';
            return (
              <li key={s.k} className={'onb-step ' + cls}>
                <span className="num">{i < step ? <IconCheck size={12} stroke={3} /> : i + 1}</span>
                <span className="lbl">{s.l}</span>
              </li>
            );
          })}
        </ul>
        <button className="onb-side-skip" onClick={onDone}>пропустить — настрою позже</button>
      </aside>

      <main className="onb-main">
        {step === 0 && (
          <>
            <div className="onb-head">
              <h2>Расскажи о себе</h2>
              <p className="sub">Имя видят все. Email контролирует Zitadel — здесь не показывается.</p>
            </div>
            <div className="onb-body">
              <div className="onb-avatar-row">
                <div className="onb-avatar" onClick={() => fileRef.current?.click()}>
                  {avatarUrl
                    ? <img src={avatarUrl} alt="" />
                    : initials(displayName || username)}
                  <span className="edit"><IconCamera size={13} stroke={2} /></span>
                </div>
                <input
                  ref={fileRef}
                  type="file"
                  accept="image/jpeg,image/png,image/webp"
                  style={{ display: 'none' }}
                  onChange={onAvatarFile}
                />
                <div className="onb-avatar-info">
                  <b>Аватар {uploading && '(загружаем…)'}</b>
                  JPG / PNG / WebP, до 5 МБ. Hex-форма обрежет автоматически.
                </div>
              </div>

              <div className="onb-field">
                <label className="onb-field-label">Отображаемое имя <span className="req">*</span></label>
                <input
                  className="onb-input"
                  value={displayName}
                  onChange={e => setDisplayName(e.target.value)}
                  placeholder="Имя или ник"
                  maxLength={64}
                  autoFocus
                />
              </div>

              <div className="onb-field">
                <label className="onb-field-label">@handle <span className="req">*</span></label>
                <input
                  className="onb-input"
                  style={{ fontFamily: 'var(--mono)' }}
                  value={username}
                  onChange={e => setUsername(e.target.value.toLowerCase().replace(/[^a-z0-9_-]/g, ''))}
                  placeholder="ваш_handle"
                  maxLength={32}
                />
              </div>

              <div className="onb-field">
                <label className="onb-field-label">
                  Био
                  <span className="cnt">{bio.length} / 200</span>
                </label>
                <textarea
                  className="onb-input"
                  value={bio}
                  onChange={e => setBio(e.target.value.slice(0, 200))}
                  placeholder="Чем занимаешься? Что любишь?"
                />
              </div>

              {error && <div className="onb-error">{error}</div>}
            </div>
          </>
        )}

        {step === 1 && (
          <>
            <div className="onb-head">
              <h2>Что тебе интересно?</h2>
              <p className="sub">Выбери от 3 до 10 тегов — лента подстроится. Можешь поменять в настройках.</p>
            </div>
            <div className="onb-body">
              <div className="onb-tags">
                {INTEREST_TAGS.map(t => {
                  const active = tags.includes(t.id);
                  return (
                    <button
                      key={t.id}
                      type="button"
                      className={'onb-tag-pick' + (active ? ' active' : '')}
                      onClick={() => toggleTag(t.id)}
                    >
                      {active && <IconCheck size={12} stroke={3} />}
                      <span className="hash">#</span>{t.l}
                    </button>
                  );
                })}
              </div>
              {error && <div className="onb-error" style={{ marginTop: 12 }}>{error}</div>}
            </div>
          </>
        )}

        {step === 2 && (
          <>
            <div className="onb-head">
              <h2>На что подписаться</h2>
              <p className="sub">Сообщества по выбранным тегам. Можно отметить несколько или пропустить.</p>
            </div>
            <div className="onb-body">
              {suggested.length === 0 ? (
                <div style={{ padding: 30, textAlign: 'center', color: 'var(--text-dim)', fontFamily: 'var(--mono)', fontSize: 12 }}>
                  {tags.length === 0 ? 'выбери теги на предыдущем шаге, чтобы увидеть рекомендации' : 'ищем подходящие сообщества…'}
                </div>
              ) : (
                <div className="onb-comm-grid">
                  {suggested.map(c => {
                    const isJoined = joined.has(c.slug);
                    return (
                      <button
                        key={c.slug}
                        type="button"
                        className={'onb-comm-card' + (isJoined ? ' joined' : '')}
                        onClick={() => toggleJoin(c.slug)}
                      >
                        <span className="ico" style={{ background: commColor(c.slug) }}>
                          {(c.slug || '?').slice(0, 2).toLowerCase()}
                        </span>
                        <div>
                          <div className="info-name">{c.name || c.slug}</div>
                          <div className="info-meta">hg/{c.slug} · {(c.followersCount ?? 0).toLocaleString('ru')} участников</div>
                        </div>
                        <span className="toggle-mini">
                          {isJoined ? <IconCheck size={14} stroke={2.5} /> : <IconPlus size={14} stroke={2} />}
                        </span>
                      </button>
                    );
                  })}
                </div>
              )}
              {error && <div className="onb-error" style={{ marginTop: 12 }}>{error}</div>}
            </div>
          </>
        )}

        <div className="onb-foot">
          <button
            className="btn ghost"
            onClick={handleBack}
            disabled={step === 0}
            style={step === 0 ? { opacity: .4, cursor: 'not-allowed' } : {}}
          >
            <IconChevronLeft size={14} stroke={2} /> Назад
          </button>
          <div className="progress"><div className="bar" style={{ width: `${progressPct}%` }} /></div>
          <span className="meta">{step + 1} из {STEPS.length}</span>
          {step < STEPS.length - 1 ? (
            <button className="btn primary" onClick={handleNext}>
              Дальше <IconChevronRight size={14} stroke={2} />
            </button>
          ) : (
            <button className="btn primary" onClick={handleFinish} disabled={loading}>
              {loading ? 'сохраняем…' : 'Готово →'}
            </button>
          )}
        </div>
      </main>
    </div>
  );
};

export default OnboardingScreen;
