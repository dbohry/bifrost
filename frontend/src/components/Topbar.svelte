<script lang="ts">
  import type { Snippet } from 'svelte';
  import { appearance } from '../lib/appearance.svelte';
  import { link } from '../lib/router.svelte';
  import { session } from '../lib/session.svelte';
  import Logo from './Logo.svelte';

  let {
    current,
    center,
    children,
  }: { current: 'dashboard' | 'settings'; center?: Snippet; children?: Snippet } = $props();
</script>

<header class="bar">
  <a class="brand" href="/" use:link aria-label="Bifrost, back to your apps">
    <Logo size={28} kind={appearance.prefs.logo} /> Bifrost
  </a>
  {#if center}
    <div class="center">{@render center()}</div>
  {/if}
  <span class="actions">
    {@render children?.()}
    <a
      class="icon-button"
      class:active={current === 'settings'}
      href="/settings"
      use:link
      aria-label="Settings"
      aria-current={current === 'settings' ? 'page' : undefined}
    >
      <!-- Sliders: unlike a sun, nobody mistakes it for a theme switch. -->
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
        <path d="M21 4h-7M10 4H3M21 12h-9M8 12H3M21 20h-5M12 20H3M14 2v4M8 10v4M16 18v4" />
      </svg>
    </a>
    <button
      class="button secondary signout"
      aria-label={`Sign out ${session.user?.username ?? ''}`.trim()}
      onclick={() => session.logout()}
    >
      <svg class="out-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M9 4H5v16h4" />
        <path d="M16 8l4 4-4 4M20 12H9" />
      </svg>
      <span class="text" aria-hidden="true">Sign out<span class="who">&nbsp;({session.user?.username})</span></span>
    </button>
  </span>
</header>

<style>
  .bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 8px 12px;
    padding: calc(8px + var(--safe-top)) max(16px, var(--safe-right)) 8px max(16px, var(--safe-left));
    background: var(--surface);
    border-bottom: 1px solid var(--border);
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: 44px;
    font-weight: 700;
    font-size: 18px;
    color: inherit;
    text-decoration: none;
  }
  .brand:focus-visible,
  .icon-button:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
    border-radius: 10px;
  }
  /* On a phone the search box gets a row of its own under the brand and buttons. */
  .center {
    order: 3;
    flex: 1 1 100%;
    min-width: 0;
    padding-bottom: 6px;
  }
  .actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .actions :global(.button) {
    height: 44px;
    padding: 0 16px;
    font-size: 14px;
  }
  .icon-button {
    width: 44px;
    height: 44px;
    border: 1px solid var(--border);
    border-radius: 10px;
    color: var(--text);
    display: grid;
    place-items: center;
  }
  .icon-button.active {
    border-color: var(--accent);
    color: var(--accent);
  }
  .who {
    display: none;
  }
  /* On a phone Sign out is just an icon, so the brand and buttons fit on one row. */
  .signout {
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .signout .text {
    display: none;
  }
  .actions :global(.button.signout) {
    padding: 0;
    width: 44px;
  }
  /* The buttons live in the bottom bar here. BottomNav uses the same query. */
  @media (max-width: 599px) and (min-height: 480px) {
    .actions {
      display: none;
    }
  }
  @media (min-width: 640px) {
    .out-icon {
      display: none;
    }
    .signout .text {
      display: inline;
    }
    .actions :global(.button.signout) {
      padding: 0 16px;
      width: auto;
    }
    .bar {
      padding: calc(12px + var(--safe-top)) max(32px, var(--safe-right)) 12px max(32px, var(--safe-left));
      flex-wrap: nowrap;
    }
    .center {
      order: 0;
      flex: 1 1 auto;
      max-width: 520px;
      padding-bottom: 0;
      margin: 0 auto;
    }
    .who {
      display: inline;
    }
  }
</style>
