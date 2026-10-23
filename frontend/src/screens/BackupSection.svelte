<script lang="ts">
  import ConfirmDialog from '../components/ConfirmDialog.svelte';
  import { api, ApiError, type ExportFile, type ImportResult } from '../lib/api';
  import { appearance } from '../lib/appearance.svelte';
  import { apps } from '../lib/apps.svelte';

  let picked = $state<{ name: string; file: ExportFile } | null>(null);
  let readError = $state('');
  let mode = $state<'merge' | 'replace'>('merge');
  let withAppearance = $state(true);
  let busy = $state(false);
  let error = $state('');
  let result = $state<ImportResult | null>(null);
  let confirming = $state(false);
  let input = $state<HTMLInputElement>();

  async function choose(e: Event) {
    readError = error = '';
    result = null;
    picked = null;
    const file = (e.currentTarget as HTMLInputElement).files?.[0];
    if (!file) return;
    if (file.size > 2 << 20) {
      readError = 'That file is too big to be a Bifrost export.';
      return;
    }
    try {
      const data = JSON.parse(await file.text());
      if (data?.format !== 'bifrost-export' || !Array.isArray(data.apps)) throw new Error('not an export');
      picked = { name: file.name, file: data };
    } catch {
      readError = 'That is not a Bifrost export file.';
    }
  }

  async function run() {
    if (!picked) return;
    busy = true;
    error = '';
    try {
      result = await api.importData(picked.file, mode, withAppearance);
      await Promise.all([apps.load(), result.appearanceApplied ? appearance.load() : Promise.resolve()]);
      picked = null;
      if (input) input.value = '';
    } catch (e) {
      error = e instanceof ApiError ? e.message : 'Something went wrong.';
    } finally {
      busy = false;
    }
  }

  function submit(e: SubmitEvent) {
    e.preventDefault();
    if (mode === 'replace') confirming = true;
    else run();
  }

  const count = $derived(picked?.file.apps.length ?? 0);
  const plural = (n: number) => `${n} app${n === 1 ? '' : 's'}`;
</script>

<section>
  <h1>Backup</h1>
  <p class="intro">Keep a copy of your apps and appearance, or move them to another Bifrost.</p>

  <div class="card">
    <div class="head">
      <span class="badge" aria-hidden="true">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 4v11M7 11l5 5 5-5M5 20h14" /></svg>
      </span>
      <div>
        <h2>Export</h2>
        <p>Downloads a file with your apps, their order and your appearance. It holds no passwords, and nothing about other users.</p>
      </div>
    </div>
    <a class="button action" href="/api/export" download="bifrost-export.json">Download export</a>
  </div>

  <form class="card" onsubmit={submit}>
    <div class="head">
      <span class="badge" aria-hidden="true">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 16V5M7 9l5-5 5 5M5 20h14" /></svg>
      </span>
      <div>
        <h2>Import</h2>
        <p>Load a file you exported from Bifrost. Only your own dashboard is changed.</p>
      </div>
    </div>

    <label class="drop" class:filled={picked}>
      <input bind:this={input} type="file" accept=".json,application/json" onchange={choose} />
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z" /><path d="M14 3v5h5" /></svg>
      <span class="drop-text">
        <strong>{picked ? picked.name : 'Choose an export file'}</strong>
        <span>{picked ? `${plural(count)}${picked.file.preferences ? ' and an appearance' : ''}. Click to choose another.` : 'A .json file from Settings → Backup'}</span>
      </span>
    </label>

    {#if readError}
      <div class="alert" role="alert"><span class="mark" aria-hidden="true">!</span>{readError}</div>
    {/if}

    {#if picked}
      <p class="sr-only" role="status">{picked.name} holds {plural(count)}{picked.file.preferences ? ' and an appearance' : ''}.</p>

      <fieldset>
        <legend>What should happen to the apps you have now?</legend>
        <label class="choice" class:on={mode === 'merge'}>
          <input type="radio" name="mode" value="merge" bind:group={mode} />
          <span><strong>Add to my dashboard</strong><br />Apps already there, with the same name and address, are skipped.</span>
        </label>
        <label class="choice" class:on={mode === 'replace'}>
          <input type="radio" name="mode" value="replace" bind:group={mode} />
          <span><strong>Replace my dashboard</strong><br />Your current apps are deleted first.</span>
        </label>
      </fieldset>

      {#if picked.file.preferences}
        <label class="check">
          <input type="checkbox" bind:checked={withAppearance} />
          Also restore the theme and logo from the file
        </label>
      {/if}

      {#if error}
        <div class="alert" role="alert"><span class="mark" aria-hidden="true">!</span>{error}</div>
      {/if}

      <div class="buttons">
        <button class="button" type="submit" disabled={busy}>{busy ? 'Importing…' : 'Import'}</button>
      </div>
    {/if}

    {#if result}
      <div class="done" role="status">
        Imported {plural(result.added)}{result.skipped ? `, skipped ${result.skipped} already there` : ''}{result.removed
          ? `, removed ${plural(result.removed)} you had before`
          : ''}{result.appearanceApplied ? ', and restored your appearance' : ''}.
      </div>
    {/if}
  </form>
</section>

{#if confirming}
  <ConfirmDialog
    title="Replace your dashboard"
    confirmLabel="Replace my apps"
    danger
    onclose={() => (confirming = false)}
    onconfirm={async () => {
      confirming = false;
      await run();
    }}
  >
    <p>All {plural(apps.items.length)} on your dashboard will be deleted and replaced by the {plural(count)} in the file. This cannot be undone.</p>
  </ConfirmDialog>
{/if}

<style>
  section {
    max-width: 560px;
  }
  h1 {
    margin: 0;
    font-size: 24px;
  }
  .intro {
    margin: 4px 0 24px;
    font-size: 14px;
    color: var(--muted);
  }
  .card {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 18px;
    padding: 20px;
    margin-bottom: 20px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
  }
  .head {
    display: flex;
    gap: 14px;
    align-items: flex-start;
  }
  .badge {
    flex-shrink: 0;
    width: 40px;
    height: 40px;
    border-radius: 12px;
    background: var(--icon-bg);
    color: var(--accent);
    display: grid;
    place-items: center;
  }
  .head h2 {
    margin: 0;
    font-size: 16px;
  }
  .head p {
    margin: 4px 0 0;
    font-size: 14px;
    line-height: 1.5;
    color: var(--muted);
  }
  .button.action {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
    height: 44px;
    text-decoration: none;
    align-self: flex-start;
  }

  /* The real file input stays for keyboards and screen readers, and the label is what you see. */
  .drop {
    position: relative;
    display: flex;
    align-items: center;
    gap: 14px;
    min-height: 72px;
    padding: 14px 16px;
    border: 2px dashed var(--border);
    border-radius: 12px;
    color: var(--muted);
    cursor: pointer;
  }
  .drop:hover {
    border-color: var(--accent);
    color: var(--text);
  }
  .drop:focus-within {
    border-color: var(--accent);
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .drop.filled {
    border-style: solid;
    border-color: var(--accent);
    color: var(--text);
  }
  .drop input {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    opacity: 0;
    cursor: pointer;
  }
  .drop svg {
    flex-shrink: 0;
  }
  .drop-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
    font-size: 13px;
    line-height: 1.4;
  }
  .drop-text strong {
    color: var(--text);
    font-size: 14px;
    overflow-wrap: anywhere;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }

  fieldset {
    margin: 0;
    padding: 0;
    border: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  legend {
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
    padding: 12px 14px;
    border: 1px solid var(--border);
    border-radius: 10px;
    font-size: 13px;
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
  .check {
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: 44px;
    font-size: 14px;
  }
  .check input {
    width: 18px;
    height: 18px;
    margin: 0;
    accent-color: var(--accent);
  }
  .buttons {
    display: flex;
    justify-content: flex-end;
  }
  @media (max-width: 480px) {
    .buttons .button,
    .button.action {
      box-sizing: border-box;
      width: 100%;
      align-self: stretch;
    }
  }
  .alert,
  .done {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    border: 1px solid var(--bad);
    border-radius: 10px;
    font-size: 13px;
    line-height: 1.4;
    color: var(--text);
  }
  .done {
    border-color: var(--ok);
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
</style>
