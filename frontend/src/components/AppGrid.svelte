<script lang="ts">
  import type { AppItem } from '../lib/api';
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
    /** Shows a dashed "Add app" tile after the last one. */
    onadd?: () => void;
  };

  let { apps, reorderable = false, editing = false, onedit, onfavourite, onmove, group, list, onmoveout, onadd }: Props = $props();

  let grid = $state<HTMLElement>();

  /**
   * Arrow keys move between the tiles, the way the eye sees them. Tab still
   * visits every control. Left and right go to the neighbour in the row, and up
   * and down to the tile nearest above or below.
   */
  function arrows(e: KeyboardEvent) {
    if (e.altKey || e.ctrlKey || e.metaKey || e.shiftKey) return;
    const here = (e.target as HTMLElement).closest<HTMLAnchorElement>('a.open');
    if (!here || !grid) return;

    const links = [...grid.querySelectorAll<HTMLAnchorElement>('a.open')];
    const at = links.indexOf(here);
    const box = here.getBoundingClientRect();
    let next: HTMLAnchorElement | undefined;

    if (e.key === 'ArrowRight') next = links[at + 1];
    else if (e.key === 'ArrowLeft') next = links[at - 1];
    else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
      const down = e.key === 'ArrowDown';
      const row = links
        .map((el) => ({ el, r: el.getBoundingClientRect() }))
        .filter(({ r }) => (down ? r.top > box.bottom - 4 : r.bottom < box.top + 4))
        // the nearest row first, then the tile whose centre is closest
        .sort((a, b) => (down ? a.r.top - b.r.top : b.r.top - a.r.top) || Math.abs(a.r.left - box.left) - Math.abs(b.r.left - box.left));
      const first = row[0];
      // only tiles in that nearest row count
      next = first && row.filter(({ r }) => Math.abs(r.top - first.r.top) < 4).sort((a, b) => Math.abs(a.r.left - box.left) - Math.abs(b.r.left - box.left))[0]?.el;
    }

    if (next) {
      e.preventDefault();
      next.focus();
    }
  }
</script>

<div
  class="grid"
  class:reorderable
  bind:this={grid}
  onkeydown={arrows}
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
