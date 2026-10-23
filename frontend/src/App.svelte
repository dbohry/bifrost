<script lang="ts">
  type Health = { status: string; version?: string };

  let health = $state<Health | null>(null);
  let failed = $state(false);

  $effect(() => {
    fetch('/healthz')
      .then((r) => (r.ok ? (r.json() as Promise<Health>) : Promise.reject(new Error(r.statusText))))
      .then((h) => (health = h))
      .catch(() => (failed = true));
  });
</script>

<main>
  <svg width="72" height="72" viewBox="0 0 96 96" role="img" aria-label="Bifrost">
    <rect x="4" y="4" width="88" height="88" rx="22" fill="var(--accent)" />
    <g fill="none" stroke="var(--on-accent)" stroke-width="7" stroke-linecap="round">
      <path d="M20 63A28 28 0 0 1 76 63" />
      <path d="M33 63A15 15 0 0 1 63 63" />
    </g>
    <circle cx="48" cy="63" r="4.5" fill="var(--on-accent)" />
  </svg>
  <h1>Bifrost</h1>
  <p class="status" role="status">
    {#if failed}
      <span class="dot bad"></span> Cannot reach the server
    {:else if health}
      <span class="dot ok"></span> Server {health.status}, version {health.version}
    {:else}
      Checking the server…
    {/if}
  </p>
</main>

<style>
  main {
    min-height: 100vh;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 32px 16px;
  }
  h1 {
    margin: 8px 0 0;
    font-size: 28px;
  }
  .status {
    margin: 0;
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    color: var(--muted);
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
  .ok {
    background: var(--ok);
  }
  .bad {
    background: var(--bad);
  }
</style>
