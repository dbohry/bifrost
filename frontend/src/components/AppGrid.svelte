<script lang="ts">
  import { tick } from 'svelte';
  import type { AppItem } from '../lib/api';
  import { ui } from '../lib/ui.svelte';
  import { sortable } from '../lib/sortable';
  import AppTile from './AppTile.svelte';

  type Props = {
    apps: AppItem[];
    /** Whether tiles can be dragged to a new place. */
    reorderable?: boolean;
    /** Edit mode on a touch screen: tiles show their buttons. */
    editing?: boolean;
    onedit: (app: AppItem) => void;
    onfavourite: (app: AppItem) => void;
    onmove?: (from: number, to: number) => void;
    /** Grids with the same `group` let a tile be dragged from one to another; `list` names this one. */
    group?: string;
    list?: string;
    onmoveout?: (fromList: string, toList: string, from: number, to: number) => void;
    /** Alt+Up or Alt+Down on the first or last row: move the tile into the category above or below. */
    onmoveedge?: (index: number, direction: 'up' | 'down') => void;
    /** Shows a dashed "Add app" tile after the last one. */
    onadd?: () => void;
  };

  let { apps, reorderable = false, editing = false, onedit, onfavourite, onmove, group, list, onmoveout, onmoveedge, onadd }: Props = $props();

  let grid = $state<HTMLElement>();

  /**
   * Keys for the tile that has focus. The arrow keys alone move focus (see keynav).
   * With Alt held, they move the tile itself: left and right one place, up and down
   * one row, and past the first or last row into the neighbouring category. E opens
   * the app's details and F stars it, because the two buttons are for the mouse.
   */
  async function keys(e: KeyboardEvent) {
    const link = (e.target as HTMLElement).closest<HTMLAnchorElement>('a.open');
    if (!link || !grid || e.ctrlKey || e.metaKey) return;
    const tile = link.closest<HTMLElement>('[data-sortable-item]');
    const items = [...grid.querySelectorAll<HTMLElement>('[data-sortable-item]')];
    const at = tile ? items.indexOf(tile) : -1;
    const app = apps[at];
    if (!app) return;

    if (!e.altKey && !e.shiftKey && (e.key === 'e' || e.key === 'f')) {
      e.preventDefault();
      if (e.key === 'e') onedit(app);
      else onfavourite(app);
      return;
    }

    const arrow = ['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].indexOf(e.key);
    if (!e.altKey || e.shiftKey || arrow < 0 || !reorderable) return;
    e.preventDefault();

    const first = items[0].getBoundingClientRect().top;
    const columns = Math.max(1, items.filter((el) => Math.abs(el.getBoundingClientRect().top - first) < 4).length);
    const to = [at - 1, at + 1, at - columns, at + columns][arrow];

    if (to >= 0 && to < items.length) {
      onmove?.(at, to);
      say(app.name, to + 1, items.length);
    } else if (arrow >= 2 && onmoveedge) {
      onmoveedge(at, arrow === 2 ? 'up' : 'down');
    } else {
      return;
    }

    // The tile may have been rebuilt in its new place, so find it again and keep focus on it.
    await tick();
    document.querySelector<HTMLAnchorElement>(`[data-app-id="${app.id}"]:not([data-flip-skip] *) a.open`)?.focus();
  }

  function say(name: string, place: number, count: number) {
    ui.announcement = `${name}, place ${place} of ${count}`;
  }
</script>

<div
  class="grid"
  class:reorderable
  bind:this={grid}
  onkeydown={keys}
  role="presentation"
  data-list={list}
  use:sortable={{ enabled: reorderable, group, onmove: (from, to) => onmove?.(from, to), onmoveout }}
>
  {#each apps as app (app.id)}
    <AppTile {app} {editing} onedit={() => onedit(app)} onfavourite={() => onfavourite(app)} />
  {/each}
  {#if onadd}
    <button class="add-tile" onclick={onadd}>
      <span class="plus" aria-hidden="true">+</span>
      Add app
    </button>
  {/if}
</div>

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 16px;
  }
  @media (max-width: 480px) {
    .grid {
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 12px;
    }
  }
  @media (hover: hover) {
    .reorderable :global([data-sortable-item]) {
      cursor: grab;
    }
  }
  /* The place the dragged tile will drop into. */
  .grid :global(.sortable-ghost) {
    opacity: 0.35;
  }
  /* Sortable's floating copy of the tile follows the pointer. */
  :global(.sortable-fallback) {
    cursor: grabbing;
    box-shadow: 0 8px 24px rgb(0 0 0 / 0.25);
    border-radius: 14px;
  }

  .add-tile {
    min-height: 120px;
    border: 2px dashed var(--border);
    border-radius: 14px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 14px;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
  }
  .add-tile:hover {
    border-color: var(--accent);
    color: var(--text);
  }
  .add-tile:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .plus {
    font-size: 26px;
    line-height: 1;
  }
</style>
