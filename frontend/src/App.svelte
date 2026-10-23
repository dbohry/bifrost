<script lang="ts">
  import BottomNav from './components/BottomNav.svelte';
  import { appearance } from './lib/appearance.svelte';
  import { arrowNav } from './lib/keynav';
  import { router } from './lib/router.svelte';
  import { session } from './lib/session.svelte';
  import { ui } from './lib/ui.svelte';
  import ChangePassword from './screens/ChangePassword.svelte';
  import Home from './screens/Home.svelte';
  import Login from './screens/Login.svelte';
  import Settings from './screens/Settings.svelte';
  import Setup from './screens/Setup.svelte';

  let failed = $state(false);

  $effect(() => {
    session.init().catch(() => (failed = true));
  });

  // The look follows the user: theirs once they are signed in, the plain default before that.
  $effect(() => {
    if (session.phase === 'authed') appearance.load();
    else if (session.phase !== 'loading') appearance.reset();
  });

  // The screen follows who is using the app. The address bar follows the screen.
  const screen = $derived.by(() => {
    switch (session.phase) {
      case 'setup':
        return 'setup';
      case 'anonymous':
        return 'login';
      case 'authed':
        return session.user?.mustChangePassword ? 'change-password' : 'home';
      default:
        return 'loading';
    }
  });

  const paths = { setup: '/setup', login: '/login', 'change-password': '/change-password' };
  const inSettings = $derived(router.path === '/settings' || router.path.startsWith('/settings/'));

  $effect(() => {
    if (screen === 'loading') return;
    if (screen === 'home') {
      // Signed in: the dashboard at /, settings under /settings. Anything else goes to /.
      if (router.path !== '/' && !inSettings) router.navigate('/', { replace: true });
    } else {
      router.navigate(paths[screen], { replace: true });
    }
  });
</script>

<svelte:window onkeydown={arrowNav} />

<!-- Said aloud by screen readers when something changes without a visible message. -->
<p class="sr-only" role="status" aria-live="polite">{ui.announcement}</p>

{#if failed}
  <p class="status" role="alert">Cannot reach the server.</p>
{:else if screen === 'setup'}
  <Setup />
{:else if screen === 'login'}
  <Login />
{:else if screen === 'change-password'}
  <ChangePassword />
{:else if screen === 'home'}
  <div class="shell">
    {#if inSettings}
      <Settings />
    {:else}
      <Home />
    {/if}
  </div>
  <BottomNav />
{:else}
  <p class="status" role="status">Loading…</p>
{/if}

<style>
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
  /* On a phone held upright the bottom bar sits over the page, so leave room for it. */
  @media (max-width: 599px) and (min-height: 480px) {
    .shell {
      padding-bottom: calc(68px + var(--safe-bottom));
    }
  }
  .status {
    min-height: 100vh;
    display: grid;
    place-items: center;
    margin: 0;
    color: var(--muted);
    font-size: 14px;
  }
</style>
