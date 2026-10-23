import { api, type Preferences, type Theme } from './api';
import { setFavicon } from './favicon';

const defaults: Preferences = {
  mode: 'system',
  lightTheme: 'bifrost-light',
  darkTheme: 'bifrost-dark',
  logo: 'app',
};

// The last theme that was applied, kept in this browser so that reloading the
// page paints the right colours straight away instead of flashing the default.
const cacheKey = 'bifrost.theme';

type Applied = { mode: 'light' | 'dark'; colors: Record<string, string> };

/** `onAccent` becomes `--on-accent`. */
const variable = (token: string) => '--' + token.replace(/[A-Z]/g, (c) => '-' + c.toLowerCase());

function paint(applied: Applied) {
  const root = document.documentElement;
  for (const [token, value] of Object.entries(applied.colors)) root.style.setProperty(variable(token), value);
  root.style.colorScheme = applied.mode;
}

function unpaint() {
  const root = document.documentElement;
  root.removeAttribute('style');
}

function remember(applied: Applied | null) {
  try {
    if (applied) localStorage.setItem(cacheKey, JSON.stringify(applied));
    else localStorage.removeItem(cacheKey);
  } catch {
    // Storage can be blocked, for example in a private window. The theme still works.
  }
}

/** Paints the cached theme. Call it before the app is drawn. */
export function applyCachedTheme() {
  try {
    const raw = localStorage.getItem(cacheKey);
    if (raw) paint(JSON.parse(raw) as Applied);
  } catch {
    // ignore a missing or broken cache
  }
}

/** The signed-in user's look: mode, themes and logo. */
class Appearance {
  themes = $state<Theme[]>([]);
  prefs = $state<Preferences>({ ...defaults });
  loaded = $state(false);
  error = $state('');

  private dark = window.matchMedia('(prefers-color-scheme: dark)');
  systemDark = $state(this.dark.matches);

  constructor() {
    // "System" follows the device, including when it switches at sunset.
    this.dark.addEventListener('change', (e) => {
      this.systemDark = e.matches;
      this.apply();
    });
  }

  lightThemes = $derived(this.themes.filter((t) => t.mode === 'light'));
  darkThemes = $derived(this.themes.filter((t) => t.mode === 'dark'));

  /** Whether the screen is light or dark right now. */
  isDark = $derived(this.prefs.mode === 'dark' || (this.prefs.mode === 'system' && this.systemDark));

  /** The theme in use right now. */
  current = $derived(this.find(this.isDark ? this.prefs.darkTheme : this.prefs.lightTheme));

  find(id: string): Theme | undefined {
    return this.themes.find((t) => t.id === id);
  }

  async load() {
    try {
      const [{ themes }, { preferences }] = await Promise.all([api.themes(), api.preferences()]);
      this.themes = themes;
      this.prefs = preferences;
      this.loaded = true;
      this.apply();
    } catch {
      // Keep whatever is showing. The default theme is always there.
    }
  }

  /** Saves a change straight away, and puts it back if the server refuses. */
  async save(change: Partial<Preferences>) {
    const before = this.prefs;
    this.prefs = { ...before, ...change };
    this.error = '';
    this.apply();
    try {
      this.prefs = (await api.savePreferences(this.prefs)).preferences;
    } catch (e) {
      this.prefs = before;
      this.error = e instanceof Error ? e.message : 'Could not save that.';
    }
    this.apply();
  }

  /** Back to the plain default look, for the sign-in screen. */
  reset() {
    this.prefs = { ...defaults };
    this.themes = [];
    this.loaded = false;
    this.error = '';
    unpaint();
    remember(null);
    setFavicon('app', null);
  }

  private apply() {
    const theme = this.current;
    if (!theme) return;
    const applied: Applied = { mode: theme.mode, colors: theme.colors };
    paint(applied);
    remember(applied);
    setFavicon(this.prefs.logo, theme.colors);
  }
}

export const appearance = new Appearance();
