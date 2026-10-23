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
    /** Categories the user already has, offered as choices. */
    categories: string[];
    onsave: (input: AppInput) => Promise<void>;
    ondelete: (id: number) => Promise<void>;
    onclose: () => void;
  };

  let { app, preset = null, onback, categories, onsave, ondelete, onclose }: Props = $props();

  let modal = $state<Modal>();

  // The fields start from the app being edited, and from blanks when adding.
  // They are copied once: the parent makes a fresh dialog for each app.
  const start = untrack(() => app);
  const picked = untrack(() => preset);
  const existing = untrack(() => categories);

  // The server keeps the whole address. It is split here so the scheme is a
  // switch and the address box holds only what people recognise.
  const parts = /^(https?):\/\/(.*)$/i.exec(start?.url ?? '');
  let scheme = $state<'http' | 'https'>(parts?.[1].toLowerCase() === 'https' ? 'https' : 'http');
  let host = $state(parts ? parts[2] : (start?.url ?? ''));
  let name = $state(start?.name ?? picked?.name ?? '');
  let category = $state(start?.category ?? picked?.category ?? '');
  let favourite = $state(start?.favourite ?? false);
  let newTab = $state(start?.newTab ?? true);

  // Status check. An app from the list starts with it on or off as decided in one place.
  let checkEnabled = $state(start?.checkEnabled ?? (picked ? checkOnForCatalogueApps : false));
  let checkUrl = $state(start?.checkUrl ?? '');
  let checkStrict = $state(start?.checkStrict ?? false);
  let checkInsecure = $state(start?.checkInsecure ?? false);
  let testing = $state(false);
  let tested = $state<TestResult | null>(null);
  let testError = $state('');

  // The logo this app can show, if it has one to offer: its own, or the preset's.
  // Turning it off goes back to the first letter of the name.
  const offered = start?.icon || picked?.id || '';
  let useIcon = $state(offered !== '');

  // A preset knows the port its app usually uses, which helps with the address.
  const port = picked && picked.port !== 80 ? picked.port : null;

  // Choices are the user's categories, plus the preset's if it is a new one.
  const choices = $derived(category && !existing.includes(category) ? [...existing, category] : existing);
  let creating = $state(false);
  let draft = $state('');

  let error = $state('');
  let busy = $state(false);
  let confirmingDelete = $state(false);
  let touched = $state(false);
  let tab = $state<'details' | 'status'>('details');

  const uid = $props.id();

  const trimmed = $derived(host.trim());
  const hostInvalid = $derived(touched && trimmed === '');
  const nameInvalid = $derived(touched && name.trim() === '');
  const showPortHint = $derived(port !== null && trimmed !== '' && !/:\d+/.test(trimmed));

  async function testNow() {
    tested = null;
    testError = '';
    const address = checkUrl.trim() || (trimmed ? `${scheme}://${trimmed}` : '');
    if (!address) {
      testError = "Enter the app's address first.";
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
    touched = true;
    if (name.trim() === '' || trimmed === '') {
      tab = 'details';
      return;
    }
    run(() =>
      onsave({
        name,
        url: `${scheme}://${trimmed}`,
        category,
        favourite,
        newTab,
        icon: useIcon ? offered : '',
        checkEnabled,
        checkUrl,
        checkStrict,
        checkInsecure,
      }),
    );
  }

  // Pasting or typing a full address picks the scheme and keeps the rest.
  function onhost() {
    const m = /^(https?):\/\/(.*)$/i.exec(host);
    if (m) {
      scheme = m[1].toLowerCase() === 'https' ? 'https' : 'http';
      host = m[2];
    }
  }

  function pickCategory(c: string) {
    category = category === c ? '' : c;
  }

  function finishCategory() {
    const v = draft.trim();
    if (v) category = v;
    draft = '';
    creating = false;
  }

  function newCategoryKey(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      finishCategory();
    } else if (e.key === 'Escape') {
      e.stopPropagation();
      draft = '';
      creating = false;
    }
  }

  function focus(node: HTMLElement) {
    node.focus();
  }
</script>

<Modal
  bind:this={modal}
  title={app ? 'Edit app' : picked ? `Configure ${picked.name}` : 'Add custom app'}
  {onclose}
>
  <form class="form" onsubmit={submit} novalidate>
    {#if !app && onback}
      <p class="step">
        <span class="bars" aria-hidden="true"><i></i><i></i></span>
        Step 2 of 2 · Check the details
      </p>
    {/if}

    {#if error}
      <div class="alert" role="alert">
        <span class="mark" aria-hidden="true">!</span>
        {error}
      </div>
    {/if}

    <div class="tabs" role="tablist" aria-label="App settings">
      <button type="button" role="tab" id="{uid}-tab-details" aria-selected={tab === 'details'} aria-controls="{uid}-details" onclick={() => (tab = 'details')}>Details</button>
      <button type="button" role="tab" id="{uid}-tab-status" aria-selected={tab === 'status'} aria-controls="{uid}-status" onclick={() => (tab = 'status')}>
        Status check
        <span class="badge" class:off={!checkEnabled}>{checkEnabled ? 'On' : 'Off'}</span>
      </button>
    </div>

    <div class="panel" id="{uid}-details" role="tabpanel" aria-labelledby="{uid}-tab-details" hidden={tab !== 'details'}>
    <div class="namerow">
      <Field label="Name" bind:value={name} invalid={nameInvalid} autocomplete="off" placeholder="Jellyfin" focusFirst />
      <AppIcon name={name || '?'} seed={app?.id ?? 0} icon={useIcon ? offered : ''} size={48} />
    </div>

    <div class="field">
      <label for="{uid}-host">Address</label>
      <div class="box" class:invalid={hostInvalid}>
        <div class="scheme" role="group" aria-label="Protocol">
          <button type="button" aria-pressed={scheme === 'http'} onclick={() => (scheme = 'http')}>http</button>
          <button type="button" aria-pressed={scheme === 'https'} onclick={() => (scheme = 'https')}>https</button>
        </div>
        <input
          id="{uid}-host"
          bind:value={host}
          oninput={onhost}
          onblur={() => (touched = true)}
          inputmode="url"
          autocomplete="off"
          autocapitalize="off"
          spellcheck="false"
          placeholder={picked ? (port ? `hostname:${port}` : 'hostname') : 'jellyfin.home.lan:8096'}
          aria-invalid={hostInvalid}
          aria-describedby="{uid}-host-hint"
        />
      </div>
      <span class="hint" class:bad={hostInvalid} id="{uid}-host-hint">
        {#if hostInvalid}
          Enter the address you open this app at, like jellyfin.home.lan:8096.
        {:else if picked && port}
          {picked.name} usually runs on port {port}. Use the address you open it at.
        {/if}
        {#if showPortHint}
          <button type="button" class="port" onclick={() => (host = `${trimmed.replace(/\/$/, '')}:${port}`)}>Add :{port}</button>
        {/if}
      </span>
    </div>

    <div class="field" role="group" aria-labelledby="{uid}-cat">
      <span class="label" id="{uid}-cat">Category <small>Optional</small></span>
      <div class="cats">
        {#each choices as c (c)}
          <button type="button" class="cat" aria-pressed={category === c} onclick={() => pickCategory(c)}>{c}</button>
        {/each}
        {#if creating}
          <input
            class="newcat"
            bind:value={draft}
            use:focus
            maxlength="40"
            aria-label="New category name"
            placeholder="Category name"
            onkeydown={newCategoryKey}
            onblur={finishCategory}
          />
        {:else}
          <button type="button" class="cat new" onclick={() => (creating = true)}>+ New</button>
        {/if}
      </div>
    </div>

    <div class="opts" role="group" aria-label="Appearance and behaviour">
      {#if offered}
        <button type="button" class="opt" role="switch" aria-checked={useIcon} onclick={() => (useIcon = !useIcon)}>
          <span class="text"><b>Show the app's logo</b></span>
          <span class="sw" aria-hidden="true"></span>
        </button>
      {/if}
      <button type="button" class="opt" role="switch" aria-checked={favourite} onclick={() => (favourite = !favourite)}>
        <span class="text"><b>Pin to favourites</b></span>
        <span class="sw" aria-hidden="true"></span>
      </button>
      <button type="button" class="opt" role="switch" aria-checked={newTab} onclick={() => (newTab = !newTab)}>
        <span class="text"><b>Open in a new tab</b></span>
        <span class="sw" aria-hidden="true"></span>
      </button>
    </div>
    </div>

    <div class="panel" id="{uid}-status" role="tabpanel" aria-labelledby="{uid}-tab-status" hidden={tab !== 'status'}>
      <div class="opts">
        <button type="button" class="opt" role="switch" aria-checked={checkEnabled} onclick={() => (checkEnabled = !checkEnabled)}>
          <span class="text"><b>Check if this app is online</b><span>Shows a badge on its icon. Set how often in Settings.</span></span>
          <span class="sw" aria-hidden="true"></span>
        </button>
      </div>

      {#if checkEnabled}
        <Field
          label="Address to check (optional)"
          bind:value={checkUrl}
          autocomplete="off"
          placeholder={trimmed ? `${scheme}://${trimmed}` : 'Same as the address on Details'}
          hint="Empty checks the address on Details. The check runs on the machine that hosts Bifrost, so an address only that machine can reach works too."
        />
        <div class="field" role="group" aria-labelledby="{uid}-rule">
          <span class="label" id="{uid}-rule">Counts as online when</span>
          <div class="seg">
            <button type="button" aria-pressed={!checkStrict} onclick={() => (checkStrict = false)}>Any answer</button>
            <button type="button" aria-pressed={checkStrict} onclick={() => (checkStrict = true)}>Success only</button>
          </div>
          <span class="hint">
            {checkStrict ? 'Only a normal page (HTTP 200 to 299) counts.' : 'Even a login page or an error page counts. Best for most apps.'}
          </span>
        </div>
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
    display: flex;
    align-items: center;
    gap: 10px;
    margin: -8px 0 0;
    font-size: 13px;
    color: var(--muted);
  }
  .bars {
    display: flex;
    gap: 4px;
  }
  .bars i {
    width: 20px;
    height: 4px;
    border-radius: 2px;
    background: var(--accent);
  }

  /* Address and category share the look of Field. */
  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .field > label,
  .label {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: 8px;
    font-size: 13px;
    font-weight: 600;
  }
  .label small {
    font-size: 12px;
    font-weight: 500;
    color: var(--muted);
  }
  .box {
    display: flex;
    align-items: center;
    height: 48px;
    box-sizing: border-box;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--bg);
    overflow: hidden;
  }
  .box:focus-within {
    border-color: var(--accent);
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }
  .box.invalid {
    border-color: var(--bad);
  }
  .box input {
    flex: 1;
    min-width: 0;
    height: 100%;
    box-sizing: border-box;
    padding: 0 12px;
    border: 0;
    background: transparent;
    color: var(--text);
    font: inherit;
    font-size: 15px;
    outline: none;
  }
  .scheme {
    display: flex;
    align-items: center;
    gap: 2px;
    height: 100%;
    padding: 0 4px;
    border-right: 1px solid var(--border);
  }
  .scheme button {
    height: 34px;
    padding: 0 10px;
    border: 0;
    border-radius: 7px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-family: ui-monospace, 'SF Mono', Menlo, Consolas, monospace;
    font-size: 13px;
    cursor: pointer;
  }
  .scheme button[aria-pressed='true'] {
    background: var(--surface);
    color: var(--text);
    box-shadow: 0 0 0 1px var(--border);
  }
  .scheme button:focus-visible {
    outline: 2px solid var(--accent);
  }
  .hint {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px 8px;
    font-size: 12px;
    color: var(--muted);
  }
  .hint.bad {
    color: var(--bad);
  }
  .port {
    height: 24px;
    padding: 0 9px;
    border: 1px dashed var(--accent);
    border-radius: 12px;
    background: transparent;
    color: var(--accent);
    font: inherit;
    font-family: ui-monospace, 'SF Mono', Menlo, Consolas, monospace;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
  }
  .port:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .cats {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .cat,
  .newcat {
    height: 36px;
    box-sizing: border-box;
    border-radius: 18px;
    font: inherit;
    font-size: 13.5px;
  }
  .cat {
    padding: 0 14px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text);
    font-weight: 600;
    cursor: pointer;
  }
  .cat[aria-pressed='true'] {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--on-accent);
  }
  .cat.new {
    border-style: dashed;
    color: var(--muted);
  }
  .cat:focus-visible,
  .newcat:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .newcat {
    width: 160px;
    padding: 0 14px;
    border: 1px solid var(--accent);
    background: var(--bg);
    color: var(--text);
  }

  .opts {
    border: 1px solid var(--border);
    border-radius: 12px;
    overflow: hidden;
  }
  .opt {
    display: flex;
    align-items: center;
    gap: 14px;
    width: 100%;
    min-height: 48px;
    padding: 6px 14px;
    border: 0;
    background: transparent;
    color: var(--text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }
  .opt + .opt {
    border-top: 1px solid var(--border);
  }
  .opt:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }
  .text {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }
  .text b {
    font-size: 14px;
  }
  .text span {
    font-size: 12.5px;
    color: var(--muted);
  }
  .sw {
    flex-shrink: 0;
    position: relative;
    width: 44px;
    height: 26px;
    border-radius: 13px;
    background: var(--border);
    transition: background 0.18s;
  }
  .sw::after {
    content: '';
    position: absolute;
    top: 3px;
    left: 3px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 3px rgb(0 0 0 / 0.3);
    transition: transform 0.18s;
  }
  .opt[aria-checked='true'] .sw {
    background: var(--accent);
  }
  .opt[aria-checked='true'] .sw::after {
    transform: translateX(18px);
  }

  .tabs {
    display: flex;
    gap: 4px;
    margin: -6px 0 0;
    border-bottom: 1px solid var(--border);
  }
  .tabs button {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 44px;
    padding: 0 12px;
    margin-bottom: -1px;
    border: 0;
    border-bottom: 2px solid transparent;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-weight: 600;
    cursor: pointer;
  }
  .tabs button[aria-selected='true'] {
    color: var(--text);
    border-bottom-color: var(--accent);
  }
  .tabs button:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }
  .badge {
    padding: 1px 8px;
    border-radius: 10px;
    background: color-mix(in srgb, var(--ok) 18%, transparent);
    color: var(--ok);
    font-size: 11px;
    font-weight: 700;
  }
  .badge.off {
    background: var(--icon-bg);
    color: var(--muted);
  }
  .panel {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .panel[hidden] {
    display: none;
  }
  .namerow {
    display: flex;
    align-items: flex-end;
    gap: 10px;
  }
  .namerow :global(.field) {
    flex: 1;
    min-width: 0;
  }
  .seg {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 3px;
    padding: 3px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--bg);
  }
  .seg button {
    height: 38px;
    border: 0;
    border-radius: 7px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
  }
  .seg button[aria-pressed='true'] {
    background: var(--surface);
    color: var(--text);
    box-shadow: 0 0 0 1px var(--border);
  }
  .seg button:focus-visible {
    outline: 2px solid var(--accent);
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
    position: sticky;
    bottom: 0;
    padding-top: 12px;
    background: var(--surface);
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
  @media (max-width: 440px) {
    .buttons {
      flex: 1;
    }
    .buttons .button[type='submit'] {
      flex: 1;
    }
  }
</style>
