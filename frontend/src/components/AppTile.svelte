<script lang="ts">
  import type { AppItem } from '../lib/api';
  import AppIcon from './AppIcon.svelte';
  import { api, type StatusHistory } from '../lib/api';
  import { statuses } from '../lib/status.svelte';
  import { span } from '../lib/time';

  let {
    app,
    editing = false,
    onedit,
    onfavourite,
  }: { app: AppItem; editing?: boolean; onedit: () => void; onfavourite: () => void } = $props();

  // The latest look at the app, if it asked to be checked and has been looked at.
  const result = $derived(app.checkEnabled ? statuses.of(app.id) : undefined);
  const down = $derived(result?.state === 'down');
  const label = $derived(
    result?.state === 'up' ? 'Online' : result?.state === 'down' ? 'Offline' : result ? 'Checking' : '',
  );

  // The card is centred under the tile. For a tile at the edge of the page, slide it back into view.
  let card = $state<HTMLElement>();
  // The last 24 hours, fetched when the card is about to show and not more than every 30 seconds.
  let history = $state<StatusHistory | null>(null);
  let fetchedAt = 0;
  const empty = Array.from({ length: 24 }, () => ({ start: 0, checks: 0, failures: 0 }));
  const bars = $derived(history?.hours ?? empty);

  const clock = (unix: number) => new Date(unix * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  function tip(h: { start: number; checks: number; failures: number }) {
    if (h.start === 0) return '';
    const when = `${clock(h.start)} to ${clock(h.start + 3600)}`;
    if (h.checks === 0) return `${when} · no checks`;
    if (h.failures === 0) return `${when} · all ${h.checks} checks passed`;
    return `${when} · ${h.failures} of ${h.checks} checks failed`;
  }

  function reveal() {
    if (Date.now() - fetchedAt > 30_000) {
      fetchedAt = Date.now();
      api.statusHistory(app.id).then((h) => (history = h)).catch(() => (fetchedAt = 0));
    }
    keepInView();
  }

  function keepInView() {
    requestAnimationFrame(() => {
      if (!card) return;
      card.style.setProperty('--shift', '0px');
      const r = card.getBoundingClientRect();
      const margin = 8;
      const shift = r.left < margin ? margin - r.left : r.right > innerWidth - margin ? innerWidth - margin - r.right : 0;
      card.style.setProperty('--shift', `${shift}px`);
    });
  }

  // While editing on a touch screen, a tap changes the app instead of opening it.
  function open(e: MouseEvent) {
    if (!editing) return;
    e.preventDefault();
    onedit();
  }
</script>

<div class="tile" class:editing class:down data-sortable-item data-app-id={app.id}>
  <a
    class="open"
    href={app.url}
    draggable="false"
    target={app.newTab ? '_blank' : undefined}
    rel="noopener noreferrer"
    onclick={open}
    onfocus={reveal}
  >
    <span class="icon-wrap">
      <AppIcon name={app.name} seed={app.id} icon={app.icon} />
      {#if result}
        <span class="status {result.state}" role="img" aria-label={label} onmouseenter={reveal}>{down ? '×' : ''}</span>
      {/if}
    </span>
    <span class="name">{app.name}</span>
    {#if down && result}
      <span class="note">Offline · {span(result.since)}</span>
    {/if}
  </a>

  {#if result}
    <!-- Hovering the badge, or focusing the tile with the keyboard, shows the details. It is decoration for the mouse; the badge and note carry the meaning. -->
    <div class="card" bind:this={card} aria-hidden="true" data-nav-skip>
      <div class="card-title"><span class="status-dot {result.state}"></span>{app.name} is {label.toLowerCase()}</div>
      <dl>
        {#if result.state === 'down'}
          <dt>Since</dt><dd>{span(result.since)} ago</dd>
        {/if}
        <dt>Last check</dt><dd>{span(result.lastChecked)} ago</dd>
        {#if result.state === 'up' && result.latencyMs > 0}
          <dt>Answered in</dt><dd>{result.latencyMs} ms</dd>
        {/if}
        {#if result.state === 'down' && result.reason}
          <dt>Reason</dt><dd>{result.reason}</dd>
        {/if}
        <dt>Checked</dt><dd class="addr">{result.address}</dd>
      </dl>
      <div class="hist">
        <div class="hist-head">
          <span>Last 24 hours</span>
          {#if history?.uptime != null}<strong>{history.uptime}% online</strong>{/if}
        </div>
        <div class="bars">
          {#each bars as h, i (i)}
            <div class="bar" class:none={h.checks === 0} data-tip={tip(h) || undefined}>
              {#if h.failures > 0}<div class="fail" style:height="{Math.max(100 * h.failures / h.checks, 12)}%"></div>{/if}
            </div>
          {/each}
        </div>
        <div class="hist-foot"><span>24 h ago</span><span>now</span></div>
      </div>
    </div>
  {/if}

  <!-- Touch screens: a small star on pinned apps, since the button itself is tucked away until Edit. -->
  {#if app.favourite}<span class="badge" aria-hidden="true">★</span>{/if}

  <button
    class="action fav"
    data-nav-skip
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
  <button class="action edit" data-nav-skip type="button" aria-label={`Edit ${app.name}`} onclick={onedit}>
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
    /* Lift the whole tile, and so its status card, above the tiles after it in the grid. */
    z-index: 5;
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
  /* A small phone fits three tiles a row only with less padding. */
  @media (max-width: 359px) {
    .open {
      padding: 20px 6px 16px;
    }
  }
  .open:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .icon-wrap {
    position: relative;
    display: block;
  }
  /* The badge sits on the corner of the icon. The ring in the tile's colour cuts it out of the icon. */
  .status {
    position: absolute;
    right: -6px;
    bottom: -6px;
    width: 18px;
    height: 18px;
    box-sizing: border-box;
    border-radius: 50%;
    border: 3px solid var(--surface);
    display: grid;
    place-items: center;
    font-size: 10px;
    font-weight: 800;
    line-height: 1;
    color: var(--bg);
  }
  .status.up {
    background: var(--ok);
  }
  .status.down {
    background: var(--bad);
  }
  .status.checking {
    background: var(--muted);
    opacity: 0.55;
  }
  .down {
    border-color: var(--bad);
  }
  .down .name {
    color: var(--muted);
  }
  .note {
    margin-top: -4px;
    font-size: 11px;
    font-weight: 600;
    color: var(--bad);
  }

  /* Hidden cards linger for a moment, so the mouse can travel from the badge onto the card. */
  .card {
    visibility: hidden;
    opacity: 0;
    transition:
      opacity 0.12s,
      visibility 0s linear 0.25s;
    position: absolute;
    z-index: 4;
    top: calc(100% + 8px);
    left: 50%;
    transform: translateX(calc(-50% + var(--shift, 0px)));
    width: max-content;
    min-width: 240px;
    max-width: min(320px, calc(100vw - 16px));
    box-sizing: border-box;
    padding: 12px 14px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 12px;
    box-shadow: 0 10px 28px rgb(0 0 0 / 0.3);
    font-size: 12px;
  }
  @media (hover: hover) {
    .open:has(.status:hover) ~ .card,
    .open:focus-visible ~ .card,
    .card:hover {
      visibility: visible;
      opacity: 1;
      transition-delay: 0s;
    }
  }
  .card-title {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
    font-size: 13px;
    font-weight: 700;
  }
  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
    background: var(--muted);
  }
  .status-dot.up {
    background: var(--ok);
  }
  .status-dot.down {
    background: var(--bad);
  }
  .hist {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-top: 12px;
    padding-top: 10px;
    border-top: 1px solid var(--border);
    color: var(--muted);
  }
  .hist-head,
  .hist-foot {
    display: flex;
    justify-content: space-between;
    gap: 8px;
  }
  .hist-head strong {
    color: var(--text);
  }
  .hist-foot {
    font-size: 11px;
  }
  .bars {
    display: flex;
    gap: 2px;
    height: 28px;
  }
  .bar {
    position: relative;
    flex: 1;
    min-width: 0;
    border-radius: 2px;
    background: var(--ok);
  }
  .bar.none {
    background: transparent;
    outline: 1px dashed var(--border);
    outline-offset: -1px;
  }
  .fail {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    min-height: 3px;
    border-radius: 0 0 2px 2px;
    background: var(--bad);
  }
  /* A small label over the bar the mouse is on. */
  .bar[data-tip]:hover::after {
    content: attr(data-tip);
    position: absolute;
    bottom: calc(100% + 6px);
    left: 50%;
    transform: translateX(-50%);
    z-index: 2;
    padding: 5px 9px;
    border-radius: 8px;
    background: var(--text);
    color: var(--bg);
    font-size: 11px;
    white-space: nowrap;
    pointer-events: none;
  }
  dl {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 4px 12px;
    margin: 0;
    color: var(--muted);
  }
  dd {
    margin: 0;
    color: var(--text);
    overflow-wrap: anywhere;
  }
  /* Up to two lines, so long names still read on a narrow phone tile. */
  .name {
    max-width: 100%;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    overflow: hidden;
    overflow-wrap: break-word;
    hyphens: auto;
    text-align: center;
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

  /*
   * With a mouse, the pencil and an empty star only show on hover or keyboard
   * focus. A filled star stays, since it says the app is a favourite.
   */
  @media (hover: hover) {
    .edit,
    .fav:not(.on) {
      opacity: 0;
    }
    .tile:hover .edit,
    .tile:focus-within .edit,
    .tile:hover .fav,
    .tile:focus-within .fav {
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
