<script lang="ts">
  import type { Snippet } from 'svelte';
  import { link } from '../lib/router.svelte';
  import { session } from '../lib/session.svelte';
  import Logo from './Logo.svelte';

  let {
    current,
    children,
  }: { current: 'dashboard' | 'settings'; children?: Snippet } = $props();
</script>

<header class="bar">
  <a class="brand" href="/" use:link aria-label="Bifrost, back to your apps">
    <Logo size={28} /> Bifrost
  </a>
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
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
        <circle cx="12" cy="12" r="3" />
        <path d="M12 2v3M12 19v3M2 12h3M19 12h3M5 5l2 2M17 17l2 2M19 5l-2 2M7 17l-2 2" />
      </svg>
    </a>
    <button class="button secondary" onclick={() => session.logout()}>
      Sign out<span class="who">&nbsp;({session.user?.username})</span>
    </button>
  </span>
</header>

<style>
  .bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 8px 16px;
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
  @media (min-width: 640px) {
    .bar {
      padding: 12px 32px;
    }
    .who {
      display: inline;
    }
  }
</style>
