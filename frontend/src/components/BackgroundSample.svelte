<script lang="ts">
  import type { BackgroundKind } from '../lib/api';

  let { kind }: { kind: BackgroundKind } = $props();
</script>

<!-- A small page with the given backdrop. The backdrop itself comes from theme.css. -->
<div class="sample" data-background={kind} aria-hidden="true">
  <span class="bar"><i></i><b></b></span>
  <span class="tiles"><i></i><i></i><i></i><i></i></span>
</div>

<style>
  .sample {
    width: 100%;
    height: 96px;
    box-sizing: border-box;
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    border: 1px solid var(--border);
    border-radius: 10px;
    overflow: hidden;
    background-color: var(--bg);
    /* theme.css fixes the backdrop to the screen; here it stays in the card. */
    background-attachment: scroll !important;
  }
  .bar {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .bar i {
    width: 10px;
    height: 10px;
    border-radius: 3px;
    background: var(--accent);
  }
  .bar b {
    margin-left: auto;
    width: 36%;
    height: 10px;
    border-radius: 4px;
    background: var(--surface);
    border: 1px solid var(--border);
  }
  .tiles {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 6px;
  }
  .tiles i {
    height: 34px;
    border-radius: 6px;
    background: var(--surface);
    border: 1px solid var(--border);
  }
</style>
