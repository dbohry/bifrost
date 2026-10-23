<script lang="ts">
  import { untrack } from 'svelte';
  import AppIcon from './AppIcon.svelte';
  import CategoryTabs from './CategoryTabs.svelte';
  import Modal from './Modal.svelte';
  import type { CatalogEntry } from '../lib/api';
  import { catalog } from '../lib/catalog.svelte';

  type Props = {
    /** A preset was picked. */
    onpick: (entry: CatalogEntry) => void;
    /** "Custom app": go on without a preset. */
    oncustom: () => void;
    onclose: () => void;
  };

  let { onpick, oncustom, onclose }: Props = $props();

  let query = $state('');
  let tab = $state('All');

  untrack(() => catalog.load());

  const q = $derived(query.trim().toLowerCase());
  const shown = $derived(
    catalog.entries.filter(
      (e) => (tab === 'All' || e.category === tab) && (q === '' || `${e.name} ${e.category}`.toLowerCase().includes(q)),
    ),
  );
</script>

<Modal title="Add app" {onclose}>
  <div class="body">
    <p class="step">Step 1 of 2 · Pick an app, or add your own</p>

    <div class="search">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><circle cx="11" cy="11" r="7" /><path d="M20 20l-4-4" /></svg>
      <input
        bind:value={query}
        type="text"
        placeholder="Search {catalog.entries.length || ''} apps"
        aria-label="Search the app list"
        autocomplete="off"
        autocapitalize="off"
        spellcheck="false"
        data-autofocus
      />
    </div>

    {#if catalog.loaded}
      <CategoryTabs wrap label="App categories" categories={catalog.categories} selected={tab} onselect={(name) => (tab = name)} />
    {/if}

    {#if catalog.error}
      <div class="notice" role="alert">
        {catalog.error}
        <button class="button secondary" type="button" onclick={() => catalog.load()}>Try again</button>
      </div>
    {:else if !catalog.loaded}
      <p class="muted" role="status">Loading the app list…</p>
    {:else if shown.length === 0}
      <p class="muted" role="status">No app matches “{query.trim()}”. You can add it as a custom app below.</p>
    {:else}
      <ul class="list">
        {#each shown as entry (entry.id)}
          <li>
            <button class="preset" type="button" onclick={() => onpick(entry)}>
              <AppIcon name={entry.name} seed={0} icon={entry.id} size={40} />
              <span class="text">
                <span class="name">{entry.name}</span>
                <span class="category">{entry.category}</span>
              </span>
            </button>
          </li>
        {/each}
      </ul>
    {/if}

    <div class="footer">
      <span>Can't find it? Add any address as a custom app.</span>
      <button class="button secondary" type="button" onclick={oncustom}>Custom app →</button>
    </div>
  </div>
</Modal>

<style>
  .body {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .step {
    margin: -8px 0 0;
    font-size: 13px;
    color: var(--muted);
  }
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
  .search input {
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

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
    gap: 8px;
    max-height: min(360px, 45vh);
    overflow-y: auto;
    padding-right: 2px;
  }
  .preset {
    width: 100%;
    min-height: 56px;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px 12px;
    border: 1px solid var(--border);
    border-radius: 12px;
    background: var(--surface);
    color: var(--text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }
  .preset:hover {
    border-color: var(--accent);
  }
  .preset:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .text {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .name {
    font-size: 14px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .category {
    font-size: 12px;
    color: var(--muted);
  }
  .muted {
    margin: 0;
    padding: 20px 0;
    font-size: 14px;
    color: var(--muted);
  }
  .notice {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px;
    border: 1px solid var(--bad);
    border-radius: 10px;
    font-size: 14px;
  }
  .footer {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding-top: 14px;
    border-top: 1px solid var(--border);
    font-size: 13px;
    color: var(--muted);
  }
  .footer .button,
  .notice .button {
    height: 44px;
    font-size: 14px;
  }

  /* On a phone the presets are a single column of rows. */
  @media (max-width: 600px) {
    .list {
      grid-template-columns: 1fr;
      max-height: 38vh;
    }
  }
</style>
