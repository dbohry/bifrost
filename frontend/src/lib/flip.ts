/**
 * A short "tiles glide to their new places" animation (FLIP: first, last,
 * invert, play) for when the dashboard switches layout. Tiles are matched by
 * the app id in `data-app-id`. Tiles in the Favourites row are copies of tiles
 * further down, so they carry `data-flip-skip` and are left out.
 */
const tiles = (root: HTMLElement) => root.querySelectorAll<HTMLElement>('[data-app-id]:not([data-flip-skip] *)');

const reduced = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches;

/** Where every tile is now. Call before the layout changes. */
export function snapshot(root: HTMLElement): Map<string, DOMRect> {
  const rects = new Map<string, DOMRect>();
  for (const el of tiles(root)) rects.set(el.dataset.appId!, el.getBoundingClientRect());
  return rects;
}

/** Slides each tile from where it was to where it is now. Call after the layout has changed. */
export function play(root: HTMLElement, before: Map<string, DOMRect>) {
  if (reduced()) return;
  const options: KeyframeAnimationOptions = { duration: 420, easing: 'cubic-bezier(0.2, 0.8, 0.2, 1)' };

  for (const el of tiles(root)) {
    const from = before.get(el.dataset.appId!);
    const to = el.getBoundingClientRect();
    if (!from) {
      el.animate([{ opacity: 0, transform: 'scale(0.9)' }, { opacity: 1, transform: 'none' }], options);
      continue;
    }
    const dx = from.left - to.left;
    const dy = from.top - to.top;
    if (Math.abs(dx) < 1 && Math.abs(dy) < 1) continue;
    el.animate([{ transform: `translate(${dx}px, ${dy}px)` }, { transform: 'translate(0, 0)' }], options);
  }

  // Headings and the Favourites row appear with a fade; they have no old place to come from.
  for (const el of root.querySelectorAll<HTMLElement>('[data-flip-fade]')) {
    el.animate([{ opacity: 0 }, { opacity: 1 }], { duration: 300, delay: 120, easing: 'ease-out', fill: 'backwards' });
  }
}
