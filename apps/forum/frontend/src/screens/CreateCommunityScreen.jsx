import { useState, useRef, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  IconCheck, IconPhoto, IconPlus, IconX,
} from '@tabler/icons-react';
import { createCommunity, uploadFile } from '../lib/api';
import ImageCropModal from '../components/ImageCropModal';

const COLORS = ['#BF616A', '#D08770', '#EBCB8B', '#A3BE8C', '#8FBCBB', '#81A1C1', '#B48EAD', '#E09832'];

const slugify = (s) =>
  s.toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '')
    .trim()
    .replace(/\s+/g, '-')
    .slice(0, 32);

const CreateCommunityScreen = () => {
  const navigate = useNavigate();
  const [slug, setSlug]                 = useState('');
  const [slugTouched, setSlugTouched]   = useState(false);
  const [name, setName]                 = useState('');
  const [description, setDescription]   = useState('');
  const [color, setColor]               = useState(COLORS[6]);
  const [bannerUrl, setBannerUrl]       = useState(null);
  const [bannerUploading, setBannerUp]  = useState(false);
  const [rules, setRules]               = useState(['']);
  const [submitting, setSubmitting]     = useState(false);
  const [error, setError]               = useState(null);
  const [cropState, setCropState]       = useState(null);
  const fileRef = useRef(null);

  const onName = (v) => {
    setName(v);
    if (!slugTouched) setSlug(slugify(v));
  };

  const onSlug = (v) => {
    setSlugTouched(true);
    setSlug(v.toLowerCase().replace(/[^a-z0-9_-]/g, '').slice(0, 32));
  };

  const onBannerFile = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setCropState({ src: URL.createObjectURL(file) });
    e.target.value = '';
  };

  const handleCropDone = async (file) => {
    setCropState(null);
    setBannerUp(true);
    try {
      const url = await uploadFile(file);
      setBannerUrl(url);
    } catch {
      setError('не получилось загрузить баннер');
    } finally {
      setBannerUp(false);
    }
  };

  const setRule = (i, v) =>
    setRules(r => r.map((x, idx) => idx === i ? v : x));
  const addRule = () => setRules(r => [...r, '']);
  const removeRule = (i) => setRules(r => r.filter((_, idx) => idx !== i));

  const valid =
    /^[a-z0-9][a-z0-9_-]{2,31}$/.test(slug) &&
    name.trim().length >= 2;

  const submit = async () => {
    if (!valid || submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      const cleanRules = rules.map(r => r.trim()).filter(Boolean);
      const created = await createCommunity({
        slug,
        name: name.trim(),
        description: description.trim() || null,
        rules: cleanRules,
        bannerUrl: bannerUrl || null,
      });
      navigate(`/c/${created.slug || slug}`);
    } catch (err) {
      setError(err.message ?? 'не удалось создать');
      setSubmitting(false);
    }
  };

  const previewIcon = useMemo(
    () => (slug || name || '?').slice(0, 2).toLowerCase(),
    [slug, name],
  );

  return (
    <div className="cnew-shell">
      <div className="cnew-main">
        <div className="cnew-head">
          <h1>Новое <span className="amp">сообщество</span></h1>
          <p className="sub">Один аккаунт = одно сообщество. Slug менять нельзя — выбирай вдумчиво.</p>
        </div>

        <div className="cnew-card">
          <div className="cnew-card-head">
            <h3>Идентификация</h3>
            <span className="step">шаг 1 из 3</span>
          </div>
          <div className="cnew-card-body">
            <div className="cnew-row">
              <div className="key">Название</div>
              <input
                className="cnew-input"
                value={name}
                onChange={e => onName(e.target.value)}
                placeholder="Distributed Systems"
                maxLength={64}
                autoFocus
              />
            </div>
            <div className="cnew-row">
              <div className="key">Slug <span className="hint">/c/&lt;имя&gt;, нельзя менять</span></div>
              <div>
                <div className="cnew-slug-input">
                  <span className="prefix">hg/</span>
                  <input
                    value={slug}
                    onChange={e => onSlug(e.target.value)}
                    placeholder="distributed"
                    maxLength={32}
                  />
                </div>
                {slug.length >= 3 && (
                  <div style={{ fontFamily: 'var(--mono)', fontSize: 11, color: valid ? 'var(--hn-green)' : 'var(--text-dim)', marginTop: 6, display: 'flex', alignItems: 'center', gap: 5 }}>
                    {valid && <IconCheck size={11} stroke={2.5} />}
                    {valid ? 'формат ОК' : 'минимум 3 символа, латиница / цифры / _ / -'}
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>

        <div className="cnew-card">
          <div className="cnew-card-head">
            <h3>Внешний вид</h3>
            <span className="step">шаг 2 из 3</span>
          </div>
          <div className="cnew-card-body">
            <div className="cnew-row">
              <div className="key">Цвет иконки</div>
              <div className="cnew-color-row">
                {COLORS.map(c => (
                  <button
                    key={c}
                    type="button"
                    className={'cnew-color' + (color === c ? ' active' : '')}
                    style={{ background: c }}
                    onClick={() => setColor(c)}
                  >
                    <IconCheck size={13} stroke={3} />
                  </button>
                ))}
              </div>
            </div>
            <div className="cnew-row">
              <div className="key">Баннер <span className="hint">1200×300, JPG/PNG, до 5 МБ</span></div>
              <div
                className="cnew-banner-up"
                onClick={() => fileRef.current?.click()}
                style={bannerUrl ? { backgroundImage: `url(${bannerUrl})`, borderColor: 'var(--hn-honey-light)' } : undefined}
              >
                {!bannerUrl && (
                  <>
                    <span className="ico"><IconPhoto size={24} stroke={2} /></span>
                    <span className="lbl">{bannerUploading ? 'загружаем…' : 'Загрузить баннер'}</span>
                    <span className="hint">или перетащи сюда</span>
                  </>
                )}
              </div>
              <input
                ref={fileRef}
                type="file"
                accept="image/jpeg,image/png,image/webp"
                style={{ display: 'none' }}
                onChange={onBannerFile}
              />
            </div>
          </div>
        </div>

        <div className="cnew-card">
          <div className="cnew-card-head">
            <h3>Описание и правила</h3>
            <span className="step">шаг 3 из 3</span>
          </div>
          <div className="cnew-card-body">
            <div className="cnew-row">
              <div className="key">Описание <span className="hint">markdown OK</span></div>
              <textarea
                className="cnew-input"
                value={description}
                onChange={e => setDescription(e.target.value)}
                placeholder="О чём сообщество, для кого, что одобряется и что нет"
                maxLength={500}
              />
            </div>
            <div className="cnew-row">
              <div className="key">Правила <span className="hint">видны на странице сообщества</span></div>
              <div className="cnew-rule-list">
                {rules.map((r, i) => (
                  <div className="cnew-rule" key={i}>
                    <span className="num">{i + 1}</span>
                    <input
                      value={r}
                      onChange={e => setRule(i, e.target.value)}
                      placeholder="Сформулируй правило одной фразой"
                      maxLength={140}
                    />
                    <button type="button" className="x" onClick={() => removeRule(i)} aria-label="Удалить">
                      <IconX size={13} stroke={2} />
                    </button>
                  </div>
                ))}
                <button type="button" className="cnew-add-rule" onClick={addRule}>
                  <IconPlus size={13} stroke={2} />
                  Добавить правило
                </button>
              </div>
            </div>
          </div>
        </div>

        {error && (
          <div style={{ fontFamily: 'var(--mono)', fontSize: 12, color: 'var(--hn-red)', padding: '10px 14px', background: 'rgba(191,97,106,.08)', borderRadius: 8 }}>
            {error}
          </div>
        )}

        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, paddingTop: 6 }}>
          <button className="btn ghost" onClick={() => navigate(-1)} disabled={submitting}>
            Отмена
          </button>
          <button className="btn primary" onClick={submit} disabled={!valid || submitting}>
            {submitting ? 'создаём…' : 'Создать сообщество'}
          </button>
        </div>
      </div>

      <aside className="cnew-side">
        <div className="cnew-preview-card">
          <div className="cnew-preview-head">
            превью
            <span className="live"><span className="pulse" />live</span>
          </div>
          <div className="cnew-preview-body">
            <div
              className="cnew-preview-banner"
              style={bannerUrl ? { backgroundImage: `url(${bannerUrl})` } : undefined}
            />
            <div className="cnew-preview-row">
              <div className="cnew-preview-icon" style={{ background: color }}>{previewIcon}</div>
              <div className="cnew-preview-info">
                <div className="name">{name || 'Название'}</div>
                <div className="slug">hg/{slug || 'slug'}</div>
                {description && <div className="tagline">{description}</div>}
              </div>
            </div>
            {rules.some(r => r.trim()) && (
              <div className="cnew-preview-rules">
                <div className="cnew-preview-rules-head">правила</div>
                <ol>
                  {rules.filter(r => r.trim()).map((r, i) => <li key={i}>{r}</li>)}
                </ol>
              </div>
            )}
          </div>
        </div>
        <div className="tip-card">
          <h4>почему важно</h4>
          <ul>
            <li><b>Slug</b> навсегда — сообщество живёт под этим именем.</li>
            <li><b>Цвет</b> — узнаваемая иконка в ленте.</li>
            <li><b>Правила</b> модераторы будут применять.</li>
          </ul>
        </div>
      </aside>

      {cropState && (
        <ImageCropModal
          image={cropState.src}
          aspect={7.2}
          minWidth={1200}
          hint="1200×160"
          onDone={handleCropDone}
          onCancel={() => setCropState(null)}
        />
      )}
    </div>
  );
};

export default CreateCommunityScreen;
