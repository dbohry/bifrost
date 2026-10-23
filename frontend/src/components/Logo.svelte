<script lang="ts">
  let { size = 64 }: { size?: number } = $props();
</script>

<svg width={size} height={size} viewBox="0 0 96 96" role="img" aria-label="Bifrost">
  <rect x="4" y="4" width="88" height="88" rx="22" fill="var(--accent)" />
  <g fill="none" stroke="var(--on-accent)" stroke-width="7" stroke-linecap="round">
    <path d="M20 63A28 28 0 0 1 76 63" />
    <path d="M33 63A15 15 0 0 1 63 63" />
  </g>
  <circle cx="48" cy="63" r="4.5" fill="var(--on-accent)" />
</svg>
