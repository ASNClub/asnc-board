import { useEffect, useRef, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { IconSend } from '@tabler/icons-react';
import { getChatMessages, sendChatMessage, openChatSocket } from '../lib/api';
import { useAuth } from '../AuthContext';
import { initials, relativeTime } from '../lib/utils';

const Avatar = ({ user }) => (
  <div className="av">
    {user?.avatarUrl
      ? <img src={user.avatarUrl} alt="" />
      : initials(user?.username ?? '?')}
  </div>
);

const ChatDock = ({ meId = null }) => {
  const { isAuthenticated, login } = useAuth();
  const navigate = useNavigate();

  const [messages, setMessages] = useState([]);
  const [draft, setDraft] = useState('');
  const [sending, setSending] = useState(false);

  const scrollRef = useRef(null);
  const wsRef = useRef(null);

  const scrollToBottom = useCallback(() => {
    requestAnimationFrame(() => {
      const el = scrollRef.current;
      if (el) el.scrollTop = el.scrollHeight;
    });
  }, []);

  useEffect(() => {
    let cancelled = false;
    getChatMessages(50).then(list => {
      if (cancelled) return;
      setMessages(list ?? []);
      scrollToBottom();
    }).catch(() => {});
    return () => { cancelled = true; };
  }, [scrollToBottom]);

  useEffect(() => {
    if (!isAuthenticated) return;
    let stopped = false;
    let retry = 0;

    const connect = () => {
      if (stopped) return;
      const ws = openChatSocket();
      wsRef.current = ws;
      ws.onopen = () => { retry = 0; };
      ws.onmessage = (ev) => {
        try {
          const msg = JSON.parse(ev.data);
          setMessages(prev => {
            if (prev.some(m => m.id === msg.id)) return prev;
            return [...prev, msg];
          });
          scrollToBottom();
        } catch {}
      };
      ws.onclose = () => {
        if (stopped) return;
        retry = Math.min(retry + 1, 6);
        setTimeout(connect, 1000 * retry);
      };
      ws.onerror = () => { try { ws.close(); } catch {} };
    };

    connect();
    return () => {
      stopped = true;
      try { wsRef.current?.close(); } catch {}
    };
  }, [isAuthenticated, meId, scrollToBottom]);

  const send = async () => {
    const c = draft.trim();
    if (!c || sending) return;
    setSending(true);
    try {
      await sendChatMessage(c);
      setDraft('');
    } catch (e) {
      // silent fail — user sees message unsent (draft stays)
    } finally {
      setSending(false);
    }
  };

  const onKey = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      send();
    }
  };

  return (
    <div className="rail-chat">
      <div className="rail-title">Общий чат</div>

      <div className="chat-msgs" ref={scrollRef}>
        {messages.length === 0 ? (
          <div style={{ textAlign: 'center', color: 'var(--text-dim)', fontSize: 12, padding: '14px 8px' }}>
            пока тихо. напиши первым.
          </div>
        ) : messages.map(m => (
          <div key={m.id} className="chat-msg">
            <Avatar user={m.author} />
            <div className="chat-msg-body">
              <div className="chat-msg-meta">
                <span
                  className="author"
                  onClick={() => m.author?.username && navigate(`/u/${m.author.username}`)}
                >
                  {m.author?.displayName || m.author?.username || '?'}
                </span>
                <span className="time">{relativeTime(m.createdAt)}</span>
              </div>
              <div className="chat-msg-text">{m.content}</div>
            </div>
          </div>
        ))}
      </div>

      {isAuthenticated ? (
        <div className="chat-composer">
          <textarea
            className="chat-composer-input"
            placeholder="написать сообщение…"
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={onKey}
            maxLength={1000}
            rows={1}
          />
          <div className="chat-composer-row">
            <span className="hint">{draft.length}/1000</span>
            <button type="button" className="send" disabled={!draft.trim() || sending} onClick={send}>
              <IconSend size={13} stroke={2} />
              отправить
            </button>
          </div>
        </div>
      ) : (
        <div className="chat-composer">
          <button type="button" className="btn primary" style={{ justifyContent: 'center' }} onClick={login}>
            войди, чтобы писать
          </button>
        </div>
      )}
    </div>
  );
};

export default ChatDock;
