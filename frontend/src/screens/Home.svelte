<script lang="ts">
  import Logo from '../components/Logo.svelte';
  import { session } from '../lib/session.svelte';
</script>

<header>
  <span class="brand"><Logo size={28} /> Bifrost</span>
  <span class="who">
    {session.user?.username}
    <button class="button secondary" onclick={() => session.logout()}>Sign out</button>
  </span>
</header>

<main>
  <h1>Welcome, {session.user?.username}</h1>
  <p>Your dashboard will appear here.</p>
</main>

<style>
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 12px 32px;
    background: var(--surface);
    border-bottom: 1px solid var(--border);
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    font-weight: 700;
    font-size: 18px;
  }
  .who {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 14px;
    color: var(--muted);
  }
  .who .button {
    height: 44px;
    font-size: 14px;
  }
  main {
    max-width: 1120px;
    margin: 0 auto;
    padding: 32px;
  }
  h1 {
    margin: 0;
    font-size: 26px;
  }
  p {
    margin: 4px 0 0;
    color: var(--muted);
  }
</style>
