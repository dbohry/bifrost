import type { LogoKind } from './api';

// The tab icon is the user's chosen logo, drawn a little bolder than in the page
// because it is shown at 16 px. The app icon is the arches on a tile, the arches
// are bare in the theme's accent colour, and the spectrum keeps its five colours.
const svg = (inner: string, top = 0) =>
  `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 ${top} 96 96">${inner}</svg>`;

function tile(accent: string, onAccent: string): string {
  return svg(
    `<rect x="4" y="4" width="88" height="88" rx="22" fill="${accent}"/>` +
      `<g fill="none" stroke="${onAccent}" stroke-width="7" stroke-linecap="round">` +
      '<path d="M20 63A28 28 0 0 1 76 63"/><path d="M33 63A15 15 0 0 1 63 63"/></g>' +
      `<circle cx="48" cy="63" r="4.5" fill="${onAccent}"/>`,
  );
}

function arches(accent: string): string {
  return svg(
    `<g fill="none" stroke="${accent}" stroke-width="11" stroke-linecap="round">` +
      '<path d="M10 76A38 38 0 0 1 86 76"/><path d="M26 76A22 22 0 0 1 70 76"/><path d="M42 76A6 6 0 0 1 54 76"/></g>',
    9, // the arches sit in the lower half of the square, so move the view up to centre them
  );
}

const rainbow = ['#e5484d', '#f59f00', '#2f9e63', '#3b9ddd', '#7048e8'];

function spectrum(): string {
  const bands = rainbow
    .map((colour, i) => {
      const r = 40 - i * 7;
      return `<path d="M${48 - r} 78A${r} ${r} 0 0 1 ${48 + r} 78" stroke="${colour}"/>`;
    })
    .join('');
  return svg(`<g fill="none" stroke-width="7" stroke-linecap="round">${bands}</g>`, 9);
}

/** Points the tab icon at the user's chosen logo, in their theme's colours. Pass null for the default. */
export function setFavicon(kind: LogoKind, colors: Record<string, string> | null) {
  const link = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
  if (!link) return;

  link.type = 'image/svg+xml';
  if (!colors) {
    link.href = '/favicon.svg';
    return;
  }
  const drawing =
    kind === 'spectrum' ? spectrum() : kind === 'arches' ? arches(colors.accent) : tile(colors.accent, colors.onAccent);
  link.href = 'data:image/svg+xml,' + encodeURIComponent(drawing);
}
