<script lang="ts">
  import { link, router } from '../lib/router.svelte';
  import { ui } from '../lib/ui.svelte';

  const onApps = $derived(router.path === '/');
  const onSettings = $derived(router.path === '/settings' || router.path.startsWith('/settings/'));

  function ask(what: 'search' | 'add') {
    // The dashboard does the work, so first make sure it is the one on screen.
    router.navigate('/');
    ui.request = what;
  }
</script>

<!-- Phones only. The thumb reaches the bottom edge; the top edge is out of reach. -->
<nav class="bottom" aria-label="Main">
  <a href="/" use:link aria-current={onApps ? 'page' : undefined}>
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <rect x="4" y="4" width="6" height="6" rx="1.5" /><rect x="14" y="4" width="6" height="6" rx="1.5" />
      <rect x="4" y="14" width="6" height="6" rx="1.5" /><rect x="14" y="14" width="6" height="6" rx="1.5" />
    </svg>
    Apps
  </a>
  <button type="button" onclick={() => ask('search')}>
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
      <circle cx="11" cy="11" r="7" /><path d="M20 20l-4-4" />
    </svg>
    Search
  </button>
  <button type="button" onclick={() => ask('add')}>
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
      <circle cx="12" cy="12" r="9" /><path d="M12 8v8M8 12h8" />
    </svg>
    Add
  </button>
  <a href="/settings" use:link aria-current={onSettings ? 'page' : undefined}>
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
      <path d="M21 4h-7M10 4H3M21 12h-9M8 12H3M21 20h-5M12 20H3M14 2v4M8 10v4M16 18v4" />
    </svg>
    Settings
  </a>
</nav>

<style>
  .bottom {
    position: fixed;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 5;
    display: none;
    padding: 6px calc(8px + var(--safe-right)) calc(6px + var(--safe-bottom)) calc(8px + var(--safe-left));
    background: var(--surface);
    border-top: 1px solid var(--border);
  }
  /*
   * A narrow, tall screen: a phone held upright. A phone on its side is wide
   * enough for the top bar, and too short to give a strip of it away. Topbar
   * uses the same query to know when to hand its buttons over.
   */
  @media (max-width: 599px) and (min-height: 480px) {
    .bottom {
      display: flex;
    }
  }
  a,
  button {
    flex: 1;
    min-height: 52px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 3px;
    border: 0;
    border-radius: 10px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 11px;
    text-decoration: none;
    cursor: pointer;
  }
  /* The label stays in the text colour: small accent text can miss 4.5:1 on some themes. */
  a[aria-current='page'] {
    color: var(--text);
    font-weight: 700;
  }
  a[aria-current='page'] svg {
    color: var(--accent);
  }
  a:focus-visible,
  button:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }
</style>
