<script lang="ts">
  import type { Snippet } from 'svelte';
  import Modal from './Modal.svelte';
  import { ApiError } from '../lib/api';

  type Props = {
    title: string;
    confirmLabel: string;
    /** Styles the confirm button as destructive. */
    danger?: boolean;
    /** Runs when confirmed. The parent decides what to show next. */
    onconfirm: () => Promise<void>;
    onclose: () => void;
    children: Snippet;
  };

  let { title, confirmLabel, danger = false, onconfirm, onclose, children }: Props = $props();

  let modal = $state<Modal>();
  let busy = $state(false);
  let error = $state('');

  async function confirm() {
    error = '';
    busy = true;
    try {
      await onconfirm();
    } catch (e) {
      error = e instanceof ApiError ? e.message : 'Something went wrong.';
      busy = false;
    }
  }
</script>

<Modal bind:this={modal} {title} {onclose}>
  <div class="body">
    {#if error}
      <div class="alert" role="alert">
        <span class="mark" aria-hidden="true">!</span>
        {error}
      </div>
    {/if}
    <div class="message">{@render children()}</div>
    <div class="buttons">
      <button class="button secondary" type="button" disabled={busy} data-autofocus onclick={() => modal?.close()}>Cancel</button>
      <button class="button" class:danger type="button" disabled={busy} onclick={confirm}>
        {busy ? 'Working…' : confirmLabel}
      </button>
    </div>
  </div>
</Modal>

<style>
  .body {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .message {
    font-size: 14px;
    line-height: 1.5;
  }
  .message :global(p) {
    margin: 0 0 8px;
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
