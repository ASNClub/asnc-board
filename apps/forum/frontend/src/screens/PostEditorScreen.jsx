import { useState, useRef, useEffect, useMemo } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  IconBold, IconItalic, IconStrikethrough,
  IconQuote, IconList, IconListNumbers,
  IconCode, IconLink, IconPhoto, IconTable,
  IconMessageCircle, IconFileText, IconHelpCircle, IconLink as IconLinkType,
  IconEye, IconLoader2,
} from '@tabler/icons-react';
import { createPost, updatePost, getPost, uploadFile, getMyCommunity } from '../lib/api';
import { commColor } from '../lib/utils';
import { renderMarkdown } from '../lib/markdown';

const TYPE_TABS = [
  { id: 'discussion', label: 'Обсуждение', Icon: IconMessageCircle },
  { id: 'article',    label: 'Статья',     Icon: IconFileText },
  { id: 'question',   label: 'Вопрос',     Icon: IconHelpCircle },
];

const PostEditorScreen = ({ slug, postId }) => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const isEdit = !!postId;
  const onBack = () => isEdit ? navigate(`/p/${postId}`) : slug ? navigate(`/c/${slug}`) : navigate('/');

  const initialType = (() => {
    const t = searchParams.get('type');
    return ['discussion', 'article', 'question'].includes(t) ? t : 'discussion';
  })();

  const [title, setTitle] = useState('');
  const [body,  setBody]  = useState('');
  const [type,  setType]  = useState(initialType);
  const [tags,  setTags]  = useState([]);
  const [tagDraft, setTagDraft] = useState('');
  const [resolvedSlug, setResolvedSlug] = useState(slug ?? '');
  const [resolvedName, setResolvedName] = useState('');
  const [busy, setBusy] = useState(false);
  const [err,  setErr]  = useState(null);
  const [uploading, setUploading] = useState(false);
  const [showPreview, setShowPreview] = useState(false);

  const textareaRef = useRef(null);
  const fileRef = useRef(null);

  useEffect(() => {
    if (!isEdit) return;
    let cancelled = false;
    getPost(postId).then(p => {
      if (cancelled || !p) return;
      const raw = p.content ?? '';
      const tagMatch = raw.match(/\n\n((?:#[\wа-яА-ЯёЁ_-]+\s*)+)\s*$/i);
      if (tagMatch) {
        const found = tagMatch[1].split(/\s+/).filter(Boolean).map(t => t.replace(/^#/, '').toLowerCase());
        setTags(found);
        setBody(raw.slice(0, raw.length - tagMatch[0].length));
      } else {
        setBody(raw);
      }
      setTitle(p.title ?? '');
      if (p.kind) setType(p.kind);
      if (p.communitySlug) {
        setResolvedSlug(p.communitySlug);
        setResolvedName(p.communitySlug);
      }
    }).catch(e => setErr(e.message ?? 'не удалось загрузить пост'));
    return () => { cancelled = true; };
  }, [postId, isEdit]);

  useEffect(() => {
    if (slug || isEdit) return;
    let cancelled = false;
    getMyCommunity().then(mine => {
      if (cancelled) return;
      const mineSlug = mine?.Slug ?? mine?.slug;
      const mineName = mine?.Name ?? mine?.name;
      if (mineSlug) {
        setResolvedSlug(mineSlug);
        setResolvedName(mineName ?? mineSlug);
      } else {
        navigate('/c/new', { replace: true });
      }
    }).catch(() => {
      if (!cancelled) navigate('/c/new', { replace: true });
    });
    return () => { cancelled = true; };
  }, [slug, isEdit, navigate]);

  const targetSlug = slug || resolvedSlug;

  const submit = async () => {
    setErr(null);
    const t = title.trim();
    const c = body.trim();
    if (!isEdit && !targetSlug) { setErr('выбери сообщество для публикации'); return; }
    if (!t && !c)    { setErr('заголовок или текст должны быть не пустыми'); return; }
    setBusy(true);
    try {
      const tagBlock = tags.length > 0 ? `\n\n${tags.map(x => `#${x}`).join(' ')}` : '';
      const finalTitle = t || c.slice(0, 80);
      const finalContent = c + tagBlock;
      if (isEdit) {
        await updatePost(postId, { title: finalTitle, content: finalContent, kind: type });
        navigate(`/p/${postId}`);
      } else {
        const post = await createPost(targetSlug, { title: finalTitle, content: finalContent, kind: type });
        navigate(`/p/${post.id}`);
      }
    } catch (e) {
      setErr(e.message || (isEdit ? 'не удалось сохранить' : 'не удалось опубликовать'));
    } finally {
      setBusy(false);
    }
  };

  useEffect(() => {
    const onKey = (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') { e.preventDefault(); submit(); }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  });

  const insertAround = (left, right = '', placeholder = '') => {
    const ta = textareaRef.current;
    if (!ta) return;
    const { selectionStart: s, selectionEnd: e, value } = ta;
    const sel = value.slice(s, e) || placeholder;
    const next = value.slice(0, s) + left + sel + right + value.slice(e);
    setBody(next);
    requestAnimationFrame(() => {
      ta.focus();
      const cursor = s + left.length + sel.length;
      ta.setSelectionRange(cursor, cursor);
    });
  };

  const insertLine = (prefix) => {
    const ta = textareaRef.current;
    if (!ta) return;
    const { selectionStart: s, value } = ta;
    const lineStart = value.lastIndexOf('\n', s - 1) + 1;
    const next = value.slice(0, lineStart) + prefix + value.slice(lineStart);
    setBody(next);
    requestAnimationFrame(() => {
      ta.focus();
      const pos = s + prefix.length;
      ta.setSelectionRange(pos, pos);
    });
  };

  const onPickFile = async (e) => {
    const f = e.target.files?.[0];
    e.target.value = '';
    if (!f) return;
    setUploading(true);
    try {
      const url = await uploadFile(f);
      const md = f.type.startsWith('image/') ? `![${f.name}](${url})` : `[${f.name}](${url})`;
      insertAround(md, '', '');
    } catch (er) {
      setErr(er.message || 'не удалось загрузить файл');
    } finally {
      setUploading(false);
    }
  };

  const addTag = (raw) => {
    const t = raw.trim().replace(/^#/, '').toLowerCase();
    if (!t) return;
    if (tags.includes(t)) { setTagDraft(''); return; }
    if (tags.length >= 8) return;
    setTags([...tags, t]);
    setTagDraft('');
  };

  const removeTag = (t) => setTags(tags.filter(x => x !== t));

  const onTagKey = (e) => {
    if (e.key === 'Enter' || e.key === ',' || e.key === ' ') {
      e.preventDefault();
      addTag(tagDraft);
    } else if (e.key === 'Backspace' && !tagDraft && tags.length > 0) {
      setTags(tags.slice(0, -1));
    }
  };

  const pickedComm = useMemo(() => {
    if (slug) return { slug, name: slug };
    if (resolvedSlug) return { slug: resolvedSlug, name: resolvedName || resolvedSlug };
    return null;
  }, [slug, resolvedSlug, resolvedName]);

  const charCount = body.length;
  const previewHtml = useMemo(() => renderMarkdown(body), [body]);
  const excerpt = body.replace(/[#*`>_\[\]()!]/g, '').replace(/\n+/g, ' ').slice(0, 180);

  return (
    <div className="editor-shell">
      <main className="editor-main">

        <div className="editor-head">
          <div>
            <h1>{isEdit ? <>Правим <span className="amp">пост</span></> : <>Что посадим в <span className="amp">сад</span>?</>}</h1>
          </div>
          <div className="editor-actions">
            <button className="btn" onClick={onBack} disabled={busy}>отмена</button>
            <button className="btn primary" onClick={submit} disabled={busy}>
              {busy ? (isEdit ? 'сохраняем…' : 'публикуем…') : (isEdit ? 'сохранить' : 'опубликовать')}
            </button>
          </div>
        </div>

        <div className="type-tabs">
          {TYPE_TABS.map(({ id, label, Icon }) => (
            <button
              key={id}
              type="button"
              className={'type-tab' + (type === id ? ' active' : '')}
              onClick={() => setType(id)}
            >
              <Icon size={14} stroke={2} /> {label}
            </button>
          ))}
        </div>

        <div className="editor-card">
          <input
            className="title-input"
            placeholder="Заголовок — короткий и по делу"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            maxLength={200}
          />

          <div className="editor-toolbar">
            <button type="button" className="tb-btn" title="Жирный (⌘B)" onClick={() => insertAround('**', '**', 'жирный')}>
              <IconBold size={14} stroke={2.2} />
            </button>
            <button type="button" className="tb-btn" title="Курсив" onClick={() => insertAround('*', '*', 'курсив')}>
              <IconItalic size={14} stroke={2} />
            </button>
            <button type="button" className="tb-btn" title="Зачёркнутый" onClick={() => insertAround('~~', '~~', 'текст')}>
              <IconStrikethrough size={14} stroke={2} />
            </button>
            <span className="tb-divider"></span>
            <button type="button" className="tb-btn" title="Заголовок 2" onClick={() => insertLine('## ')}>
              <span style={{ fontFamily: 'var(--serif)', fontWeight: 700, fontSize: 12 }}>H2</span>
            </button>
            <button type="button" className="tb-btn" title="Заголовок 3" onClick={() => insertLine('### ')}>
              <span style={{ fontFamily: 'var(--serif)', fontWeight: 700, fontSize: 11 }}>H3</span>
            </button>
            <span className="tb-divider"></span>
            <button type="button" className="tb-btn" title="Цитата" onClick={() => insertLine('> ')}>
              <IconQuote size={14} stroke={2} />
            </button>
            <button type="button" className="tb-btn" title="Список" onClick={() => insertLine('- ')}>
              <IconList size={14} stroke={2} />
            </button>
            <button type="button" className="tb-btn" title="Нумерованный список" onClick={() => insertLine('1. ')}>
              <IconListNumbers size={14} stroke={2} />
            </button>
            <span className="tb-divider"></span>
            <button type="button" className="tb-btn" title="Код" onClick={() => insertAround('`', '`', 'код')}>
              <IconCode size={14} stroke={2} />
            </button>
            <button type="button" className="tb-btn" title="Ссылка (⌘K)" onClick={() => insertAround('[', '](url)', 'текст')}>
              <IconLink size={14} stroke={2} />
            </button>
            <button
              type="button" className="tb-btn"
              title={uploading ? 'загружаем…' : 'Изображение / файл'}
              onClick={() => fileRef.current?.click()}
              disabled={uploading}
            >
              {uploading ? <IconLoader2 size={14} stroke={2} className="spin" /> : <IconPhoto size={14} stroke={2} />}
            </button>
            <input ref={fileRef} type="file" hidden onChange={onPickFile} />
            <button type="button" className="tb-btn" title="Таблица" onClick={() => insertAround('\n| col | col |\n| --- | --- |\n| ', ' |  |\n', 'cell')}>
              <IconTable size={14} stroke={2} />
            </button>
            <span className="tb-spacer"></span>
            <span className="tb-meta">markdown</span>
          </div>

          {showPreview ? (
            <div
              className="editor-body"
              dangerouslySetInnerHTML={{ __html: previewHtml || '<p class="placeholder">пусто — нечего показать</p>' }}
            />
          ) : (
            <textarea
              ref={textareaRef}
              className="editor-body"
              placeholder="Начни писать… поддерживается markdown."
              value={body}
              onChange={(e) => setBody(e.target.value)}
              spellCheck={false}
              style={{ width: '100%', border: 'none', resize: 'vertical', background: 'transparent' }}
            />
          )}

          <div className="editor-row" style={{ borderTop: '1px solid var(--border)', padding: '8px 16px' }}>
            <div className="lbl">теги</div>
            <div className="val tag-pill-input" style={{ padding: 0 }}>
              {tags.map(t => (
                <span key={t} className="tag-pill">
                  #{t}
                  <span className="x" onClick={() => removeTag(t)}>×</span>
                </span>
              ))}
              <input
                className="tag-input"
                placeholder={tags.length === 0 ? 'добавить тег…' : ''}
                value={tagDraft}
                onChange={(e) => setTagDraft(e.target.value)}
                onKeyDown={onTagKey}
                onBlur={() => addTag(tagDraft)}
              />
            </div>
          </div>

          <div className="editor-foot">
            <div className="foot-hint">
              <kbd>⌘</kbd><kbd>↵</kbd> опубликовать
              <span style={{ color: 'var(--border-strong)' }}>·</span>
              <span>{charCount} / 40 000 символов</span>
            </div>
            <div className="editor-foot-actions">
              <button
                className="btn ghost"
                style={{ fontSize: 12 }}
                onClick={() => setShowPreview(p => !p)}
                type="button"
              >
                <IconEye size={13} stroke={2} />
                {showPreview ? 'редактировать' : 'предпросмотр'}
              </button>
            </div>
          </div>
        </div>

        {err && (
          <div className="error-banner" style={{ marginTop: 4 }}>{err}</div>
        )}
      </main>

      <aside className="editor-side">
        <div className="preview-card">
          <div className="preview-card-head">
            предпросмотр
            <span className="live"><span className="pulse"></span>live</span>
          </div>
          <div className="preview-card-body">
            {pickedComm && (
              <div className="preview-mini-pill">
                <span className="ico" style={{ background: commColor(pickedComm.slug) }}>
                  {pickedComm.slug[0]?.toUpperCase()}
                </span>
                hg/{pickedComm.slug}
              </div>
            )}
            <h3 className="preview-mini-title">
              {title.trim() || 'Заголовок появится здесь'}
            </h3>
            <p className="preview-mini-excerpt">
              {excerpt || 'Текст поста появится здесь по мере набора.'}
            </p>
          </div>
        </div>

        <div className="tip-card">
          <h4>несколько правил сада</h4>
          <ul>
            <li>пиши <b>заголовок</b> как обещание — что человек получит, прочитав</li>
            <li>держи <b>теги</b> в нижнем регистре, латиницей</li>
            <li>код вставляй через <b>```</b> с указанием языка</li>
            <li>длинные ссылки можно через <b>⌘K</b></li>
          </ul>
        </div>
      </aside>
    </div>
  );
};

export default PostEditorScreen;
