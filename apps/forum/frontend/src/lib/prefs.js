export function linkTarget() {
  return localStorage.getItem('hg.linkTarget') === 'self' ? '_self' : '_blank';
}

export function getHiddenTags() {
  try {
    const arr = JSON.parse(localStorage.getItem('hg.hiddenTags') || '[]');
    return Array.isArray(arr) ? arr.map(t => String(t).toLowerCase()) : [];
  } catch { return []; }
}

export function getShowRSS() {
  return localStorage.getItem('hg.showRSS') !== '0';
}
