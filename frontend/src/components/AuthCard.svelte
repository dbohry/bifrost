<script lang="ts">
  import type { Snippet } from 'svelte';
  import Logo from './Logo.svelte';
  import { session } from '../lib/session.svelte';

  let {
    title,
    subtitle,
    error = '',
    children,
  }: { title: string; subtitle?: string; error?: string; children: Snippet } = $props();
</script>

<div class="page">
  <main>
    <header>
      <Logo />
      <h1>{title}</h1>
      {#if subtitle}<p>{subtitle}</p>{/if}
    </header>

    {#if error}
      <div class="alert" role="alert">
        <span class="mark" aria-hidden="true">!</span>
        {error}
      </div>
    {/if}

    {@render children()}
  </main>
  <p class="footer">Bifrost{session.version ? ` v${session.version}` : ''}</p>
</div>

<style>
  .page {
    min-height: 100vh;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 20px;
    padding: 32px 16px;
  }
  main {
    width: 100%;
    max-width: 400px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 24px;
    padding: 32px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 18px;
  }
  header {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    text-align: center;
  }
  h1 {
    margin: 0;
    font-size: 22px;
  }
  header p {
    margin: 6px 0 0;
    font-size: 14px;
    line-height: 1.5;
    color: var(--muted);
  }
  .alert {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    border: 1px solid var(--bad);
    border-radius: 10px;
    font-size: 13px;
    line-height: 1.4;
  }
  .mark {
    flex-shrink: 0;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--bad);
    color: var(--on-accent);
    display: grid;
    place-items: center;
    font-size: 12px;
    font-weight: 700;
  }
  .footer {
    margin: 0;
    font-size: 12px;
    color: var(--muted);
  }
</style>
