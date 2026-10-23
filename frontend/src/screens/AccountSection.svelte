<script lang="ts">
  import Field from '../components/Field.svelte';
  import { api, ApiError } from '../lib/api';

  let current = $state('');
  let next = $state('');
  let confirm = $state('');
  let error = $state('');
  let done = $state(false);
  let busy = $state(false);

  const mismatch = $derived(confirm !== '' && next !== confirm);

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    done = false;
    if ([...next].length < 8) {
      error = 'Password must be at least 8 characters.';
      return;
    }
    if (next !== confirm) {
      error = 'The two passwords do not match.';
      return;
    }

    busy = true;
    try {
      await api.changePassword(current, next);
      current = next = confirm = '';
      done = true;
    } catch (err) {
      error = err instanceof ApiError ? err.message : 'Something went wrong.';
    } finally {
      busy = false;
    }
  }
</script>

<section>
  <h1>Account</h1>
  <p class="intro">Change the password you use to sign in.</p>

  <form class="form" onsubmit={submit}>
    {#if error}
      <div class="alert" role="alert"><span class="mark" aria-hidden="true">!</span>{error}</div>
    {/if}
    {#if done}
      <div class="done" role="status">Password changed. You were signed out on your other devices.</div>
    {/if}

    <Field label="Current password" type="password" bind:value={current} autocomplete="current-password" />
    <Field label="New password" type="password" bind:value={next} autocomplete="new-password" hint="At least 8 characters." />
    <Field label="Confirm new password" type="password" bind:value={confirm} autocomplete="new-password" invalid={mismatch} />
    <div class="buttons">
      <button class="button" type="submit" disabled={busy}>{busy ? 'Saving…' : 'Change password'}</button>
    </div>
  </form>
</section>

<style>
  section {
    max-width: 420px;
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
