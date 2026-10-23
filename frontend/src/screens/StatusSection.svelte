<script lang="ts">
  import { api, ApiError } from '../lib/api';

  // The usual choices, in seconds, and "Custom" for anything else.
  const intervals = [60, 120, 300];
  const timeouts = [1, 5, 10, 30];

  const label = (seconds: number) => (seconds < 60 ? `${seconds} s` : `${seconds / 60} min`);

  let loaded = $state(false);
  let interval = $state<number | 'custom'>(60);
  let timeout = $state<number | 'custom'>(5);
  let customInterval = $state('');
  let customTimeout = $state('');
  let error = $state('');
  let saved = $state(false);
  let busy = $state(false);

  $effect(() => {
    api
      .statusSettings()
      .then(({ settings }) => {
        if (intervals.includes(settings.interval)) interval = settings.interval;
        else {
          interval = 'custom';
          customInterval = String(settings.interval);
        }
        if (timeouts.includes(settings.timeout)) timeout = settings.timeout;
        else {
          timeout = 'custom';
          customTimeout = String(settings.timeout);
        }
        loaded = true;
      })
      .catch((e) => (error = e instanceof Error ? e.message : 'Cannot load the settings.'));
  });

  const seconds = (choice: number | 'custom', custom: string) => (choice === 'custom' ? Number(custom) : choice);

  async function save(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    saved = false;
    const every = seconds(interval, customInterval);
    const wait = seconds(timeout, customTimeout);
    if (!Number.isInteger(every) || every < 10 || every > 86400) {
      error = 'Check every: enter a whole number of seconds from 10 to 86400 (24 hours).';
      return;
    }
    if (!Number.isInteger(wait) || wait < 1 || wait > 60) {
      error = 'Give up after: enter a whole number of seconds from 1 to 60.';
      return;
    }
    if (wait > every) {
      error = 'Give up after cannot be longer than the time between checks.';
      return;
    }
    busy = true;
    try {
      await api.saveStatusSettings({ interval: every, timeout: wait });
      saved = true;
    } catch (err) {
      error = err instanceof ApiError ? err.message : 'Something went wrong.';
    } finally {
      busy = false;
    }
  }
</script>

<section>
  <h1>Status checks</h1>
  <p class="intro">
    Apps with <strong>Check if this app is online</strong> switched on get a badge. Bifrost, not your browser, asks each
    app for a reply, so the answer is the same on every device. An app is shown as offline after two failed checks in a row.
  </p>

  {#if loaded}
    <form onsubmit={save} novalidate>
      {#if error}
        <div class="alert" role="alert"><span class="mark" aria-hidden="true">!</span>{error}</div>
      {/if}
      {#if saved}
        <div class="done" role="status">Saved. It applies from the next check.</div>
      {/if}

      <fieldset>
        <legend>Check every</legend>
        <div class="segmented">
          {#each intervals as v (v)}
            <label class:on={interval === v}><input type="radio" name="interval" value={v} bind:group={interval} />{label(v)}</label>
          {/each}
          <label class:on={interval === 'custom'}><input type="radio" name="interval" value="custom" bind:group={interval} />Custom</label>
        </div>
        {#if interval === 'custom'}
          <label class="custom">
            <span>Seconds</span>
            <input type="number" inputmode="numeric" min="10" max="86400" step="1" bind:value={customInterval} />
          </label>
        {/if}
        <p class="hint">Between 10 seconds and 24 hours.</p>
      </fieldset>

      <fieldset>
        <legend>Give up after</legend>
        <div class="segmented">
          {#each timeouts as v (v)}
            <label class:on={timeout === v}><input type="radio" name="timeout" value={v} bind:group={timeout} />{label(v)}</label>
          {/each}
          <label class:on={timeout === 'custom'}><input type="radio" name="timeout" value="custom" bind:group={timeout} />Custom</label>
        </div>
        {#if timeout === 'custom'}
          <label class="custom">
            <span>Seconds</span>
            <input type="number" inputmode="numeric" min="1" max="60" step="1" bind:value={customTimeout} />
          </label>
        {/if}
        <p class="hint">How long to wait for an answer, from 1 to 60 seconds. A slow app needs more.</p>
      </fieldset>

      <div class="buttons">
        <button class="button" type="submit" disabled={busy}>{busy ? 'Saving…' : 'Save'}</button>
      </div>
    </form>
  {:else if error}
    <div class="alert" role="alert"><span class="mark" aria-hidden="true">!</span>{error}</div>
  {:else}
    <p class="intro" role="status">Loading…</p>
  {/if}
</section>

<style>
  section {
    max-width: 520px;
  }
  h1 {
    margin: 0;
    font-size: 24px;
  }
  .intro {
    margin: 4px 0 24px;
    font-size: 14px;
    line-height: 1.5;
    color: var(--muted);
  }
  form {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  fieldset {
    margin: 0;
    padding: 0;
    border: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  legend {
    padding: 0;
    margin-bottom: 10px;
    font-size: 13px;
    font-weight: 600;
  }
  .segmented {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .segmented label {
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 64px;
    height: 44px;
    padding: 0 16px;
    box-sizing: border-box;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--surface);
    color: var(--muted);
    font-size: 14px;
    cursor: pointer;
  }
  .segmented label.on {
    border-color: var(--accent);
    background: var(--accent);
    color: var(--on-accent);
    font-weight: 600;
  }
  .segmented label:focus-within {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .segmented input {
    position: absolute;
    opacity: 0;
    width: 1px;
    height: 1px;
  }
  .custom {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 13px;
    font-weight: 600;
  }
  .custom input {
    width: 120px;
    height: 44px;
    box-sizing: border-box;
    padding: 0 12px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--bg);
    color: var(--text);
    font: inherit;
    font-size: 16px;
  }
  .custom input:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }
  .hint {
    margin: 0;
    font-size: 12px;
    color: var(--muted);
  }
  .buttons {
    display: flex;
    justify-content: flex-end;
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
