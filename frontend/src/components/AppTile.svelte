<script lang="ts">
  import type { AppItem } from '../lib/api';
  import AppIcon from './AppIcon.svelte';

  let {
    app,
    editing = false,
    onedit,
    onfavourite,
  }: { app: AppItem; editing?: boolean; onedit: () => void; onfavourite: () => void } = $props();

  // While editing on a touch screen, a tap changes the app instead of opening it.
  function open(e: MouseEvent) {
    if (!editing) return;
    e.preventDefault();
    onedit();
  }
</script>

<div class="tile" class:editing data-sortable-item data-app-id={app.id}>
  <a
    class="open"
    href={app.url}
    draggable="false"
    target={app.newTab ? '_blank' : undefined}
    rel="noopener noreferrer"
    onclick={open}
  >
    <AppIcon name={app.name} seed={app.id} icon={app.icon} />
    <span class="name">{app.name}</span>
  </a>

  <!-- Touch screens: a small star on pinned apps, since the button itself is tucked away until Edit. -->
  {#if app.favourite}<span class="badge" aria-hidden="true">★</span>{/if}

  <button
    class="action fav"
    class:on={app.favourite}
    type="button"
    aria-label={app.favourite ? `Remove ${app.name} from favourites` : `Add ${app.name} to favourites`}
    aria-pressed={app.favourite}
    onclick={onfavourite}
  >
    <svg width="16" height="16" viewBox="0 0 24 24" aria-hidden="true">
      <path d="M12 2.5l2.9 6.1 6.6.8-4.9 4.6 1.3 6.6L12 17.3l-5.9 3.3 1.3-6.6L2.5 9.4l6.6-.8z" />
    </svg>
  </button>
  <button class="action edit" type="button" aria-label={`Edit ${app.name}`} onclick={onedit}>
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M4 20h4L19 9l-4-4L4 16z" />
      <path d="M13.5 6.5l4 4" />
    </svg>
  </button>
</div>

<style>
  .tile {
    position: relative;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    transition: transform 0.12s, border-color 0.12s;
  }
  .tile:hover,
  .tile:focus-within {
    border-color: var(--accent);
  }
  @media (hover: hover) and (prefers-reduced-motion: no-preference) {
    .tile:hover {
      transform: translateY(-2px);
    }
  }
  .open {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    padding: 24px 12px 20px;
    color: inherit;
    text-decoration: none;
    border-radius: 14px;
  }
  .open:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .name {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 14px;
    font-weight: 600;
  }

  /* 44px touch targets, with a small icon in the corner of each. */
  .action {
    position: absolute;
    top: 0;
    width: 44px;
    height: 44px;
    border: 0;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    display: grid;
    place-items: center;
    border-radius: 14px;
  }
  .action:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }
  .fav {
    right: 0;
  }
  .fav svg {
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linejoin: round;
  }
  .fav.on {
    color: #e0a100;
  }
  .fav.on svg {
    fill: currentColor;
  }
  .edit {
    left: 0;
  }

  /* With a mouse, the pencil only shows on hover or keyboard focus. */
  @media (hover: hover) {
    .edit {
      opacity: 0;
    }
    .tile:hover .edit,
    .tile:focus-within .edit {
      opacity: 1;
    }
  }

  /*
   * On a touch screen there is no hover, and two buttons on every small tile
   * would crowd it. They appear when Edit is switched on, and a tap on the tile
   * then opens its details.
   */
  .badge {
    display: none;
    position: absolute;
    top: 8px;
    right: 10px;
    font-size: 14px;
    line-height: 1;
    color: #e0a100;
    pointer-events: none;
  }
  @media (hover: none) {
    .tile:not(.editing) .action {
      display: none;
    }
    .tile:not(.editing) .badge {
      display: block;
    }
    .tile.editing {
      border-style: dashed;
      border-color: var(--accent);
    }
  }
</style>
