<script lang="ts">
  import Field from './Field.svelte';
  import Modal from './Modal.svelte';
  import { ApiError } from '../lib/api';

  type Props = {
    /** Creates the user. The parent decides what to show next. */
    oncreate: (username: string, isAdmin: boolean) => Promise<void>;
    onclose: () => void;
  };

  let { oncreate, onclose }: Props = $props();

  let modal = $state<Modal>();
  let username = $state('');
  let isAdmin = $state(false);
  let error = $state('');
  let busy = $state(false);

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    busy = true;
    try {
      await oncreate(username.trim(), isAdmin);
    } catch (err) {
      error = err instanceof ApiError ? err.message : 'Something went wrong.';
      busy = false;
    }
  }
</script>

<Modal bind:this={modal} title="Add user" {onclose}>
  <form class="form" onsubmit={submit}>
    {#if error}
      <div class="alert" role="alert">
        <span class="mark" aria-hidden="true">!</span>
        {error}
      </div>
    {/if}

    <Field
      label="Username"
      bind:value={username}
      autocomplete="off"
      hint="3 to 32 characters: letters, numbers, dots, dashes and underscores."
      focusFirst
    />
    <label class="check">
      <input type="checkbox" bind:checked={isAdmin} />
      Administrator (can add and remove users)
    </label>
    <p class="note">A temporary password is made for them and shown once. They choose their own when they first sign in.</p>

    <div class="buttons">
      <button class="button secondary" type="button" disabled={busy} onclick={() => modal?.close()}>Cancel</button>
      <button class="button" type="submit" disabled={busy}>{busy ? 'Adding…' : 'Add user'}</button>
    </div>
  </form>
</Modal>

<style>
  .note {
    margin: 0;
    font-size: 13px;
    line-height: 1.5;
    color: var(--muted);
  }
  .buttons {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
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
</style>
