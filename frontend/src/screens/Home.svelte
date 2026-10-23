<script lang="ts">
  import AppDialog from '../components/AppDialog.svelte';
  import AppGrid from '../components/AppGrid.svelte';
  import CatalogPicker from '../components/CatalogPicker.svelte';
  import CategoryTabs from '../components/CategoryTabs.svelte';
  import SearchBox from '../components/SearchBox.svelte';
  import Topbar from '../components/Topbar.svelte';
  import type { AppInput, AppItem, CatalogEntry } from '../lib/api';
  import { apps, groupByCategory } from '../lib/apps.svelte';
  import { greeting } from '../lib/tile';
  import { ui } from '../lib/ui.svelte';

  // What is open, if anything. Adding is two steps: pick a preset (or a custom app), then
  // check the details. Editing goes straight to the details.
  type Dialog =
    | { step: 'pick' }
    | { step: 'details'; app: AppItem | null; preset: CatalogEntry | null };
  let dialog = $state<Dialog | null>(null);

  const adding = () => (dialog = { step: 'pick' });
  const editing = (app: AppItem) => (dialog = { step: 'details', app, preset: null });
  // Narrowed once here, so the template and the handlers below stay simple.
  const details = $derived(dialog?.step === 'details' ? dialog : null);

  // Touch screens have no hover, so the edit and star buttons wait behind this switch.
  let editMode = $state(false);

  let query = $state('');
  let searchInput = $state<HTMLInputElement>();
  let tab = $state('All');

  $effect(() => {
    apps.load();
    return () => apps.clear();
  });

  // The bottom bar on a phone asks for search or add. Do it, and clear the request.
  $effect(() => {
    const request = ui.request;
    if (!request) return;
    ui.request = null;
    if (request === 'add') {
      adding();
    } else {
      searchInput?.scrollIntoView({ block: 'nearest' });
      searchInput?.focus();
    }
  });

  const save = (input: AppInput) =>
    details?.app ? apps.update(details.app.id, input) : apps.create(input);

  const count = $derived(apps.items.length);

  // --- search ---
  const q = $derived(query.trim().toLowerCase());
  const searching = $derived(q !== '');
  const matches = $derived(
    searching ? apps.items.filter((a) => [a.name, a.category, a.url].some((f) => f.toLowerCase().includes(q))) : [],
  );

  /** Enter opens the best match. With no match it searches the web instead. */
  function enter() {
    if (!searching) return;
    const first = matches[0];
    if (!first) {
      window.open(`https://duckduckgo.com/?q=${encodeURIComponent(query.trim())}`, '_blank', 'noopener,noreferrer');
    } else if (first.newTab) {
      window.open(first.url, '_blank', 'noopener,noreferrer');
    } else {
      location.assign(first.url);
    }
  }

  // The "/" key jumps to the search box, unless you are typing somewhere or a dialog is open.
  function keydown(e: KeyboardEvent) {
    if (e.key !== '/' || e.ctrlKey || e.metaKey || e.altKey || dialog) return;
    const el = e.target as HTMLElement | null;
    if (el && (el.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName))) return;
    e.preventDefault();
    searchInput?.focus();
  }

  // --- category tabs ---
  // If the chosen category no longer exists (its last app was moved or deleted), show everything.
  const activeTab = $derived(tab !== 'All' && apps.categories.includes(tab) ? tab : 'All');
  const groups = $derived(
    searching
      ? groupByCategory(matches)
      : activeTab === 'All'
        ? apps.groups
        : apps.groups.filter((g) => g.name === activeTab),
  );

  // Dragging changes the order of a whole category, so it is off while a search hides some of its apps.
  const reorderable = $derived(!searching);

  // Where the app being edited sits among the apps of its category.
  const order = $derived.by(() => {
    const app = details?.app;
    if (!app) return null;
    const group = apps.groups.find((g) => g.name === app.category);
    const index = group?.apps.findIndex((a) => a.id === app.id) ?? -1;
    return group && index >= 0 ? { index, count: group.apps.length } : null;
  });
</script>

<svelte:window onkeydown={keydown} />

<Topbar current="dashboard">
  {#snippet center()}
    <SearchBox bind:value={query} bind:input={searchInput} onenter={enter} />
  {/snippet}
  <button class="button add" onclick={() => adding()}>
    <span aria-hidden="true">+</span> Add app
  </button>
</Topbar>

<main>
  <div class="hello">
    <div>
      <h1>{greeting()}</h1>
      {#if apps.loaded}
        <p>{count === 1 ? '1 app' : `${count} apps`}</p>
      {/if}
    </div>
    <div class="controls">
      {#if apps.loaded && apps.categories.length > 1 && !searching}
        <CategoryTabs categories={apps.categories} selected={activeTab} onselect={(name) => (tab = name)} />
      {/if}
      {#if apps.loaded && count > 0}
        <button class="button secondary edit-toggle" type="button" aria-pressed={editMode} onclick={() => (editMode = !editMode)}>
          {editMode ? 'Done' : 'Edit'}
        </button>
      {/if}
    </div>
  </div>

  {#if apps.error}
    <div class="alert" role="alert">
      {apps.error}
      {#if apps.loaded}
        <button class="button secondary" onclick={() => apps.dismissError()}>Dismiss</button>
      {:else}
        <button class="button secondary" onclick={() => apps.load()}>Try again</button>
      {/if}
    </div>
  {/if}

  {#if !apps.loaded}
    {#if !apps.error}<p class="muted" role="status">Loading your apps…</p>{/if}
  {:else if count === 0}
    <section class="empty">
      <h2>No apps yet</h2>
      <p>Add the first app to build your dashboard.</p>
      <button class="button" onclick={() => adding()}>Add your first app</button>
    </section>
  {:else if searching && matches.length === 0}
    <section class="empty" role="status">
      <h2>No app matches “{query.trim()}”</h2>
      <p>Press Enter to search the web instead.</p>
      <button class="button" onclick={enter}>Search the web</button>
    </section>
  {:else}
    {#if !searching && activeTab === 'All' && apps.favourites.length > 0}
      <section aria-labelledby="favourites">
        <h2 id="favourites">Favourites</h2>
        <AppGrid
          apps={apps.favourites}
          editing={editMode}
          onedit={editing}
          onfavourite={(app) => apps.toggleFavourite(app)}
        />
      </section>
    {/if}

    {#each groups as group, i (group.name)}
      <section aria-labelledby="group-{i}">
        <h2 id="group-{i}">{searching ? `${group.name} (${group.apps.length})` : group.name}</h2>
        <AppGrid
          apps={group.apps}
          {reorderable}
          editing={editMode}
          onedit={editing}
          onfavourite={(app) => apps.toggleFavourite(app)}
          onmove={(from, to) => apps.move(group.name, from, to)}
          onadd={!searching && i === groups.length - 1 ? () => adding() : undefined}
        />
      </section>
    {/each}
  {/if}
</main>

{#if dialog?.step === 'pick'}
  <CatalogPicker
    onpick={(preset) => (dialog = { step: 'details', app: null, preset })}
    oncustom={() => (dialog = { step: 'details', app: null, preset: null })}
    onclose={() => (dialog = null)}
  />
{:else if details}
  {#key `${details.app?.id ?? 'new'}-${details.preset?.id ?? ''}`}
    <AppDialog
      app={details.app}
      preset={details.preset}
      onback={details.app ? undefined : adding}
      categories={apps.categories}
      {order}
      onmove={(delta) => details.app && apps.moveBy(details.app, delta)}
      onsave={save}
      ondelete={(id) => apps.remove(id)}
      onclose={() => (dialog = null)}
    />
  {/key}
{/if}

<style>
  main {
    max-width: 1120px;
    margin: 0 auto;
    padding: 24px max(16px, var(--safe-right)) 64px max(16px, var(--safe-left));
    display: flex;
    flex-direction: column;
    gap: 28px;
  }
  @media (min-width: 640px) {
    main {
      padding: 32px max(32px, var(--safe-right)) 64px max(32px, var(--safe-left));
    }
  }
  .hello {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    justify-content: space-between;
    gap: 16px;
  }
  .hello h1 {
    margin: 0;
    font-size: 26px;
  }
  .hello p,
  .muted {
    margin: 4px 0 0;
    font-size: 14px;
    color: var(--muted);
  }
  .controls {
    display: flex;
    align-items: center;
    gap: 8px;
    max-width: 100%;
  }
  /* The tabs scroll sideways on a phone instead of wrapping. */
  .controls :global(.tabs) {
    min-width: 0;
    max-width: 100%;
  }
  .edit-toggle {
    flex-shrink: 0;
    height: 44px;
    padding: 0 18px;
    font-size: 14px;
  }
  .edit-toggle[aria-pressed='true'] {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--on-accent);
  }
  /* Only touch screens need it. With a mouse, the buttons show on hover. */
  @media (hover: hover) {
    .edit-toggle {
      display: none;
    }
  }

  section h2 {
    margin: 0 0 12px;
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 1px;
    text-transform: uppercase;
    color: var(--muted);
  }

  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 48px 16px;
    text-align: center;
    background: var(--surface);
    border: 1px dashed var(--border);
    border-radius: 14px;
  }
  .empty h2 {
    margin: 0;
    font-size: 18px;
    text-transform: none;
    letter-spacing: 0;
    color: var(--text);
  }
  .empty p {
    margin: 0 0 12px;
    color: var(--muted);
  }
  .alert {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px;
    border: 1px solid var(--bad);
    border-radius: 10px;
    font-size: 14px;
  }
  .alert .button {
    height: 44px;
    font-size: 14px;
  }
</style>
