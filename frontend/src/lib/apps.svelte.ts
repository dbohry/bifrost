import { api, type AppInput, type AppItem } from './api';

export type Group = { name: string; apps: AppItem[] };

/** The signed-in user's apps. */
class Apps {
  items = $state<AppItem[]>([]);
  loaded = $state(false);
  error = $state('');

  favourites = $derived(this.items.filter((a) => a.favourite));

  /** Apps grouped by category, in the order the first app of each group appears. */
  groups = $derived.by(() => {
    const byName = new Map<string, AppItem[]>();
    for (const app of this.items) {
      const list = byName.get(app.category);
      if (list) list.push(app);
      else byName.set(app.category, [app]);
    }
    return [...byName].map(([name, apps]): Group => ({ name, apps }));
  });

  categories = $derived(this.groups.map((g) => g.name));

  async load() {
    this.error = '';
    try {
      this.items = (await api.apps()).apps;
      this.loaded = true;
    } catch (e) {
      this.error = e instanceof Error ? e.message : 'Cannot load your apps.';
    }
  }

  async create(input: AppInput) {
    const { app } = await api.createApp(input);
    this.items = [...this.items, app];
  }

  async update(id: number, input: AppInput) {
    const { app } = await api.updateApp(id, input);
    this.items = this.items.map((a) => (a.id === id ? app : a));
  }

  async remove(id: number) {
    await api.deleteApp(id);
    this.items = this.items.filter((a) => a.id !== id);
  }

  /** Flips the star straight away, and puts it back if the server says no. */
  async toggleFavourite(app: AppItem) {
    const before = this.items;
    const { id, position: _p, ...rest } = app;
    this.items = this.items.map((a) => (a.id === id ? { ...a, favourite: !a.favourite } : a));
    try {
      await api.updateApp(id, { ...rest, favourite: !app.favourite });
    } catch (e) {
      this.items = before;
      this.error = e instanceof Error ? e.message : 'Could not update the app.';
    }
  }

  clear() {
    this.items = [];
    this.loaded = false;
    this.error = '';
  }
}

export const apps = new Apps();
