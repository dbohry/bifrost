<script lang="ts">
  import { untrack } from 'svelte';
  import Field from './Field.svelte';
  import Modal from './Modal.svelte';
  import { ApiError, type AppInput, type AppItem } from '../lib/api';

  type Props = {
    /** The app being edited, or null to add a new one. */
    app: AppItem | null;
    /** Categories the user already has, offered as suggestions. */
    categories: string[];
    onsave: (input: AppInput) => Promise<void>;
    ondelete: (id: number) => Promise<void>;
    onclose: () => void;
  };

  let { app, categories, onsave, ondelete, onclose }: Props = $props();

  let modal = $state<Modal>();

  // The fields start from the app being edited, and from blanks when adding.
  // They are copied once: the parent makes a fresh dialog for each app.
  const start = untrack(() => app);
  let name = $state(start?.name ?? '');
  let url = $state(start?.url ?? '');
  let category = $state(start?.category ?? '');
  let favourite = $state(start?.favourite ?? false);
  let newTab = $state(start?.newTab ?? true);

  let error = $state('');
  let busy = $state(false);
  let confirmingDelete = $state(false);

  const uid = $props.id();
  const listId = `${uid}-categories`;

  async function run(action: () => Promise<void>) {
    error = '';
    busy = true;
    try {
      await action();
      modal?.close();
    } catch (e) {
      error = e instanceof ApiError ? e.message : 'Something went wrong.';
      confirmingDelete = false;
    } finally {
      busy = false;
    }
  }

  function submit(e: SubmitEvent) {
    e.preventDefault();
    run(() => onsave({ name, url, category, favourite, newTab }));
  }
</script>

<Modal bind:this={modal} title={app ? 'Edit app' : 'Add app'} {onclose}>
  <form class="form" onsubmit={submit}>
    {#if error}
      <div class="alert" role="alert">
        <span class="mark" aria-hidden="true">!</span>
        {error}
      </div>
    {/if}

    <Field label="Name" bind:value={name} autocomplete="off" placeholder="Jellyfin" focusFirst />
    <Field label="Address" bind:value={url} autocomplete="off" placeholder="jellyfin.home.lan:8096" />
    <Field
      label="Category"
      bind:value={category}
      autocomplete="off"
      list={listId}
      placeholder="Media"
      hint="Type a new name to create a category. Leave empty for Other."
    />
    <datalist id={listId}>
      {#each categories as c (c)}<option value={c}></option>{/each}
    </datalist>

    <label class="check">
      <input type="checkbox" bind:checked={favourite} />
      Pin to favourites
    </label>
    <label class="check">
      <input type="checkbox" bind:checked={newTab} />
      Open in a new tab
    </label>

    <footer>
      {#if app && confirmingDelete}
        <p class="confirm">Delete {app.name}? This cannot be undone.</p>
        <div class="buttons">
          <button class="button secondary" type="button" disabled={busy} onclick={() => (confirmingDelete = false)}>Keep it</button>
          <button class="button danger" type="button" disabled={busy} onclick={() => run(() => ondelete(app.id))}>Delete</button>
        </div>
      {:else}
        {#if app}
          <button class="button secondary danger-text" type="button" disabled={busy} onclick={() => (confirmingDelete = true)}>Delete</button>
        {/if}
        <div class="buttons">
          <button class="button secondary" type="button" disabled={busy} onclick={() => modal?.close()}>Cancel</button>
          <button class="button" type="submit" disabled={busy}>
            {busy ? 'Saving…' : app ? 'Save changes' : 'Add app'}
          </button>
        </div>
      {/if}
    </footer>
  </form>
</Modal>

<style>
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
  footer {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }
  .buttons {
    display: flex;
    gap: 8px;
    margin-left: auto;
  }
  .confirm {
    margin: 0;
    flex: 1 1 200px;
    font-size: 14px;
  }
</style>
