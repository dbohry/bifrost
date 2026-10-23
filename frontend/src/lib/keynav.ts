/**
 * Moving around with the arrow keys alone. From whatever has focus, the arrow
 * goes to the nearest control in that direction, judged by where things are on
 * the screen: the top bar, search box, layout switch, category tabs, tiles,
 * dialogs and settings all work the same way, with no per-screen code.
 *
 * Things an arrow already does stay as they are: the caret in a text field
 * (until it reaches the edge of the text), a menu. Anything marked `data-nav-skip`
 * is left out (the hover buttons on a tile, which have their own keys). With nothing
 * focused, an arrow enters the page, at the element marked `data-nav-start` if there is one.
 *
 * Up and down go to the nearest row, and within it to the control closest to
 * where you are. Otherwise the controls on a row far from the left edge (the top
 * bar, the category tabs) could never be reached from the tiles below them.
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
    if (el.type === 'range') return true;
    // A text field keeps left and right for the caret, until the caret has nowhere left to go,
    // and up and down when it offers suggestions.
    if (TEXT_TYPES.has(el.type)) {
      if (el.hasAttribute('list')) return true;
      if (dir === 'up' || dir === 'down') return false;
      const { selectionStart: start, selectionEnd: end } = el;
      if (start === null || end === null || start !== end) return true;
      return dir === 'left' ? start > 0 : start < el.value.length;
    }
  }
  return el.isContentEditable;
}

/** The radio buttons that share a group with this one. */
function groupOf(radio: HTMLInputElement): HTMLInputElement[] {
  if (!radio.name) return [radio];
  const root = radio.form ?? radio.getRootNode();
  return [...(root as ParentNode).querySelectorAll<HTMLInputElement>('input[type=radio]')].filter((r) => r.name === radio.name);
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
  if (!(here instanceof HTMLElement)) return;

  // An open dialog is its own little world.
  const dialog = document.querySelector('dialog[open]');
  const scope: ParentNode = dialog ?? document;

  // With nothing focused (just signed in, or the control that had focus is gone), an arrow
  // goes into the page instead of doing nothing.
  if (here === document.body) {
    const start = scope.querySelector('[data-nav-start]') ?? scope.querySelector('main') ?? scope;
    const first = [...start.querySelectorAll<HTMLElement>(FOCUSABLE)].find(usable);
    if (first) {
      e.preventDefault();
      first.focus();
    }
    return;
  }
  if (keepsArrow(here, dir) || (dialog && !dialog.contains(here))) return;

  const from = here.getBoundingClientRect();
  const cx = from.left + from.width / 2;
  const cy = from.top + from.height / 2;
  const vertical = dir === 'up' || dir === 'down';

  type Candidate = { el: HTMLElement; along: number; across: number };
  const found: Candidate[] = [];

  for (const el of scope.querySelectorAll<HTMLElement>(FOCUSABLE)) {
    if (el === here || !usable(el)) continue;
    const r = el.getBoundingClientRect();
    const dx = r.left + r.width / 2 - cx;
    const dy = r.top + r.height / 2 - cy;

    // How far along the arrow, and how far off to the side.
    let along: number;
    let across: number;
    if (dir === 'down') [along, across] = [dy, Math.abs(dx)];
    else if (dir === 'up') [along, across] = [-dy, Math.abs(dx)];
    else if (dir === 'right') [along, across] = [dx, Math.abs(dy)];
    else [along, across] = [-dx, Math.abs(dy)];

    // The candidate must be clearly further along than half the size of what has focus,
    // so neighbours in the same row or column never count as "below" or "beside".
    const size = vertical ? from.height : from.width;
    if (along < size * 0.5) continue;
    // Sideways, only things in the same row.
    if (!vertical && across > Math.max(from.height, r.height) * 0.75) continue;
    found.push({ el, along, across });
  }

  let best: HTMLElement | null = null;
  if (vertical) {
    // The nearest row, and in it whatever is closest to where we are.
    const nearest = Math.min(...found.map((c) => c.along));
    const row = Math.max(16, from.height * 0.6);
    let bestAcross = Infinity;
    for (const c of found) {
      if (c.along > nearest + row || c.across >= bestAcross) continue;
      bestAcross = c.across;
      best = c.el;
    }
  } else {
    let bestScore = Infinity;
    for (const c of found) {
      const score = c.along + c.across * 4;
      if (score < bestScore) {
        bestScore = score;
        best = c.el;
      }
    }
  }

  const radio = here instanceof HTMLInputElement && here.type === 'radio' ? here : null;
  if (best) {
    e.preventDefault();
    best.focus();
    // Within a group of radio buttons an arrow also chooses, as it does natively.
    if (radio && best instanceof HTMLInputElement && best.type === 'radio' && groupOf(radio).includes(best)) best.click();
  } else if (radio) {
    // Nothing lies that way on the screen: do what the browser does, to the next or previous
    // one in the group, but stop at the ends instead of wrapping round.
    e.preventDefault();
    const group = groupOf(radio).filter(usable);
    const next = group[group.indexOf(radio) + (dir === 'down' || dir === 'right' ? 1 : -1)];
    if (next) {
      next.focus();
      next.click();
    }
  }
}
