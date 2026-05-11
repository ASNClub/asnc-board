// Apply visual settings from localStorage to the document.
// Called on app boot and after each settings change.

const ACCENT_COLORS = {
  honey:     { base: '#B5781C', dark: '#7F5410', pale: '#F1DDB0', bright: '#C68A2C' },
  green:     { base: '#8FAA78', dark: '#6B8A57', pale: '#D2E0C5', bright: '#9CB888' },
  indigo:    { base: '#6E8DAE', dark: '#4E7197', pale: '#C4D2E2', bright: '#83A0BD' },
  lavender:  { base: '#9F7B98', dark: '#7B5775', pale: '#D6C7D1', bright: '#AF8DA8' },
  terracota: { base: '#B97561', dark: '#904E39', pale: '#E5C9BD', bright: '#C58974' },
};

// Dark-theme accent overrides: pale/light/dark recomputed for dark surfaces
// so honey-pale doesn't blast eyes against #1A1814.
const DARK_ACCENTS = {
  honey:     { base: '#E09832', dark: '#E5B763', pale: '#3A2F1A', bright: '#F1B848', light: '#8A6B30' },
  green:     { base: '#A3BE8C', dark: '#B7CFA0', pale: '#1F2A1A', bright: '#B7CFA0', light: '#5E7350' },
  indigo:    { base: '#81A1C1', dark: '#9BB6CF', pale: '#1A2330', bright: '#9BB6CF', light: '#4A6280' },
  lavender:  { base: '#B48EAD', dark: '#C9A8C2', pale: '#2A1F2A', bright: '#C9A8C2', light: '#6E4E68' },
  terracota: { base: '#D08770', dark: '#DEA38C', pale: '#2E1F1A', bright: '#DEA38C', light: '#7A4F40' },
};

const FONT_SIZES = { s: '13px', m: '14.5px', l: '16px' };

export function applyAppearance() {
  const root = document.documentElement;
  const body = document.body;
  if (!root || !body) return;

  const theme = localStorage.getItem('hg.theme') || 'light';
  let resolved = theme;
  if (theme === 'auto') {
    resolved = window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  root.setAttribute('data-theme', resolved);

  const accent = localStorage.getItem('hg.accent') || 'honey';
  const lightPal = ACCENT_COLORS[accent] || ACCENT_COLORS.honey;
  const darkPal  = DARK_ACCENTS[accent]  || DARK_ACCENTS.honey;
  const palette = resolved === 'dark' ? darkPal : lightPal;
  root.style.setProperty('--hn-honey', palette.base);
  root.style.setProperty('--hn-honey-dark', palette.dark);
  root.style.setProperty('--hn-honey-pale', palette.pale);
  root.style.setProperty('--hn-honey-bright', palette.bright);
  root.style.setProperty('--hn-honey-light', palette.light ?? palette.bright);

  const fontSize = localStorage.getItem('hg.fontSize') || 'm';
  root.style.fontSize = FONT_SIZES[fontSize] || FONT_SIZES.m;

  body.classList.toggle('no-serif', localStorage.getItem('hg.serif') === '0');
  body.classList.toggle('no-anim',  localStorage.getItem('hg.anim')  === '0');
}

// Re-apply on system theme change if theme=auto.
export function watchSystemTheme() {
  if (!window.matchMedia) return;
  const mq = window.matchMedia('(prefers-color-scheme: dark)');
  const handler = () => {
    if ((localStorage.getItem('hg.theme') || 'light') === 'auto') applyAppearance();
  };
  mq.addEventListener?.('change', handler);
}
