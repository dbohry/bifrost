<script lang="ts">
  import { onMount, type Snippet } from 'svelte';

  type Props = {
    title: string;
    /** Called once the dialog has closed, however it was closed. */
    onclose: () => void;
    children: Snippet;
  };

  let { title, onclose, children }: Props = $props();

  let dialog = $state<HTMLDialogElement>();
  const uid = $props.id();
  const titleId = `${uid}-title`;

  /** Closes the dialog. The parent's `onclose` runs afterwards. */
  export function close() {
    dialog?.close();
  }

  onMount(() => {
    dialog?.showModal();
    // The browser focuses the first control, which is the close button.
    // A field marked data-autofocus gets focus instead.
    dialog?.querySelector<HTMLElement>('[data-autofocus]')?.focus();
  });

  // Clicking the dimmed area around the dialog closes it, like Escape does.
  function backdrop(e: MouseEvent) {
    if (e.target === dialog) dialog?.close();
  }
</script>

<dialog bind:this={dialog} aria-labelledby={titleId} {onclose} onclick={backdrop}>
  <div class="sheet">
    <span class="handle" aria-hidden="true"></span>
    <header>
      <h2 id={titleId}>{title}</h2>
      <button class="close" type="button" aria-label="Close" onclick={close}>
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M6 6l12 12M18 6L6 18" /></svg>
      </button>
    </header>
    {@render children()}
  </div>
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

  .handle {
    display: none;
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
      padding-left: max(24px, var(--safe-left));
      padding-right: max(24px, var(--safe-right));
    }
    .handle {
      display: block;
      width: 40px;
      height: 5px;
      margin: -8px auto -6px;
      border-radius: 3px;
      background: var(--border);
    }
  }
</style>
