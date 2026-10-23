<script lang="ts">
  import Topbar from '../components/Topbar.svelte';
  import { link, router } from '../lib/router.svelte';
  import { session } from '../lib/session.svelte';
  import AccountSection from './AccountSection.svelte';
  import AppearanceSection from './AppearanceSection.svelte';
  import BackupSection from './BackupSection.svelte';
  import StatusSection from './StatusSection.svelte';
  import UsersSection from './UsersSection.svelte';

  const isAdmin = $derived(session.user?.isAdmin ?? false);

  // Only administrators see the Users section. Anyone else who opens its
  // address is sent to their account page, as is any address that is not a section.
  const sections: Record<string, string> = {
    '/settings/appearance': 'appearance',
    '/settings/status': 'status',
    '/settings/backup': 'backup',
  };
  const section = $derived(
    router.path === '/settings/users' ? (isAdmin ? 'users' : 'account') : (sections[router.path] ?? 'account'),
  );

  $effect(() => {
    if (router.path !== '/settings' && section === 'account') router.navigate('/settings', { replace: true });
  });

  // Escape leaves Settings, unless a dialog is open: that one closes first.
  function leave(e: KeyboardEvent) {
    if (e.key === 'Escape' && !e.defaultPrevented && !document.querySelector('dialog[open]')) router.navigate('/');
  }
</script>

<svelte:window onkeydown={leave} />

<Topbar current="settings" />

<div class="layout">
  <nav aria-label="Settings sections">
    <a href="/settings" use:link aria-current={section === 'account' ? 'page' : undefined}>Account</a>
    <a href="/settings/appearance" use:link aria-current={section === 'appearance' ? 'page' : undefined}>Appearance</a>
    <a href="/settings/status" use:link aria-current={section === 'status' ? 'page' : undefined}>Status checks</a>
    <a href="/settings/backup" use:link aria-current={section === 'backup' ? 'page' : undefined}>Backup</a>
    {#if isAdmin}
      <a href="/settings/users" use:link aria-current={section === 'users' ? 'page' : undefined}>Users</a>
    {/if}
  </nav>

  <main>
    {#if section === 'users'}
      <UsersSection />
    {:else if section === 'appearance'}
      <AppearanceSection />
    {:else if section === 'status'}
      <StatusSection />
    {:else if section === 'backup'}
      <BackupSection />
    {:else}
      <AccountSection />
    {/if}
  </main>
</div>

<style>
  .layout {
    max-width: 1120px;
    margin: 0 auto;
    padding: 24px max(16px, var(--safe-right)) 64px max(16px, var(--safe-left));
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  nav {
    display: flex;
    gap: 6px;
    overflow-x: auto;
  }
  nav a {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    height: 44px;
    padding: 0 16px;
    border-radius: 10px;
    color: var(--text);
    text-decoration: none;
    font-size: 14px;
  }
  nav a:hover {
    background: var(--icon-bg);
  }
  nav a[aria-current='page'] {
    background: var(--icon-bg);
    font-weight: 600;
  }
  nav a:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  main {
    flex: 1;
    min-width: 0;
  }
  @media (min-width: 760px) {
    .layout {
      padding: 32px max(32px, var(--safe-right)) 64px max(32px, var(--safe-left));
      flex-direction: row;
      align-items: flex-start;
      gap: 32px;
    }
    nav {
      flex: 0 0 200px;
      flex-direction: column;
      overflow: visible;
    }
  }
</style>
