<script lang="ts">
  import { onMount, untrack } from 'svelte';
  import Field from './Field.svelte';
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

  let dialog = $state<HTMLDialogElement>();
  let nameInput = $state<HTMLInputElement>();

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
  const titleId = `${uid}-title`;
  const listId = `${uid}-categories`;

  onMount(() => {
    dialog?.showModal();
    nameInput?.focus();
  });

  async function run(action: () => Promise<void>) {
    error = '';
    busy = true;
    try {
      await action();
      onclose();
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

  // Clicking the dimmed area around the dialog closes it, like Escape does.
  function backdrop(e: MouseEvent) {
    if (e.target === dialog) dialog?.close();
  }
</script>

<dialog bind:this={dialog} aria-labelledby={titleId} onclose={onclose} onclick={backdrop}>
  <form class="sheet" onsubmit={submit}>
    <header>
      <h2 id={titleId}>{app ? 'Edit app' : 'Add app'}</h2>
      <button class="close" type="button" aria-label="Close" onclick={() => dialog?.close()}>
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M6 6l12 12M18 6L6 18" /></svg>
      </button>
    </header>

    {#if error}
      <div class="alert" role="alert">
        <span class="mark" aria-hidden="true">!</span>
        {error}
      </div>
    {/if}

    <div class="fields">
      <Field label="Name" bind:value={name} bind:input={nameInput} autocomplete="off" placeholder="Jellyfin" />
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
    </div>

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
          <button class="button secondary" type="button" disabled={busy} onclick={() => dialog?.close()}>Cancel</button>
          <button class="button" type="submit" disabled={busy}>
            {busy ? 'Saving…' : app ? 'Save changes' : 'Add app'}
          </button>
        </div>
      {/if}
    </footer>
  </form>
</dialog>

<style>
  dialog {
    box-sizing: border-box;
    width: min(480px, calc(100vw - 32px));
    max-height: calc(100vh - 32px);
    padding: 0;
    border: 1px solid var(--border);
    border-radius: 18px;
    background: var(--surface);
    color: var(--text);
    overflow: auto;
  }
  dialog::backdrop {
    background: rgb(5 6 15 / 0.6);
  }
  .sheet {
    display: flex;
    flex-direction: column;
    gap: 18px;
    padding: 20px 24px 24px;
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  h2 {
    margin: 0;
    font-size: 18px;
  }
  .close {
    width: 44px;
    height: 44px;
    margin-right: -10px;
    border: 0;
    border-radius: 10px;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    display: grid;
    place-items: center;
  }
  .close:focus-visible {
    outline: 2px solid var(--accent);
  }
  .fields {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
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
  .danger {
    background: var(--bad);
    color: var(--on-accent);
  }
  .danger-text {
    color: var(--bad);
  }

  /* On a phone the dialog slides up from the bottom edge as a sheet. */
  @media (max-width: 600px) {
    dialog {
      width: 100%;
      max-width: 100%;
      max-height: 92vh;
      margin: auto 0 0;
      border-radius: 20px 20px 0 0;
      border-bottom: 0;
    }
    .sheet {
      padding-bottom: max(24px, env(safe-area-inset-bottom));
    }
  }
</style>
