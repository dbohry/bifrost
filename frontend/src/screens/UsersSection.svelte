<script lang="ts">
  import AddUserDialog from '../components/AddUserDialog.svelte';
  import ConfirmDialog from '../components/ConfirmDialog.svelte';
  import TempPasswordDialog from '../components/TempPasswordDialog.svelte';
  import type { UserSummary } from '../lib/api';
  import { session } from '../lib/session.svelte';
  import { lastSeen } from '../lib/time';
  import { users } from '../lib/users.svelte';

  type Dialog =
    | { kind: 'add' }
    | { kind: 'reset'; user: UserSummary }
    | { kind: 'remove'; user: UserSummary }
    | { kind: 'password'; username: string; password: string; created: boolean };

  let dialog = $state<Dialog | null>(null);

  $effect(() => {
    users.load();
    return () => users.clear();
  });

  const close = () => (dialog = null);
</script>

<section>
  <div class="head">
    <div>
      <h1>Users</h1>
      <p class="intro">Everyone who can sign in to this Bifrost. Each person has their own apps.</p>
    </div>
    <button class="button" onclick={() => (dialog = { kind: 'add' })}>+ Add user</button>
  </div>

  {#if users.error}
    <div class="alert" role="alert">
      {users.error}
      <button class="button secondary" onclick={() => users.load()}>Try again</button>
    </div>
  {:else if !users.loaded}
    <p class="muted" role="status">Loading users…</p>
  {:else}
    <ul class="list">
      {#each users.items as u (u.id)}
        {@const you = u.id === session.user?.id}
        <li>
          <span class="avatar" aria-hidden="true">{u.username[0].toUpperCase()}</span>
          <span class="who">
            <span class="name">{you ? `${u.username} (you)` : u.username}</span>
            <span class="seen">{lastSeen(u.lastSeenAt)}</span>
          </span>
          <span class="chips">
            {#if u.mustChangePassword}<span class="chip temp">Temporary password</span>{/if}
            <span class="chip" class:admin={u.isAdmin}>{u.isAdmin ? 'Admin' : 'User'}</span>
          </span>
          {#if !you}
            <span class="row-actions">
              <button class="button secondary" onclick={() => (dialog = { kind: 'reset', user: u })}>Reset password</button>
              <button class="remove" aria-label={`Remove ${u.username}`} onclick={() => (dialog = { kind: 'remove', user: u })}>
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M6 6l12 12M18 6L6 18" /></svg>
              </button>
            </span>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}

  <aside class="locked-out">
    <h2>Locked out, or forgot a password?</h2>
    <p>
      Anyone can be given a new temporary password above. If the administrator is the one locked out, there is no
      reset on the sign-in page. Reset it from the machine that runs Bifrost instead. This needs access to that
      machine, so only the person who runs it can do it.
    </p>
    <pre><code>docker exec bifrost /bifrost reset-password &lt;username&gt;</code></pre>
    <ul>
      <li>It prints a temporary password. The user must choose a new one the next time they sign in.</li>
      <li>Sessions on all devices are signed out after a reset.</li>
    </ul>
  </aside>
</section>

{#if dialog?.kind === 'add'}
  <AddUserDialog
    onclose={close}
    oncreate={async (username, isAdmin) => {
      const password = await users.create(username, isAdmin);
      dialog = { kind: 'password', username, password, created: true };
    }}
  />
{:else if dialog?.kind === 'reset'}
  {@const target = dialog.user}
  <ConfirmDialog
    title="Reset password"
    confirmLabel="Reset password"
    onclose={close}
    onconfirm={async () => {
      const password = await users.resetPassword(target.id);
      dialog = { kind: 'password', username: target.username, password, created: false };
    }}
  >
    <p>Give <strong>{target.username}</strong> a new temporary password?</p>
    <p>They will be signed out on every device, and their current password will stop working.</p>
  </ConfirmDialog>
{:else if dialog?.kind === 'remove'}
  {@const target = dialog.user}
  <ConfirmDialog
    title="Remove user"
    confirmLabel="Remove user"
    danger
    onclose={close}
    onconfirm={async () => {
      await users.remove(target.id);
      close();
    }}
  >
    <p>Remove <strong>{target.username}</strong>?</p>
    <p>Their account and all of their apps are deleted. This cannot be undone.</p>
  </ConfirmDialog>
{:else if dialog?.kind === 'password'}
  <TempPasswordDialog username={dialog.username} password={dialog.password} created={dialog.created} onclose={close} />
{/if}

<style>
  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 24px;
  }
  h1 {
    margin: 0;
    font-size: 24px;
  }
  .intro,
  .muted {
    margin: 4px 0 0;
    font-size: 14px;
    color: var(--muted);
  }

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    overflow: hidden;
  }
  .list li {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px 14px;
    padding: 14px 18px;
    border-bottom: 1px solid var(--border);
  }
  .list li:last-child {
    border-bottom: 0;
  }
  .avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: var(--icon-bg);
    color: var(--muted);
    display: grid;
    place-items: center;
    font-weight: 700;
  }
  .who {
    flex: 1 1 160px;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .name {
    font-size: 15px;
    font-weight: 600;
    overflow-wrap: anywhere;
  }
  .seen {
    font-size: 12px;
    color: var(--muted);
  }
  .chips {
    display: flex;
    gap: 6px;
  }
  .chip {
    height: 24px;
    padding: 0 10px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    font-size: 12px;
    font-weight: 600;
    background: var(--icon-bg);
    color: var(--text);
    white-space: nowrap;
  }
  .chip.admin {
    background: var(--text);
    color: var(--bg);
  }
  .chip.temp {
    background: transparent;
    border: 1px solid var(--border);
    color: var(--muted);
  }
  .row-actions {
    display: flex;
    gap: 8px;
  }
  .row-actions .button {
    height: 44px;
    font-size: 13px;
    padding: 0 14px;
  }
  .remove {
    width: 44px;
    height: 44px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    display: grid;
    place-items: center;
  }
  .remove:hover {
    color: var(--bad);
    border-color: var(--bad);
  }
  .remove:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .locked-out {
    margin-top: 28px;
    padding: 20px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    font-size: 14px;
    line-height: 1.5;
  }
  .locked-out h2 {
    margin: 0 0 8px;
    font-size: 16px;
  }
  .locked-out p {
    margin: 0 0 12px;
    color: var(--muted);
  }
  .locked-out ul {
    margin: 12px 0 0;
    padding-left: 18px;
    color: var(--muted);
    font-size: 13px;
  }
  pre {
    margin: 0;
    padding: 12px 16px;
    border-radius: 10px;
    background: #1f2430;
    color: #e8eaf0;
    font-size: 13px;
    overflow-x: auto;
  }
  .alert {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px;
    border: 1px solid var(--bad);
    border-radius: 10px;
    font-size: 14px;
  }
</style>
