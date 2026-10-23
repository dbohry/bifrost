<script lang="ts">
  type Props = {
    value: string;
    input?: HTMLInputElement;
    /** Enter was pressed. */
    onenter: () => void;
  };

  let { value = $bindable(''), input = $bindable(), onenter }: Props = $props();

  function keydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      onenter();
    } else if (e.key === 'Escape' && value !== '') {
      e.preventDefault();
      value = '';
    }
  }
</script>

<div class="search" role="search">
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
    <circle cx="11" cy="11" r="7" />
    <path d="M20 20l-4-4" />
  </svg>
  <input
    bind:this={input}
    bind:value
    type="text"
    placeholder="Search apps or the web"
    aria-label="Search apps or the web"
    aria-keyshortcuts="/"
    autocomplete="off"
    autocapitalize="off"
    spellcheck="false"
    enterkeyhint="search"
    onkeydown={keydown}
  />
  <kbd aria-hidden="true">/</kbd>
</div>

<style>
  .search {
    display: flex;
    align-items: center;
    gap: 10px;
    height: 44px;
    box-sizing: border-box;
    padding: 0 14px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--bg);
    color: var(--muted);
  }
  .search:focus-within {
    border-color: var(--accent);
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }
  input {
    flex: 1;
    min-width: 0;
    height: 100%;
    border: 0;
    background: transparent;
    color: var(--text);
    font: inherit;
    font-size: 14px;
    outline: none;
  }
  kbd {
    padding: 0 6px;
    border: 1px solid var(--border);
    border-radius: 5px;
    font: inherit;
    font-size: 12px;
  }
  /* The shortcut hint is for keyboards, so it is hidden on touch screens. */
  @media (hover: none) {
    kbd {
      display: none;
    }
  }
</style>
