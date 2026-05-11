import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { IconPlus, IconTrash, IconShieldLock } from '@tabler/icons-react';
import { qk } from '../lib/queryKeys';
import { listBannedWords, createBannedWord, deleteBannedWord } from '../lib/api';
import { relativeTime } from '../lib/utils';

const SCOPES = [
  { k: 'both', l: 'Юзернейм + паблик' },
  { k: 'username', l: 'Только юзернейм' },
  { k: 'slug', l: 'Только паблик' },
];

const AdminScreen = () => {
  const qc = useQueryClient();
  const [word, setWord] = useState('');
  const [scope, setScope] = useState('both');

  const { data: words = [] } = useQuery({
    queryKey: qk.bannedWords(),
    queryFn: listBannedWords,
  });

  const addMut = useMutation({
    mutationFn: (data) => createBannedWord(data),
    onSettled: () => qc.invalidateQueries({ queryKey: qk.bannedWords() }),
  });

  const delMut = useMutation({
    mutationFn: (id) => deleteBannedWord(id),
    onSettled: () => qc.invalidateQueries({ queryKey: qk.bannedWords() }),
  });

  const handleAdd = async () => {
    const w = word.trim().toLowerCase();
    if (!w) return;
    await addMut.mutateAsync({ word: w, scope });
    setWord('');
  };

  const handleDelete = (bw) => {
    if (!confirm(`Убрать «${bw.word}» из бан-листа?`)) return;
    delMut.mutate(bw.id);
  };

  return (
    <div className="settings-shell" style={{ maxWidth: 900, margin: '0 auto', padding: '32px 20px' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 24 }}>
        <IconShieldLock size={24} stroke={1.6} />
        <h1 style={{ fontSize: 22, fontWeight: 700 }}>Админ-панель</h1>
      </div>

      <div className="settings-card">
        <div className="settings-card-head">
          <h2>Бан-лист слов</h2>
          <span className="sub">Запрещённые слова в юзернеймах и слагах пабликов. Проверка по вхождению подстроки.</span>
        </div>

        <div style={{ display: 'flex', gap: 8, padding: '12px 16px', borderBottom: '1px solid var(--border)' }}>
          <input
            placeholder="слово…"
            value={word}
            onChange={e => setWord(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && handleAdd()}
            style={{ flex: 1, padding: '6px 10px', borderRadius: 6, border: '1px solid var(--border)' }}
          />
          <select
            value={scope}
            onChange={e => setScope(e.target.value)}
            style={{ padding: '6px 10px', borderRadius: 6, border: '1px solid var(--border)' }}
          >
            {SCOPES.map(s => <option key={s.k} value={s.k}>{s.l}</option>)}
          </select>
          <button className="btn primary" onClick={handleAdd} disabled={!word.trim() || addMut.isPending} style={{ gap: 4 }}>
            <IconPlus size={14} /> Добавить
          </button>
        </div>

        <div style={{ maxHeight: 500, overflowY: 'auto' }}>
          {words.length === 0 && (
            <div style={{ padding: '20px 16px', color: 'var(--text-dim)', fontSize: 13 }}>
              Бан-лист пуст
            </div>
          )}
          {words.map(bw => (
            <div key={bw.id} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '8px 16px', borderBottom: '1px solid var(--border-light, var(--border))' }}>
              <code style={{ flex: 1, fontSize: 13, fontFamily: 'var(--mono)' }}>{bw.word}</code>
              <span style={{ fontSize: 11, color: 'var(--text-dim)', minWidth: 100 }}>
                {bw.scope === 'both' ? 'юзер + паблик' : bw.scope === 'username' ? 'юзернейм' : 'паблик'}
              </span>
              <span style={{ fontSize: 11, color: 'var(--text-dim)', minWidth: 80 }}>{relativeTime(bw.createdAt)}</span>
              <button className="icon-btn" onClick={() => handleDelete(bw)} title="Удалить" style={{ color: 'var(--danger, #d33)' }}>
                <IconTrash size={14} />
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default AdminScreen;
