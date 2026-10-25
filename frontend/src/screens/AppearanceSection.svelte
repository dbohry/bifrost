<script lang="ts">
  import BackgroundSample from '../components/BackgroundSample.svelte';
  import Logo from '../components/Logo.svelte';
  import ThemePreview from '../components/ThemePreview.svelte';
  import type { BackgroundKind, LogoKind, Mode, Theme } from '../lib/api';
  import { appearance } from '../lib/appearance.svelte';

  const modes: { value: Mode; label: string }[] = [
    { value: 'light', label: 'Light' },
    { value: 'dark', label: 'Dark' },
    { value: 'system', label: 'System' },
  ];

  const logos: { value: LogoKind; name: string; sub: string }[] = [
    { value: 'app', name: 'App icon', sub: 'Arches on a tile' },
    { value: 'arches', name: 'Arches', sub: 'Bare arches, one colour' },
    { value: 'spectrum', name: 'Spectrum', sub: 'Five-colour rainbow' },
  ];

  const backgrounds: { value: BackgroundKind; name: string; sub: string }[] = [
    { value: 'plain', name: 'Plain', sub: 'Just the theme colour' },
    { value: 'glow', name: 'Glow and dots', sub: 'Accent glows over a dot grid' },
    { value: 'bridge', name: 'Rainbow bridge', sub: 'A soft spectrum arc' },
    { value: 'wash', name: 'Quiet wash', sub: 'An accent tint fading down' },
    { value: 'grid', name: 'Blueprint grid', sub: 'Thin square grid lines' },
  ];

  // The preview shows the theme in use, until you point at or pick another.
  let previewId = $state<string | null>(null);
  const preview = $derived((previewId && appearance.find(previewId)) || appearance.current);

  const swatch = (t: Theme) => [t.colors.bg, t.colors.surface, t.colors.text, t.colors.accent];

  const modeHint = $derived(
    appearance.prefs.mode === 'system'
      ? `Following this device, which is using its ${appearance.systemDark ? 'dark' : 'light'} appearance now.`
      : `Always ${appearance.prefs.mode}, whatever this device uses.`,
  );
</script>

<section>
  <h1>Appearance</h1>
  <p class="intro">Choose how Bifrost looks. Saved for you, and used on every device you sign in on.</p>

  {#if appearance.error}
    <div class="alert" role="alert"><span class="mark" aria-hidden="true">!</span>{appearance.error}</div>
  {/if}

  <fieldset class="group">
    <legend>Mode</legend>
    <div class="segmented">
      {#each modes as m (m.value)}
        <label class:on={appearance.prefs.mode === m.value}>
          <input
            type="radio"
            name="mode"
            value={m.value}
            checked={appearance.prefs.mode === m.value}
            onchange={() => appearance.save({ mode: m.value })}
          />
          {m.label}
        </label>
      {/each}
    </div>
    <p class="hint">{modeHint} Light and dark each remember their own theme.</p>
  </fieldset>

  <div class="themes">
    <div class="lists">
      {#each [{ title: 'Light theme', slot: 'lightTheme' as const, list: appearance.lightThemes }, { title: 'Dark theme', slot: 'darkTheme' as const, list: appearance.darkThemes }] as group (group.slot)}
        <fieldset class="group">
          <legend>
            {group.title}
            {#if (group.slot === 'darkTheme') === appearance.isDark}<span class="inuse">in use now</span>{/if}
          </legend>
          <div class="options">
            {#each group.list as t (t.id)}
              <label class="option" class:on={appearance.prefs[group.slot] === t.id}>
                <input
                  type="radio"
                  name={group.slot}
                  value={t.id}
                  checked={appearance.prefs[group.slot] === t.id}
                  onchange={() => {
                    previewId = t.id;
                    appearance.save({ [group.slot]: t.id });
                  }}
                  onfocus={() => (previewId = t.id)}
                />
                <span class="name">{t.name}</span>
                <span class="swatch" aria-hidden="true">
                  {#each swatch(t) as colour, i (i)}<i style:background={colour}></i>{/each}
                </span>
              </label>
            {/each}
          </div>
        </fieldset>
      {/each}
    </div>

    <div class="preview-pane">
      <h2>Preview</h2>
      {#if preview}
        <ThemePreview theme={preview} />
        <p class="hint">{preview.name}</p>
      {/if}
    </div>
  </div>

  <fieldset class="group">
    <legend>Background</legend>
    <div class="backgrounds">
      {#each backgrounds as b (b.value)}
        <label class="bg-card" class:on={appearance.prefs.background === b.value}>
          <input
            type="radio"
            name="background"
            value={b.value}
            checked={appearance.prefs.background === b.value}
            onchange={() => appearance.save({ background: b.value })}
          />
          <BackgroundSample kind={b.value} />
          <span class="name">{b.name}</span>
          <span class="sub">{b.sub}</span>
          {#if b.value === 'plain'}<span class="tag">Default</span>{/if}
        </label>
      {/each}
    </div>
  </fieldset>

  <fieldset class="group">
    <legend>Logo</legend>
    <div class="logos">
      {#each logos as l (l.value)}
        <label class="logo-card" class:on={appearance.prefs.logo === l.value}>
          <input
            type="radio"
            name="logo"
            value={l.value}
            checked={appearance.prefs.logo === l.value}
            onchange={() => appearance.save({ logo: l.value })}
          />
          <Logo size={56} kind={l.value} decorative />
          <span class="name">{l.name}</span>
          <span class="sub">{l.sub}</span>
          {#if l.value === 'app'}<span class="tag">Default</span>{/if}
        </label>
      {/each}
    </div>
  </fieldset>
</section>

<style>
  section {
    display: flex;
    flex-direction: column;
    gap: 28px;
  }
  h1 {
    margin: 0;
    font-size: 24px;
  }
  .intro {
    margin: -22px 0 0;
    font-size: 14px;
    color: var(--muted);
  }
  fieldset {
    margin: 0;
    padding: 0;
    border: 0;
    min-width: 0;
  }
  legend {
    padding: 0;
    margin-bottom: 10px;
    font-size: 14px;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .hint {
    margin: 10px 0 0;
    font-size: 12px;
    line-height: 1.5;
    color: var(--muted);
  }
  .inuse {
    font-size: 11px;
    font-weight: 600;
    padding: 2px 8px;
    border-radius: 10px;
    background: var(--icon-bg);
    color: var(--text);
  }

  /* The radio buttons are real, but drawn as the cards around them. */
  input[type='radio'] {
    position: absolute;
    opacity: 0;
    width: 1px;
    height: 1px;
  }
  label:has(input:focus-visible) {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .segmented {
    display: inline-flex;
    border: 1px solid var(--border);
    border-radius: 10px;
    overflow: hidden;
    background: var(--surface);
  }
  .segmented label {
    position: relative;
    height: 44px;
    padding: 0 22px;
    display: flex;
    align-items: center;
    font-size: 14px;
    cursor: pointer;
  }
  .segmented label.on {
    background: var(--text);
    color: var(--bg);
    font-weight: 600;
  }

  .themes {
    display: flex;
    flex-wrap: wrap;
    gap: 24px;
    align-items: flex-start;
  }
  .lists {
    flex: 1 1 320px;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 22px;
  }
  .options {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .option {
    position: relative;
    display: flex;
    align-items: center;
    gap: 14px;
    min-height: 52px;
    box-sizing: border-box;
    padding: 8px 14px;
    background: var(--surface);
    border: 2px solid var(--border);
    border-radius: 12px;
    cursor: pointer;
  }
  .option.on {
    border-color: var(--accent);
  }
  .option .name {
    flex: 1;
    font-size: 14px;
    font-weight: 600;
  }
  .swatch {
    display: flex;
    border: 1px solid var(--border);
    border-radius: 8px;
    overflow: hidden;
  }
  .swatch i {
    width: 22px;
    height: 28px;
  }
  .preview-pane {
    flex: 1 1 320px;
    min-width: 0;
  }
  .preview-pane h2 {
    margin: 0 0 10px;
    font-size: 14px;
  }

  .logos,
  .backgrounds {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }
  .logo-card,
  .bg-card {
    position: relative;
    flex: 1 1 150px;
    max-width: 200px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    padding: 18px 12px 14px;
    background: var(--surface);
    border: 2px solid var(--border);
    border-radius: 14px;
    cursor: pointer;
    text-align: center;
  }
  .logo-card.on,
  .bg-card.on {
    border-color: var(--accent);
  }
  .logo-card .name,
  .bg-card .name {
    margin-top: 4px;
    font-size: 14px;
    font-weight: 600;
  }
  .logo-card .sub,
  .bg-card .sub {
    font-size: 12px;
    color: var(--muted);
  }
  .tag {
    margin-top: 2px;
    height: 22px;
    padding: 0 10px;
    border-radius: 11px;
    display: flex;
    align-items: center;
    font-size: 11px;
    font-weight: 600;
    background: var(--icon-bg);
    color: var(--text);
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
