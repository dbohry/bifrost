<script lang="ts">
  import type { Theme } from '../lib/api';

  let { theme }: { theme: Theme } = $props();

  const c = $derived(theme.colors);
  const tiles = [
    ['J', 'Jellyfin', 'online'],
    ['N', 'Nextcloud', 'online'],
    ['P', 'Pi-hole', 'online'],
    ['W', 'WireGuard', 'offline'],
  ];
</script>

<!-- A small picture of the dashboard in the given theme. It is decoration only. -->
<div
  class="preview"
  aria-hidden="true"
  style:background={c.bg}
  style:color={c.text}
  style:border-color={c.border}
>
  <div class="bar" style:background={c.surface} style:border-color={c.border}>
    <span class="logo" style:background={c.accent}></span>
    <b>Bifrost</b>
    <span class="search" style:border-color={c.border} style:background={c.bg} style:color={c.muted}>Search</span>
    <span class="add" style:background={c.accent} style:color={c.onAccent}>+ Add</span>
  </div>
  <div class="body">
    <div class="pills">
      <span style:background={c.text} style:color={c.bg}>All</span>
      <span style:border-color={c.border}>Media</span>
      <span style:border-color={c.border}>Network</span>
    </div>
    <div class="grid">
      {#each tiles as [letter, name, state] (name)}
        <div class="tile" style:background={c.surface} style:border-color={c.border}>
          <span class="icon" style:background={c.iconBg} style:color={c.muted}>{letter}</span>
          <span class="name">{name}</span>
          <span class="state" style:color={c.muted}>
            <i style:background={state === 'online' ? c.ok : c.bad}></i>{state}
          </span>
        </div>
      {/each}
    </div>
  </div>
</div>

<style>
  .preview {
    border: 1px solid;
    border-radius: 12px;
    overflow: hidden;
    font-size: 10px;
  }
  .bar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    border-bottom: 1px solid;
  }
  .logo {
    width: 16px;
    height: 16px;
    border-radius: 5px;
  }
  .search {
    flex: 1;
    padding: 4px 8px;
    border: 1px solid;
    border-radius: 6px;
  }
  .add {
    padding: 4px 10px;
    border-radius: 6px;
    font-weight: 700;
  }
  .body {
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .pills {
    display: flex;
    gap: 6px;
  }
  .pills span {
    padding: 3px 10px;
    border: 1px solid transparent;
    border-radius: 10px;
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 8px;
  }
  .tile {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 10px 4px 8px;
    border: 1px solid;
    border-radius: 8px;
  }
  .icon {
    width: 26px;
    height: 26px;
    border-radius: 7px;
    display: grid;
    place-items: center;
    font-weight: 700;
    font-size: 11px;
  }
  .name {
    font-weight: 600;
  }
  .state {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 9px;
  }
  .state i {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }
</style>
