import { api, ApiError, type AppInput, type AppItem } from './api';

export type Group = { name: string; apps: AppItem[] };

/** Groups apps by category, in the order the first app of each category appears. */
export function groupByCategory(items: AppItem[]): Group[] {
  const byName = new Map<string, AppItem[]>();
  for (const app of items) {
    const list = byName.get(app.category);
    if (list) list.push(app);
    else byName.set(app.category, [app]);
  }
  return [...byName].map(([name, apps]) => ({ name, apps }));
}

/** The signed-in user's apps. */
class Apps {
  items = $state<AppItem[]>([]);
  loaded = $state(false);
  error = $state('');

  favourites = $derived(this.items.filter((a) => a.favourite));

  groups = $derived(groupByCategory(this.items));

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

  /**
   * Moves the app at index `from` of a category to index `to`, and saves the new
   * order. Only the apps of that category swap places, so the other categories
   * stay where they are. The screen changes straight away and goes back if the
   * server refuses.
   */
  async move(category: string, from: number, to: number) {
    if (from === to) return;
    const slots = this.items.flatMap((a, i) => (a.category === category ? [i] : []));
    if (from < 0 || to < 0 || from >= slots.length || to >= slots.length) return;

    const inCategory = slots.map((i) => this.items[i]);
    const [moved] = inCategory.splice(from, 1);
    inCategory.splice(to, 0, moved);

    const next = [...this.items];
    slots.forEach((slot, k) => (next[slot] = inCategory[k]));

    await this.commitOrder(next);
  }

  /**
   * Moves the app at index `from` of the whole dashboard to index `to`, and
   * saves the order. This is the move for the single list that is not split
   * by category.
   */
  async moveAll(from: number, to: number) {
    if (from === to || from < 0 || to < 0 || from >= this.items.length || to >= this.items.length) return;
    const next = [...this.items];
    const [moved] = next.splice(from, 1);
    next.splice(to, 0, moved);

    await this.commitOrder(next);
  }

  /** Shows a new order straight away and saves it, putting the old one back if the server refuses. */
  private async commitOrder(next: AppItem[]) {
    const before = this.items;
    this.items = next.map((a, i) => ({ ...a, position: i + 1 }));
    try {
      await api.reorderApps(next.map((a) => a.id));
    } catch (e) {
      this.items = before;
      this.error = e instanceof Error ? e.message : 'Could not save the new order.';
      // Another window may have changed the list. Fetch it again.
      if (e instanceof ApiError && e.status === 409) await this.refresh();
    }
  }

  dismissError() {
    this.error = '';
  }

  private async refresh() {
    try {
      this.items = (await api.apps()).apps;
    } catch {
      // the error from the failed save stays on screen
    }
  }

  clear() {
    this.items = [];
    this.loaded = false;
    this.error = '';
  }
}

export const apps = new Apps();
