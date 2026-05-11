import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  IconChevronUp, IconChevronDown, IconMessage, IconBookmark, IconBookmarkFilled,
  IconUpload, IconExternalLink, IconFileText, IconHelpCircle, IconMessageCircle,
  IconDots, IconBan, IconTrash,
} from '@tabler/icons-react';
import { useQueryClient } from '@tanstack/react-query';
import { useVotePost, useBookmarkPost, useBlockUser } from '../lib/mutations';
import { qk } from '../lib/queryKeys';
import { excerpt } from '../lib/utils';
import { linkTarget } from '../lib/prefs';
import { adminDeletePost } from '../lib/api';

const KIND_META = {
  article:  { label: 'Статья',     Icon: IconFileText },
  question: { label: 'Вопрос',     Icon: IconHelpCircle },
  discussion: { label: 'Обсуждение', Icon: IconMessageCircle },
};

const PostCard = ({ post, density = 'cards', onOpen }) => {
  const navigate = useNavigate();
  const qc = useQueryClient();
  const me = qc.getQueryData(qk.me());
  const voteMut = useVotePost();
  const bookmarkMut = useBookmarkPost();
  const blockMut = useBlockUser();
  const [votedLocal, setVotedLocal] = useState(null);
  const [votesLocal, setVotesLocal] = useState(null);
  const voted = votedLocal !== null ? votedLocal === 'up' : post.voted === 'up';
  const votes = votesLocal !== null ? votesLocal : post.votes;
  const [savedLocal, setSavedLocal] = useState(null);
  const saved = savedLocal !== null ? savedLocal : !!post.bookmarked;
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef(null);

  useEffect(() => {
    if (!menuOpen) return;
    const onDoc = (e) => {
      if (menuRef.current && !menuRef.current.contains(e.target)) setMenuOpen(false);
    };
    document.addEventListener('mousedown', onDoc);
    return () => document.removeEventListener('mousedown', onDoc);
  }, [menuOpen]);

  const stop = (e) => e.stopPropagation();

  const handleBlock = (e) => {
    stop(e);
    setMenuOpen(false);
    if (!post.author) return;
    if (!confirm(`Заблокировать @${post.author}? Его посты исчезнут из ленты.`)) return;
    blockMut.mutate({ username: post.author, next: true });
  };

  const handleAdminDelete = async (e) => {
    stop(e);
    setMenuOpen(false);
    if (!confirm('Удалить пост? Это действие необратимо.')) return;
    try {
      await adminDeletePost(post.id);
      qc.invalidateQueries({ queryKey: ['feed'] });
      qc.invalidateQueries({ queryKey: ['trending'] });
    } catch (err) {
      alert(err.message ?? 'не удалось');
    }
  };

  const toggleVote = (e) => {
    stop(e);
    const next = !voted;
    setVotedLocal(next ? 'up' : null);
    setVotesLocal((votes ?? 0) + (next ? 1 : -1));
    voteMut.mutate({ postId: post.id, next }, {
      onError: () => { setVotedLocal(null); setVotesLocal(null); },
    });
  };

  const toggleSave = (e) => {
    stop(e);
    const next = !saved;
    setSavedLocal(next);
    bookmarkMut.mutate({ postId: post.id, next }, {
      onError: () => setSavedLocal(null),
    });
  };

  const handleShare = async (e) => {
    stop(e);
    try { await navigator.clipboard.writeText(`${window.location.origin}/p/${post.shortId || post.id}`); } catch {}
  };

  const compact = density === 'compact';
  const showThumb = !compact && !!post.thumb;

  return (
    <article
      className={'post' + (post.isRSS ? ' rss' : '') + (showThumb ? '' : ' no-thumb')}
      onClick={() => onOpen?.(post.shortId || post.id)}
    >
      <div>
        <div className="post-head">
          {post.isRSS ? (
            <span className="source-pill rss">
              <span className="ico">
                {post.sourceFavicon ? <img src={post.sourceFavicon} alt="" /> : '▸'}
              </span>
              {post.sourceName ?? 'RSS'}
            </span>
          ) : post.community ? (
            <Link
              to={`/c/${post.community}`}
              className="source-pill"
              onClick={stop}
              style={{ cursor: 'pointer', textDecoration: 'none', color: 'inherit' }}
            >
              <span className="ico" style={{ background: post.commColor }}>
                {post.community[0]?.toUpperCase()}
              </span>
              hg/{post.community}
            </Link>
          ) : null}

          {post.author && (
            <>
              <span className="dot">·</span>
              <Link
                to={`/u/${post.author}`}
                className="author"
                onClick={stop}
                style={{ textDecoration: 'none', color: 'inherit' }}
              >@{post.author}</Link>
            </>
          )}
          <span className="dot">·</span>
          <span>{post.time}</span>

          {post.kind && KIND_META[post.kind] && !post.isRSS && (() => {
            const { label, Icon } = KIND_META[post.kind];
            return (
              <>
                <span className="dot">·</span>
                <span className="post-kind"><Icon size={11} stroke={1.8} />{label}</span>
              </>
            );
          })()}

          {post.isRSS && post.externalUrl && (
            <a
              className="ext-link"
              onClick={(e) => {
                stop(e);
                const t = linkTarget();
                if (t === '_self') window.location.assign(post.externalUrl);
                else window.open(post.externalUrl, '_blank', 'noopener,noreferrer');
              }}
            >
              {(() => { try { return new URL(post.externalUrl).hostname.replace(/^www\./, ''); } catch { return 'оригинал'; } })()}
              <IconExternalLink size={11} stroke={2} />
            </a>
          )}
        </div>

        <h2 className="post-title">
          <Link
            to={`/p/${post.shortId || post.id}`}
            onClick={stop}
            style={{ textDecoration: 'none', color: 'inherit' }}
          >{post.title}</Link>
        </h2>
        {!compact && !post.isRSS && post.excerpt && (
          <p className="post-excerpt">{excerpt(post.excerpt, 220)}</p>
        )}
        {post.tags?.length > 0 && (
          <div className="tags">
            {post.tags.map((t) => (
              <span key={t} className="tag"><span className="hash">#</span>{t}</span>
            ))}
          </div>
        )}

        <div className="post-foot">
          <button type="button" className={'action up' + (voted ? ' active' : '')} onClick={toggleVote}>
            <IconChevronUp size={14} stroke={2} />
            <span className="num">{votes}</span>
          </button>
          <button type="button" className="action down" onClick={stop}>
            <IconChevronDown size={14} stroke={2} />
          </button>
          <button
            type="button"
            className="action"
            onClick={(e) => { stop(e); onOpen?.(post.shortId || post.id); }}
          >
            <IconMessage size={13} stroke={1.7} />
            <span className="num">{post.comments}</span> комментариев
          </button>
          <div className="foot-spacer" />
          <button type="button" className="action" onClick={handleShare} title="Поделиться">
            <IconUpload size={13} stroke={1.7} />
          </button>
          <button
            type="button"
            className="action"
            onClick={toggleSave}
            title={saved ? 'Сохранено' : 'Сохранить'}
            style={saved ? { color: 'var(--hn-honey-dark)' } : undefined}
          >
            {saved ? <IconBookmarkFilled size={13} /> : <IconBookmark size={13} stroke={1.7} />}
          </button>
          {!post.isRSS && post.author && (
            <div ref={menuRef} style={{ position: 'relative' }}>
              <button
                type="button"
                className="action"
                onClick={(e) => { stop(e); setMenuOpen(o => !o); }}
                title="Ещё"
              >
                <IconDots size={13} stroke={1.7} />
              </button>
              {menuOpen && (
                <div
                  onClick={stop}
                  style={{
                    position: 'absolute', right: 0, top: 'calc(100% + 4px)',
                    background: 'var(--bg)', border: '1px solid var(--border)',
                    borderRadius: 8, boxShadow: '0 6px 18px rgba(0,0,0,0.08)',
                    padding: 4, zIndex: 10, minWidth: 200,
                  }}
                >
                  <button
                    type="button"
                    onClick={handleBlock}
                    style={{
                      display: 'flex', alignItems: 'center', gap: 8, width: '100%',
                      padding: '8px 10px', border: 0, background: 'transparent',
                      cursor: 'pointer', fontSize: 13, color: '#B23A48',
                      borderRadius: 6, textAlign: 'left',
                    }}
                  >
                    <IconBan size={14} stroke={1.8} /> Не показывать от @{post.author}
                  </button>
                  {me?.isAdmin && (
                    <button
                      type="button"
                      onClick={handleAdminDelete}
                      style={{
                        display: 'flex', alignItems: 'center', gap: 8, width: '100%',
                        padding: '8px 10px', border: 0, background: 'transparent',
                        cursor: 'pointer', fontSize: 13, color: '#7B2D2D',
                        borderRadius: 6, textAlign: 'left',
                      }}
                    >
                      <IconTrash size={14} stroke={1.8} /> Удалить пост (admin)
                    </button>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {showThumb && (
        <div className="post-thumb">
          <img src={post.thumb} alt="" />
        </div>
      )}
    </article>
  );
};

export default PostCard;
