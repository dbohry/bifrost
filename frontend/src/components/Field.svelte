<script lang="ts">
  type Props = {
    label: string;
    value: string;
    type?: 'text' | 'password';
    autocomplete?: AutoFill;
    hint?: string;
    invalid?: boolean;
    list?: string;
    placeholder?: string;
    /** Take focus when the surrounding dialog opens. */
    focusFirst?: boolean;
    input?: HTMLInputElement;
  };

  let {
    label,
    value = $bindable(''),
    type = 'text',
    autocomplete,
    hint,
    invalid = false,
    list,
    placeholder,
    focusFirst = false,
    input = $bindable(),
  }: Props = $props();

  const id = $props.id();
  let shown = $state(false);
</script>

<div class="field">
  <label for={id}>{label}</label>
  <div class="box" class:invalid>
    <input
      {id}
      bind:this={input}
      bind:value
      type={type === 'password' && shown ? 'text' : type}
      {autocomplete}
      {list}
      {placeholder}
      data-autofocus={focusFirst ? '' : undefined}
      aria-invalid={invalid}
      aria-describedby={hint ? `${id}-hint` : undefined}
      autocapitalize="off"
      spellcheck="false"
    />
    {#if type === 'password'}
      <button
        type="button"
        class="toggle"
        aria-label={shown ? 'Hide password' : 'Show password'}
        aria-pressed={shown}
        onclick={() => (shown = !shown)}
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7S2 12 2 12z" />
          <circle cx="12" cy="12" r="3" />
          {#if shown}<path d="M4 4l16 16" />{/if}
        </svg>
      </button>
    {/if}
  </div>
  {#if hint}<span class="hint" id="{id}-hint">{hint}</span>{/if}
</div>

<style>
  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  label {
    font-size: 13px;
    font-weight: 600;
  }
  .box {
    display: flex;
    align-items: center;
    height: 48px;
    box-sizing: border-box;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--bg);
  }
  .box:focus-within {
    border-color: var(--accent);
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }
  .box.invalid {
    border-color: var(--bad);
  }
  input {
    flex: 1;
    min-width: 0;
    height: 100%;
    box-sizing: border-box;
    padding: 0 14px;
    border: 0;
    background: transparent;
    color: var(--text);
    font: inherit;
    font-size: 15px;
    outline: none;
  }
  .toggle {
    width: 44px;
    height: 44px;
    border: 0;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    display: grid;
    place-items: center;
  }
  .hint {
    font-size: 12px;
    color: var(--muted);
  }
</style>
