// Минимальный markdown-парсер: H2/H3, code-блоки, inline-code, bold/italic,
// blockquote, ul, ссылки, картинки, hr. Используется в редакторе постов и
// в треде для рендера тела поста и комментариев.
import { linkTarget } from './prefs';

export function renderMarkdown(src) {
  if (!src) return '';
  const target = linkTarget();
  const escape = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  const lines = src.split('\n');
  let out = '';
  let inCode = false;
  let inList = false;
  let para = [];

  const safeUrl = (url) => /^(https?:|\/|#)/i.test(url) ? url : '#';

  const inline = (text) => {
    let t = escape(text);
    t = t.replace(/`([^`]+)`/g, '<code>$1</code>');
    t = t.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    t = t.replace(/\*([^*]+)\*/g, '<em>$1</em>');
    t = t.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (_, alt, url) => `<img src="${safeUrl(url)}" alt="${alt}" />`);
    t = t.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_, label, url) => `<a href="${safeUrl(url)}" target="${target}" rel="noreferrer">${label}</a>`);
    return t;
  };

  const flushPara = () => {
    if (para.length === 0) return;
    out += `<p>${inline(para.join(' '))}</p>`;
    para = [];
  };

  for (const line of lines) {
    if (line.trim().startsWith('```')) {
      flushPara();
      if (inList) { out += '</ul>'; inList = false; }
      if (inCode) { out += '</code></pre>'; inCode = false; }
      else        { out += '<pre><code>'; inCode = true; }
      continue;
    }
    if (inCode) { out += escape(line) + '\n'; continue; }

    if (/^#{1,3}\s/.test(line)) {
      flushPara();
      if (inList) { out += '</ul>'; inList = false; }
      const m = line.match(/^(#{1,3})\s+(.*)$/);
      out += `<h${m[1].length}>${inline(m[2])}</h${m[1].length}>`;
      continue;
    }
    if (/^>\s?/.test(line)) {
      flushPara();
      if (inList) { out += '</ul>'; inList = false; }
      out += `<blockquote>${inline(line.replace(/^>\s?/, ''))}</blockquote>`;
      continue;
    }
    if (/^[-*]\s+/.test(line)) {
      flushPara();
      if (!inList) { out += '<ul>'; inList = true; }
      out += `<li>${inline(line.replace(/^[-*]\s+/, ''))}</li>`;
      continue;
    }
    if (/^---+\s*$/.test(line)) {
      flushPara();
      if (inList) { out += '</ul>'; inList = false; }
      out += '<hr/>';
      continue;
    }
    if (line.trim() === '') {
      flushPara();
      if (inList) { out += '</ul>'; inList = false; }
      continue;
    }
    para.push(line);
  }
  flushPara();
  if (inList) out += '</ul>';
  if (inCode) out += '</code></pre>';
  return out;
}
