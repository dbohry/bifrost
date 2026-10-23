/**
 * Moving around with the arrow keys alone. From whatever has focus, the arrow
 * goes to the nearest control in that direction, judged by where things are on
 * the screen: the top bar, search box, layout switch, category tabs, tiles,
 * dialogs and settings all work the same way, with no per-screen code.
 *
 * Things an arrow already does stay as they are: the caret in a text field,
 * a group of radio buttons, a menu. Anything marked `data-nav-skip` is left
 * out (the hover buttons on a tile, which have their own keys).
 */
const FOCUSABLE = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled]):not([type=hidden])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',');

type Direction = 'up' | 'down' | 'left' | 'right';
const keys: Record<string, Direction> = { ArrowUp: 'up', ArrowDown: 'down', ArrowLeft: 'left', ArrowRight: 'right' };

const TEXT_TYPES = new Set(['text', 'search', 'email', 'url', 'tel', 'password', 'number', '']);

/** Whether the arrow already means something to this control. */
function keepsArrow(el: HTMLElement, dir: Direction): boolean {
  if (el instanceof HTMLSelectElement || el instanceof HTMLTextAreaElement) return true;
  if (el instanceof HTMLInputElement) {
    if (el.type === 'radio' || el.type === 'range') return true;
    // A text field keeps left and right for the caret, and up and down when it offers suggestions.
    if (TEXT_TYPES.has(el.type)) return dir === 'left' || dir === 'right' || el.hasAttribute('list');
  }
  return el.isContentEditable;
}

function usable(el: HTMLElement): boolean {
  if (el.closest('[data-nav-skip], [inert], [hidden]')) return false;
  const r = el.getBoundingClientRect();
  if (r.width === 0 || r.height === 0) return false;
  return getComputedStyle(el).visibility !== 'hidden';
}

export function arrowNav(e: KeyboardEvent) {
  const dir = keys[e.key];
  if (!dir || e.defaultPrevented || e.altKey || e.ctrlKey || e.metaKey || e.shiftKey) return;

  const here = e.target;
  if (!(here instanceof HTMLElement) || here === document.body || keepsArrow(here, dir)) return;

  // An open dialog is its own little world.
  const dialog = document.querySelector('dialog[open]');
  if (dialog && !dialog.contains(here)) return;
  const scope: ParentNode = dialog ?? document;

  const from = here.getBoundingClientRect();
  const cx = from.left + from.width / 2;
  const cy = from.top + from.height / 2;

  let best: HTMLElement | null = null;
  let bestScore = Infinity;

  for (const el of scope.querySelectorAll<HTMLElement>(FOCUSABLE)) {
    if (el === here || !usable(el)) continue;
    const r = el.getBoundingClientRect();
    const dx = r.left + r.width / 2 - cx;
    const dy = r.top + r.height / 2 - cy;

    // How far along the arrow, and how far off to the side. The candidate must be clearly
    // further along than half the size of what has focus, so neighbours in the same row
    // or column never count as "below" or "beside".
    let along: number;
    let across: number;
    if (dir === 'down') [along, across] = [dy, Math.abs(dx)];
    else if (dir === 'up') [along, across] = [-dy, Math.abs(dx)];
    else if (dir === 'right') [along, across] = [dx, Math.abs(dy)];
    else [along, across] = [-dx, Math.abs(dy)];

    const size = dir === 'up' || dir === 'down' ? from.height : from.width;
    if (along < size * 0.5) continue;
    // Sideways, only things in the same row. Vertically, anything below counts, but the column matters more.
    if ((dir === 'left' || dir === 'right') && across > Math.max(from.height, r.height) * 0.75) continue;
    const score = along + across * (dir === 'left' || dir === 'right' ? 4 : 2);
    if (score < bestScore) {
      bestScore = score;
      best = el;
    }
  }

  if (best) {
    e.preventDefault();
    best.focus();
  }
}
