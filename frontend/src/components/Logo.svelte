<script lang="ts">
  import type { LogoKind } from '../lib/api';

  // `decorative` hides the logo from screen readers, for when something next to it already says what it is.
  let { size = 64, kind = 'app', decorative = false }: { size?: number; kind?: LogoKind; decorative?: boolean } = $props();
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 96 96"
  role={decorative ? undefined : 'img'}
  aria-label={decorative ? undefined : 'Bifrost'}
  aria-hidden={decorative ? 'true' : undefined}
>
  {#if kind === 'arches'}
    <g fill="none" stroke="var(--accent)" stroke-width="8" stroke-linecap="round">
      <path d="M12 72A36 36 0 0 1 84 72" />
      <path d="M26 72A22 22 0 0 1 70 72" />
      <path d="M40 72A8 8 0 0 1 56 72" />
    </g>
  {:else if kind === 'spectrum'}
    <g fill="none" stroke-width="6" stroke-linecap="round">
      <path d="M10 74A38 38 0 0 1 86 74" stroke="#e5484d" />
      <path d="M17 74A31 31 0 0 1 79 74" stroke="#f59f00" />
      <path d="M24 74A24 24 0 0 1 72 74" stroke="#2f9e63" />
      <path d="M31 74A17 17 0 0 1 65 74" stroke="#3b9ddd" />
      <path d="M38 74A10 10 0 0 1 58 74" stroke="#7048e8" />
    </g>
  {:else}
    <rect x="4" y="4" width="88" height="88" rx="22" fill="var(--accent)" />
    <g fill="none" stroke="var(--on-accent)" stroke-width="7" stroke-linecap="round">
      <path d="M20 63A28 28 0 0 1 76 63" />
      <path d="M33 63A15 15 0 0 1 63 63" />
    </g>
    <circle cx="48" cy="63" r="4.5" fill="var(--on-accent)" />
  {/if}
</svg>
