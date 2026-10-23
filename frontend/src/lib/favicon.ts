import type { LogoKind } from './api';

// The browser tab icon is tiny, so every logo uses the same simple mark there:
// the arches on a tile. (The three-arch logo blurs at 16 px, and the five
// colours of the spectrum logo turn muddy.) Spectrum uses the brand blue.
const brand = { accent: '#3b5bdb', onAccent: '#ffffff' };

function icon(accent: string, onAccent: string): string {
  return (
    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">' +
    `<rect x="4" y="4" width="88" height="88" rx="22" fill="${accent}"/>` +
    `<g fill="none" stroke="${onAccent}" stroke-width="7" stroke-linecap="round">` +
    '<path d="M20 63A28 28 0 0 1 76 63"/><path d="M33 63A15 15 0 0 1 63 63"/></g>' +
    `<circle cx="48" cy="63" r="4.5" fill="${onAccent}"/></svg>`
  );
}

/** Points the tab icon at the user's chosen logo, in their theme's colours. Pass null for the default. */
export function setFavicon(kind: LogoKind, colors: Record<string, string> | null) {
  const link = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
  if (!link) return;

  if (!colors) {
    link.type = 'image/svg+xml';
    link.href = '/favicon.svg';
    return;
  }
  const { accent, onAccent } = kind === 'spectrum' ? brand : { accent: colors.accent, onAccent: colors.onAccent };
  link.type = 'image/svg+xml';
  link.href = 'data:image/svg+xml,' + encodeURIComponent(icon(accent, onAccent));
}
