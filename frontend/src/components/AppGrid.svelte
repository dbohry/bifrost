<script lang="ts">
  import type { AppItem } from '../lib/api';
  import { sortable } from '../lib/sortable';
  import AppTile from './AppTile.svelte';

  type Props = {
    apps: AppItem[];
    /** Whether tiles can be dragged to a new place. */
    reorderable?: boolean;
    onedit: (app: AppItem) => void;
    onfavourite: (app: AppItem) => void;
    onmove?: (from: number, to: number) => void;
    /** Shows a dashed "Add app" tile after the last one. */
    onadd?: () => void;
  };

  let { apps, reorderable = false, onedit, onfavourite, onmove, onadd }: Props = $props();
</script>

<div
  class="grid"
  class:reorderable
  use:sortable={{ enabled: reorderable, onmove: (from, to) => onmove?.(from, to) }}
>
  {#each apps as app (app.id)}
    <AppTile {app} onedit={() => onedit(app)} onfavourite={() => onfavourite(app)} />
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
