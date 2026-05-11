import { useMemo, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { IconChevronUp, IconPlus, IconPencil, IconTrash, IconCheck, IconX, IconDots } from '@tabler/icons-react';
import { qk } from '../lib/queryKeys';
import {
  listFeedback, listRoadmapItems,
  createRoadmapItem, updateRoadmapItem, deleteRoadmapItem,
  updateFeedbackStatus, deleteFeedback,
} from '../lib/api';
import { useVoteFeedback, useCreateFeedback } from '../lib/mutations';
import { useAuth } from '../AuthContext';
import { useMe } from '../App';
import { relativeTime } from '../lib/utils';

const PHASES = [
  { key: 'wip', title: 'Сейчас в', em: 'работе', when: 'in progress' },
  { key: 'next', title: 'Дальше —', em: 'очередь', when: 'next' },
  { key: 'later', title: 'На', em: 'потом', when: 'later · обдумываем' },
  { key: 'done', title: 'Что уже', em: 'собрали', when: 'done' },
];

const FEEDBACK_TYPES = [
  { k: 'idea',     l: 'Идея',  ico: '💡' },
  { k: 'bug',      l: 'Баг',   ico: '🐞' },
  { k: 'question', l: 'Вопрос', ico: '❓' },
  { k: 'other',    l: 'Прочее', ico: '💬' },
];

const TYPE_LABEL = { idea: 'идея', bug: 'баг', question: 'вопрос', other: 'прочее' };

const FEEDBACK_STATUSES = [
  { k: 'open',        l: 'открыто' },
  { k: 'planned',     l: 'запланировано' },
  { k: 'in_progress', l: 'в работе' },
  { k: 'done',        l: 'сделано' },
  { k: 'rejected',    l: 'отклонено' },
];

const EMPTY_ITEM = { phase: 'next', title: '', description: '', tags: [], eta: '', featured: false, sortOrder: 0 };

const ItemEditor = ({ initial, onSave, onCancel }) => {
  const [form, setForm] = useState(() => ({
    phase: initial?.phase ?? 'next',
    title: initial?.title ?? '',
    description: initial?.description ?? '',
    tags: initial?.tags?.join(', ') ?? '',
    eta: initial?.eta ?? '',
    featured: initial?.featured ?? false,
    sortOrder: initial?.sortOrder ?? 0,
  }));
  const set = (k, v) => setForm(f => ({ ...f, [k]: v }));

  const submit = () => {
    const tags = form.tags.split(',').map(t => t.trim()).filter(Boolean);
    onSave({
      phase: form.phase,
      title: form.title,
      description: form.description,
      tags,
      eta: form.eta || null,
      featured: form.featured,
      sortOrder: Number(form.sortOrder) || 0,
    });
  };

  return (
    <div className="rm-card editing" style={{ border: '1.5px dashed var(--honey)', padding: 14, display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div style={{ display: 'flex', gap: 8 }}>
        <select value={form.phase} onChange={e => set('phase', e.target.value)} style={{ flex: '0 0 auto', padding: '4px 8px', borderRadius: 6, border: '1px solid var(--border)' }}>
          {PHASES.map(p => <option key={p.key} value={p.key}>{p.key}</option>)}
        </select>
        <input placeholder="Заголовок" value={form.title} onChange={e => set('title', e.target.value)} style={{ flex: 1, padding: '4px 8px', borderRadius: 6, border: '1px solid var(--border)' }} />
        <label style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 12, whiteSpace: 'nowrap' }}>
          <input type="checkbox" checked={form.featured} onChange={e => set('featured', e.target.checked)} /> featured
        </label>
      </div>
      <textarea placeholder="Описание" value={form.description} onChange={e => set('description', e.target.value)} rows={3} style={{ padding: '4px 8px', borderRadius: 6, border: '1px solid var(--border)', resize: 'vertical' }} />
      <div style={{ display: 'flex', gap: 8 }}>
        <input placeholder="Теги через запятую" value={form.tags} onChange={e => set('tags', e.target.value)} style={{ flex: 1, padding: '4px 8px', borderRadius: 6, border: '1px solid var(--border)' }} />
        <input placeholder="ETA" value={form.eta} onChange={e => set('eta', e.target.value)} style={{ width: 120, padding: '4px 8px', borderRadius: 6, border: '1px solid var(--border)' }} />
        <input type="number" placeholder="sort" value={form.sortOrder} onChange={e => set('sortOrder', e.target.value)} style={{ width: 60, padding: '4px 8px', borderRadius: 6, border: '1px solid var(--border)' }} />
      </div>
      <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
        <button className="btn ghost" onClick={onCancel} style={{ gap: 4 }}><IconX size={14} /> отмена</button>
        <button className="btn primary" onClick={submit} disabled={!form.title.trim()} style={{ gap: 4 }}><IconCheck size={14} /> сохранить</button>
      </div>
    </div>
  );
};

const RmCard = ({ item, statusKey, isAdmin, onEdit, onDelete }) => (
  <div className={'rm-card' + (item.featured ? ' featured' : '')}>
    <div className="rm-card-head">
      <div className="rm-card-title">{item.title}</div>
      <span className={`rm-status ${statusKey}`}><span className="dot"></span>{statusKey === 'wip' ? 'in progress' : statusKey}</span>
      {isAdmin && (
        <span style={{ marginLeft: 'auto', display: 'flex', gap: 4 }}>
          <button className="icon-btn" onClick={() => onEdit(item)} title="Редактировать" style={{ width: 28, height: 28, minWidth: 28 }}><IconPencil size={14} /></button>
          <button className="icon-btn" onClick={() => onDelete(item)} title="Удалить" style={{ width: 28, height: 28, minWidth: 28, color: 'var(--danger, #d33)' }}><IconTrash size={14} /></button>
        </span>
      )}
    </div>
    <div className="rm-card-desc">{item.description}</div>
    <div className="rm-card-meta">
      {item.tags?.map(t => (
        <span key={t} className="tag"><span className="hash">#</span>{t}</span>
      ))}
      {item.eta && <span className="rm-eta">{item.eta}</span>}
    </div>
  </div>
);

const RoadmapScreen = () => {
  const { isAuthenticated, login } = useAuth();
  const { me } = useMe() ?? {};
  const isAdmin = me?.isAdmin ?? false;
  const qc = useQueryClient();

  const [type, setType] = useState('idea');
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [anon, setAnon] = useState(false);
  const [okMsg, setOkMsg] = useState(null);
  const [errMsg, setErrMsg] = useState(null);
  const [boardSort, setBoardSort] = useState('top');

  const [editing, setEditing] = useState(null);
  const [adding, setAdding] = useState(false);

  const createMut = useCreateFeedback();
  const voteMut = useVoteFeedback();

  const { data: roadmapItems } = useQuery({
    queryKey: qk.roadmapItems(),
    queryFn: listRoadmapItems,
  });

  const { data: ideas } = useQuery({
    queryKey: qk.feedback(boardSort, ''),
    queryFn: () => listFeedback(boardSort, '', 50, 0),
  });

  const createItemMut = useMutation({
    mutationFn: (data) => createRoadmapItem(data),
    onSettled: () => qc.invalidateQueries({ queryKey: qk.roadmapItems() }),
  });
  const updateItemMut = useMutation({
    mutationFn: ({ id, data }) => updateRoadmapItem(id, data),
    onSettled: () => qc.invalidateQueries({ queryKey: qk.roadmapItems() }),
  });
  const deleteItemMut = useMutation({
    mutationFn: (id) => deleteRoadmapItem(id),
    onSettled: () => qc.invalidateQueries({ queryKey: qk.roadmapItems() }),
  });
  const updateFbStatusMut = useMutation({
    mutationFn: ({ id, status }) => updateFeedbackStatus(id, status),
    onSettled: () => qc.invalidateQueries({ queryKey: qk.feedback(boardSort, '') }),
  });
  const deleteFbMut = useMutation({
    mutationFn: (id) => deleteFeedback(id),
    onSettled: () => qc.invalidateQueries({ queryKey: qk.feedback(boardSort, '') }),
  });

  const items = ideas ?? [];
  const allItems = roadmapItems ?? [];

  const grouped = useMemo(() => {
    const map = {};
    for (const p of PHASES) map[p.key] = [];
    for (const it of allItems) {
      if (map[it.phase]) map[it.phase].push(it);
      else map[it.phase] = [it];
    }
    return map;
  }, [allItems]);

  const counts = useMemo(() => {
    const total = allItems.length;
    const wip = grouped.wip?.length ?? 0;
    const next = grouped.next?.length ?? 0;
    const later = grouped.later?.length ?? 0;
    const done = grouped.done?.length ?? 0;
    return { total, wip, next, later, done };
  }, [allItems, grouped]);

  const valid = title.trim().length >= 8 && body.trim().length >= 20 && body.trim().length <= 4000;

  const submit = async () => {
    if (!isAuthenticated) { login(); return; }
    if (!valid || createMut.isPending) return;
    setOkMsg(null);
    setErrMsg(null);
    try {
      await createMut.mutateAsync({ type, title: title.trim(), body: body.trim(), isAnon: anon });
      setOkMsg('спасибо! идея добавлена в список');
      setTitle('');
      setBody('');
      setTimeout(() => setOkMsg(null), 4000);
    } catch (e) {
      setErrMsg(e.message ?? 'не удалось отправить');
    }
  };

  const toggleVote = (idea) => {
    if (!isAuthenticated) { login(); return; }
    voteMut.mutate({ id: idea.id, next: !idea.isVoted });
  };

  const handleSaveNew = async (data) => {
    await createItemMut.mutateAsync(data);
    setAdding(false);
  };

  const handleSaveEdit = async (data) => {
    await updateItemMut.mutateAsync({ id: editing.id, data });
    setEditing(null);
  };

  const handleDelete = (item) => {
    if (!confirm(`Удалить «${item.title}»?`)) return;
    deleteItemMut.mutate(item.id);
  };

  return (
    <div className="roadmap-shell">

      <section className="roadmap-hero">
        <div>
          <h1>Что мы <span className="amp">сажаем</span><br/>в саду</h1>
          <p>Открытый план развития HoneyGarden. Без обещаний по датам — только приоритеты и направления. Чем выше в списке, тем ближе к работе.</p>
        </div>
        <div className="roadmap-stats">
          <div className="row"><span className="k">сделано</span><span className="v">{counts.done} шагов</span></div>
          <div className="row"><span className="k">в работе</span><span className="v">{counts.wip}</span></div>
          <div className="row"><span className="k">в очереди</span><span className="v">{counts.next}</span></div>
          <div className="row"><span className="k">on hold</span><span className="v">{counts.later}</span></div>
        </div>
      </section>

      <div className="roadmap-legend">
        <span className="pill"><span className="dot done"></span> done — задеплоено</span>
        <span className="pill"><span className="dot wip"></span> in progress — пилим сейчас</span>
        <span className="pill"><span className="dot next"></span> next — в ближайшем спринте</span>
        <span className="pill"><span className="dot later"></span> later — обдумываем</span>
      </div>

      {isAdmin && (
        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 12 }}>
          <button className="btn primary" onClick={() => setAdding(true)} style={{ gap: 4 }}>
            <IconPlus size={15} /> Добавить элемент
          </button>
        </div>
      )}

      {adding && (
        <ItemEditor initial={EMPTY_ITEM} onSave={handleSaveNew} onCancel={() => setAdding(false)} />
      )}

      {PHASES.map(phase => {
        const phaseItems = grouped[phase.key] ?? [];
        if (phaseItems.length === 0 && !isAdmin) return null;
        return (
          <section key={phase.key} className="rm-phase">
            <div className="rm-phase-head">
              <h2>{phase.title} <span className="em">{phase.em}</span></h2>
              <span className="when">{phase.when}</span>
              <span className="count">{phaseItems.length} {phaseItems.length === 1 ? 'фича' : 'фич'}</span>
            </div>
            <div className="rm-cards">
              {phaseItems.map((it) =>
                editing?.id === it.id
                  ? <ItemEditor key={it.id} initial={it} onSave={handleSaveEdit} onCancel={() => setEditing(null)} />
                  : <RmCard key={it.id} item={it} statusKey={phase.key} isAdmin={isAdmin} onEdit={setEditing} onDelete={handleDelete} />
              )}
            </div>
          </section>
        );
      })}

      <section className="rm-feedback">
        <div className="rm-fb-head">
          <h2>Поделись <span className="em">мыслью</span></h2>
          <p>Нашёл баг? Не хватает фичи? Пиши — идеи попадают в публичный список справа, мы голосуем и тащим их в работу по приоритету.</p>
        </div>

        <div className="fb-form">
          <div className="fb-row">
            <span className="lbl">тип обращения</span>
            <div className="fb-types">
              {FEEDBACK_TYPES.map(t => (
                <button
                  key={t.k}
                  type="button"
                  className={'fb-type' + (type === t.k ? ' active' : '')}
                  onClick={() => setType(t.k)}
                >
                  <span>{t.ico}</span> {t.l}
                </button>
              ))}
            </div>
          </div>
          <div className="fb-row">
            <span className="lbl">заголовок</span>
            <input
              className="fb-input"
              placeholder="кратко в одном предложении…"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              maxLength={200}
            />
          </div>
          <div className="fb-row">
            <span className="lbl">подробности</span>
            <textarea
              className="fb-textarea"
              placeholder="что предлагаешь / как воспроизвести / почему важно…"
              value={body}
              onChange={(e) => setBody(e.target.value)}
              maxLength={4000}
            />
          </div>
          {okMsg && <div className="fb-success">{okMsg}</div>}
          {errMsg && <div className="fb-error">{errMsg}</div>}
          <div className="fb-foot">
            <span className="fb-counter">{body.length} / 4000</span>
            <label className="fb-anon">
              <input type="checkbox" checked={anon} onChange={(e) => setAnon(e.target.checked)} />
              анонимно
            </label>
            <button className="fb-submit" onClick={submit} disabled={!valid || createMut.isPending}>
              {createMut.isPending ? 'отправляем…' : 'отправить'}
            </button>
          </div>
        </div>

        <aside className="fb-board">
          <div className="fb-board-head">
            <span className="ttl">Идеи сообщества</span>
            <span className="sub">{items.length} предложений</span>
          </div>
          <div className="fb-board-tabs">
            <button
              className={'fb-board-tab' + (boardSort === 'top' ? ' active' : '')}
              onClick={() => setBoardSort('top')}
            >Топ</button>
            <button
              className={'fb-board-tab' + (boardSort === 'new' ? ' active' : '')}
              onClick={() => setBoardSort('new')}
            >Свежие</button>
          </div>

          {items.length === 0 && (
            <div style={{ color: 'var(--text-dim)', fontSize: 13, padding: '14px 4px' }}>
              пока пусто — будь первым
            </div>
          )}

          {items.map(idea => (
            <div key={idea.id} className="fb-idea">
              <button
                type="button"
                className={'fb-vote' + (idea.isVoted ? ' voted' : '')}
                onClick={() => toggleVote(idea)}
              >
                <IconChevronUp size={13} stroke={2} />
                <span className="num">{idea.votesCount}</span>
              </button>
              <div className="fb-idea-body">
                <div className="fb-idea-title">{idea.title}</div>
                <div className="fb-idea-meta">
                  <span className={'pill-mini ' + idea.type}>{TYPE_LABEL[idea.type] ?? idea.type}</span>
                  {idea.status && idea.status !== 'open' && (
                    <span className={'pill-mini status-' + idea.status}>{FEEDBACK_STATUSES.find(s => s.k === idea.status)?.l ?? idea.status}</span>
                  )}
                  <span>{idea.isAnon ? 'анонимно' : (idea.author?.username ? `@${idea.author.username}` : 'кто-то')} · {relativeTime(idea.createdAt)}</span>
                </div>
                {isAdmin && (
                  <div style={{ display: 'flex', gap: 6, marginTop: 4, alignItems: 'center' }}>
                    <select
                      value={idea.status || 'open'}
                      onChange={(e) => updateFbStatusMut.mutate({ id: idea.id, status: e.target.value })}
                      style={{ fontSize: 11, padding: '2px 6px', borderRadius: 4, border: '1px solid var(--border)' }}
                    >
                      {FEEDBACK_STATUSES.map(s => <option key={s.k} value={s.k}>{s.l}</option>)}
                    </select>
                    <button
                      className="icon-btn"
                      onClick={() => { if (confirm(`Удалить «${idea.title}»?`)) deleteFbMut.mutate(idea.id); }}
                      title="Удалить"
                      style={{ width: 24, height: 24, minWidth: 24, color: 'var(--danger, #d33)' }}
                    >
                      <IconTrash size={13} />
                    </button>
                  </div>
                )}
              </div>
            </div>
          ))}
        </aside>
      </section>

    </div>
  );
};

export default RoadmapScreen;
