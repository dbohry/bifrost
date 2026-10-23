<script lang="ts">
  import AppDialog from '../components/AppDialog.svelte';
  import AppTile from '../components/AppTile.svelte';
  import Logo from '../components/Logo.svelte';
  import type { AppInput, AppItem } from '../lib/api';
  import { apps } from '../lib/apps.svelte';
  import { session } from '../lib/session.svelte';
  import { greeting } from '../lib/tile';

  // null: closed. { app: null }: adding. { app }: editing that app.
  let dialog = $state<{ app: AppItem | null } | null>(null);

  $effect(() => {
    apps.load();
    return () => apps.clear();
  });

  const save = (input: AppInput) =>
    dialog?.app ? apps.update(dialog.app.id, input) : apps.create(input);

  const count = $derived(apps.items.length);
</script>

<header class="bar">
  <span class="brand"><Logo size={28} /> Bifrost</span>
  <span class="actions">
    <button class="button add" onclick={() => (dialog = { app: null })}>
      <span aria-hidden="true">+</span> Add app
    </button>
    <button class="button secondary" onclick={() => session.logout()}>
      Sign out<span class="who">&nbsp;({session.user?.username})</span>
    </button>
  </span>
</header>

<main>
  <div class="hello">
    <h1>{greeting()}</h1>
    {#if apps.loaded}
      <p>{count === 1 ? '1 app' : `${count} apps`}</p>
    {/if}
  </div>

  {#if apps.error}
    <div class="alert" role="alert">
      {apps.error}
      <button class="button secondary" onclick={() => apps.load()}>Try again</button>
    </div>
  {:else if !apps.loaded}
    <p class="muted" role="status">Loading your apps…</p>
  {:else if count === 0}
    <section class="empty">
      <h2>No apps yet</h2>
      <p>Add the first app to build your dashboard.</p>
      <button class="button" onclick={() => (dialog = { app: null })}>Add your first app</button>
    </section>
  {:else}
    {#if apps.favourites.length > 0}
      <section aria-labelledby="favourites">
        <h2 id="favourites">Favourites</h2>
        <div class="grid">
          {#each apps.favourites as app (app.id)}
            <AppTile {app} onedit={() => (dialog = { app })} onfavourite={() => apps.toggleFavourite(app)} />
          {/each}
        </div>
      </section>
    {/if}

    {#each apps.groups as group, i (group.name)}
      <section aria-labelledby="group-{i}">
        <h2 id="group-{i}">{group.name}</h2>
        <div class="grid">
          {#each group.apps as app (app.id)}
            <AppTile {app} onedit={() => (dialog = { app })} onfavourite={() => apps.toggleFavourite(app)} />
          {/each}
          {#if i === apps.groups.length - 1}
            <button class="add-tile" onclick={() => (dialog = { app: null })}>
              <span class="plus" aria-hidden="true">+</span>
              Add app
            </button>
          {/if}
        </div>
      </section>
    {/each}
  {/if}
</main>

{#if dialog}
  {#key dialog.app?.id ?? 'new'}
    <AppDialog
      app={dialog.app}
      categories={apps.categories}
      onsave={save}
      ondelete={(id) => apps.remove(id)}
      onclose={() => (dialog = null)}
    />
  {/key}
{/if}

<style>
  .bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 8px 16px;
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
  .actions {
    display: flex;
    gap: 8px;
  }
  .actions .button {
    height: 44px;
    padding: 0 16px;
    font-size: 14px;
  }
  .who {
    display: none;
  }
  @media (min-width: 640px) {
    .bar {
      padding: 12px 32px;
    }
    .who {
      display: inline;
    }
  }

  main {
    max-width: 1120px;
    margin: 0 auto;
    padding: 24px 16px 64px;
    display: flex;
    flex-direction: column;
    gap: 28px;
  }
  @media (min-width: 640px) {
    main {
      padding: 32px 32px 64px;
    }
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

  section h2 {
    margin: 0 0 12px;
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 1px;
    text-transform: uppercase;
    color: var(--muted);
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 16px;
  }
  @media (max-width: 480px) {
    .grid {
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 12px;
    }
  }

  .add-tile {
    min-height: 120px;
    border: 2px dashed var(--border);
    border-radius: 14px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 14px;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
  }
  .add-tile:hover {
    border-color: var(--accent);
    color: var(--text);
  }
  .add-tile:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .plus {
    font-size: 26px;
    line-height: 1;
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
</style>
