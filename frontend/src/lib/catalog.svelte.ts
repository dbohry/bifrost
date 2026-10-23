import { api, type CatalogEntry } from './api';

/** The built-in list of preset apps. It never changes while the app is open, so it is fetched once. */
class Catalog {
  entries = $state<CatalogEntry[]>([]);
  categories = $state<string[]>([]);
  loaded = $state(false);
  error = $state('');

  async load() {
    if (this.loaded) return;
    this.error = '';
    try {
      const data = await api.catalog();
      this.entries = data.apps;
      this.categories = data.categories;
      this.loaded = true;
    } catch (e) {
      this.error = e instanceof Error ? e.message : 'Cannot load the app list.';
    }
  }

  find(id: string): CatalogEntry | undefined {
    return this.entries.find((e) => e.id === id);
  }
}

export const catalog = new Catalog();
