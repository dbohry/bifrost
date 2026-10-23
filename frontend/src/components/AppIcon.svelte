<script lang="ts">
  import { iconUrl } from '../lib/api';
  import { initial, tileColour } from '../lib/tile';

  type Props = {
    /** The app's name, for the letter shown when there is no icon. */
    name: string;
    /** Picks the tile colour, and keeps it steady if the app is renamed. */
    seed: number;
    /** The catalogue icon, or "" for the first letter. */
    icon?: string;
    size?: number;
  };

  let { name, seed, icon = '', size = 52 }: Props = $props();

  // If an icon cannot be loaded, show the letter instead of a broken picture.
  let broken = $state(false);
  $effect(() => {
    icon;
    broken = false;
  });
</script>

{#if icon && !broken}
  <!-- Catalogue logos are drawn for light backgrounds, so they sit on a white tile in every theme. -->
  <span
    class="icon art"
    style:width="{size}px"
    style:height="{size}px"
    style:padding="{Math.round(size * 0.14)}px"
    aria-hidden="true"
  >
    <img src={iconUrl(icon)} alt="" loading="lazy" decoding="async" draggable="false" onerror={() => (broken = true)} />
  </span>
{:else}
  <span
    class="icon"
    style:width="{size}px"
    style:height="{size}px"
    style:background={tileColour(seed)}
    style:font-size="{Math.round(size * 0.46)}px"
    aria-hidden="true">{initial(name)}</span
  >
{/if}

<style>
  .icon {
    flex-shrink: 0;
    border-radius: 27%;
    display: grid;
    place-items: center;
    color: #fff;
    font-weight: 700;
    box-sizing: border-box;
  }
  /*
   * A plain block, not a grid: these SVGs have no width or height of their own,
   * so inside a shrink-to-fit grid cell "100%" works out to nothing and the logo
   * is drawn 0 by 0. In a block with a fixed size, 100% is well defined.
   *
   * The padding is set from the size in the markup above. A percentage here would
   * be measured against the parent's width, not the icon's own.
   */
  .art {
    display: block;
    background: #fff;
    box-shadow: inset 0 0 0 1px rgb(0 0 0 / 0.08);
  }
  .art img {
    width: 100%;
    height: 100%;
    object-fit: contain;
    display: block;
  }
</style>
