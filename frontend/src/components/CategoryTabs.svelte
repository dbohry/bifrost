<script lang="ts">
  let {
    categories,
    selected,
    onselect,
    label = 'Categories',
    wrap = false,
  }: {
    categories: string[];
    selected: string;
    onselect: (name: string) => void;
    label?: string;
    /** Wrap onto more lines instead of scrolling sideways. */
    wrap?: boolean;
  } = $props();

  const all = 'All';
</script>

<nav class="tabs" class:wrap aria-label={label}>
  {#each [all, ...categories] as name (name)}
    <button
      type="button"
      class="tab"
      class:on={name === selected}
      aria-pressed={name === selected}
      onclick={() => onselect(name)}
    >
      {name}
    </button>
  {/each}
</nav>

<style>
  .tabs {
    display: flex;
    gap: 8px;
    overflow-x: auto;
    padding: 2px;
    margin: -2px;
  }
  .tabs.wrap {
    flex-wrap: wrap;
    overflow-x: visible;
  }
  .tab {
    flex-shrink: 0;
    height: 44px;
    padding: 0 18px;
    border: 1px solid var(--border);
    border-radius: 22px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 14px;
    cursor: pointer;
  }
  .tab:hover {
    color: var(--text);
  }
  .tab.on {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--on-accent);
    font-weight: 600;
  }
  .tab:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
</style>
