<script lang="ts">
  import { untrack } from 'svelte';
  import AppIcon from './AppIcon.svelte';
  import Field from './Field.svelte';
  import Modal from './Modal.svelte';
  import { api, ApiError, type AppInput, type AppItem, type CatalogEntry, type TestResult } from '../lib/api';
  import { checkOnForCatalogueApps } from '../lib/status.svelte';

  type Props = {
    /** The app being edited, or null to add a new one. */
    app: AppItem | null;
    /** The preset picked in the first step of adding, which fills in the form. */
    preset?: CatalogEntry | null;
    /** Go back to the first step. Only given while adding. */
    onback?: () => void;
    /** Categories the user already has, offered as suggestions. */
    categories: string[];
    onsave: (input: AppInput) => Promise<void>;
    ondelete: (id: number) => Promise<void>;
    /** Where the app sits in its category now, so it can be moved without dragging. */
    onclose: () => void;
  };

  let { app, preset = null, onback, categories, onsave, ondelete, onclose }: Props = $props();

  let modal = $state<Modal>();

  // The fields start from the app being edited, and from blanks when adding.
  // They are copied once: the parent makes a fresh dialog for each app.
  const start = untrack(() => app);
  const picked = untrack(() => preset);
  let name = $state(start?.name ?? picked?.name ?? '');
  let url = $state(start?.url ?? '');
  let category = $state(start?.category ?? picked?.category ?? '');
  let favourite = $state(start?.favourite ?? false);

  // Status check. An app from the list starts with it on or off as decided in one place.
  let checkEnabled = $state(start?.checkEnabled ?? (picked ? checkOnForCatalogueApps : false));
  let checkUrl = $state(start?.checkUrl ?? '');
  let checkStrict = $state(start?.checkStrict ?? false);
  let checkInsecure = $state(start?.checkInsecure ?? false);
  let testing = $state(false);
  let tested = $state<TestResult | null>(null);
  let testError = $state('');

  async function testNow() {
    tested = null;
    testError = '';
    const address = checkUrl.trim() || url.trim();
    if (!address) {
      testError = 'Enter the app\'s address first.';
      return;
    }
    testing = true;
    try {
      tested = await api.testStatus(address, checkStrict, checkInsecure);
    } catch (e) {
      testError = e instanceof ApiError ? e.message : 'Something went wrong.';
    } finally {
      testing = false;
    }
  }

  // The logo this app can show, if it has one to offer: its own, or the preset's.
  // Unticking it goes back to the first letter of the name.
  const offered = start?.icon || picked?.id || '';
  let useIcon = $state(offered !== '');

  // A preset knows the port its app usually uses, which helps with the address.
  const port = picked && picked.port !== 80 ? picked.port : null;

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
    run(() => onsave({ name, url, category, favourite, newTab: true, icon: useIcon ? offered : '', checkEnabled, checkUrl, checkStrict, checkInsecure }));
  }
</script>

<Modal
  bind:this={modal}
  title={app ? 'Edit app' : picked ? `Configure ${picked.name}` : 'Add custom app'}
  {onclose}
>
  <form class="form" onsubmit={submit}>
    {#if !app && onback}
      <p class="step">Step 2 of 2 · Check the details</p>
    {/if}

    {#if error}
      <div class="alert" role="alert">
        <span class="mark" aria-hidden="true">!</span>
        {error}
      </div>
    {/if}

    <Field label="Name" bind:value={name} autocomplete="off" placeholder="Jellyfin" focusFirst />
    <Field
      label="Address"
      bind:value={url}
      autocomplete="off"
      placeholder={picked ? (port ? `hostname:${port}` : 'hostname') : 'jellyfin.home.lan:8096'}
      hint={picked && port ? `${picked.name} usually runs on port ${port}. Use the address you open it at.` : undefined}
    />
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

    {#if offered}
      <label class="check logo-choice">
        <input type="checkbox" bind:checked={useIcon} />
        <AppIcon name={name || '?'} seed={app?.id ?? 0} icon={useIcon ? offered : ''} size={32} />
        Show the app's logo
      </label>
    {/if}
    <label class="check">
      <input type="checkbox" bind:checked={favourite} />
      Pin to favourites
    </label>

    <section class="status" aria-label="Status check">
      <label class="check">
        <input type="checkbox" bind:checked={checkEnabled} />
        Check if this app is online
      </label>
      <p class="hint">Bifrost asks the app for a reply now and then, and shows a badge on its icon. Set how often in Settings.</p>

      {#if checkEnabled}
        <Field
          label="Address to check (optional)"
          bind:value={checkUrl}
          autocomplete="off"
          placeholder={url || 'Same as the address above'}
          hint="Leave empty to check the address above. The check runs on the machine that hosts Bifrost, so an address only that machine can reach works too."
        />
        <fieldset>
          <legend>Counts as online when</legend>
          <label class="choice" class:on={!checkStrict}>
            <input type="radio" name="{uid}-rule" checked={!checkStrict} onchange={() => (checkStrict = false)} />
            <span><strong>It answers at all</strong><br />Even a login page or an error page. Best for most apps.</span>
          </label>
          <label class="choice" class:on={checkStrict}>
            <input type="radio" name="{uid}-rule" checked={checkStrict} onchange={() => (checkStrict = true)} />
            <span><strong>It answers with success</strong><br />Only a normal page (HTTP 200 to 299) counts.</span>
          </label>
        </fieldset>
        <label class="check">
          <input type="checkbox" bind:checked={checkInsecure} />
          Trust a self-signed certificate
        </label>
        <div class="test">
          <button class="button secondary" type="button" disabled={testing} onclick={testNow}>{testing ? 'Testing…' : 'Test now'}</button>
          <div role="status" class="test-result">
            {#if tested}
              <span class="dot" class:up={tested.up} class:down={!tested.up}></span>
              {#if tested.up}
                <span><strong>Online</strong> · answered in {tested.latencyMs} ms (HTTP {tested.code})</span>
              {:else}
                <span><strong>Offline</strong> · {tested.reason}</span>
              {/if}
            {:else if testError}
              <span class="err">{testError}</span>
            {/if}
          </div>
        </div>
      {/if}
    </section>

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
        {#if !app && onback}
          <button class="button secondary" type="button" disabled={busy} onclick={onback}>← Back</button>
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
  .step {
    margin: -8px 0 0;
    font-size: 13px;
    color: var(--muted);
  }
  .logo-choice {
    gap: 12px;
  }
  .status {
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 14px;
    border: 1px solid var(--border);
    border-radius: 12px;
  }
  .status .hint {
    margin: -8px 0 0 28px;
    font-size: 12px;
    color: var(--muted);
  }
  .status fieldset {
    margin: 0;
    padding: 0;
    border: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .status legend {
    padding: 0;
    margin-bottom: 8px;
    font-size: 13px;
    font-weight: 600;
  }
  .choice {
    display: flex;
    gap: 12px;
    align-items: flex-start;
    min-height: 44px;
    padding: 10px 12px;
    border: 1px solid var(--border);
    border-radius: 10px;
    font-size: 12px;
    line-height: 1.4;
    color: var(--muted);
    cursor: pointer;
  }
  .choice strong {
    color: var(--text);
    font-size: 14px;
  }
  .choice.on {
    border-color: var(--accent);
    outline: 1px solid var(--accent);
  }
  .choice input {
    margin: 3px 0 0;
    accent-color: var(--accent);
  }
  .test {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px;
  }
  .test .button {
    height: 44px;
    padding: 0 16px;
    font-size: 14px;
  }
  .test-result {
    display: flex;
    align-items: center;
    gap: 8px;
    min-height: 20px;
    font-size: 13px;
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }
  .dot.up {
    background: var(--ok);
  }
  .dot.down {
    background: var(--bad);
  }
  .err {
    color: var(--bad);
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
