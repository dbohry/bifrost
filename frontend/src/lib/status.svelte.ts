import { api, type StatusResult } from './api';

/** How often the page asks the server for fresh results while it is open and visible. */
const refreshEvery = 15_000;

/**
 * Whether an app added from the list starts with its status check on. It is off, so
 * nothing is contacted until the person asks. Flip it to make the list's apps start on.
 */
export const checkOnForCatalogueApps = false;

/**
 * What the server last found out about each app that has its status check on.
 * The server does the checking; this just asks for the results.
 */
class Statuses {
  results = $state<Record<number, StatusResult>>({});
  private timer: ReturnType<typeof setInterval> | undefined;
  private later: ReturnType<typeof setTimeout> | undefined;

  of(id: number): StatusResult | undefined {
    return this.results[id];
  }

  async load() {
    try {
      const { statuses } = await api.status();
      const next: Record<number, StatusResult> = {};
      for (const [id, r] of Object.entries(statuses)) next[Number(id)] = r;
      this.results = next;
    } catch {
      // Keep what is showing. A badge that is a little old is better than none.
    }
  }

  /** Start asking now and then. Call stop() when the dashboard goes away. */
  start() {
    this.stop();
    this.load();
    this.timer = setInterval(() => document.visibilityState === 'visible' && this.load(), refreshEvery);
    document.addEventListener('visibilitychange', this.visible);
  }

  stop() {
    clearInterval(this.timer);
    clearTimeout(this.later);
    this.timer = undefined;
    document.removeEventListener('visibilitychange', this.visible);
    this.results = {};
  }

  /** After saving an app, the server needs a moment to take its first look. */
  soon() {
    clearTimeout(this.later);
    this.later = setTimeout(() => this.load(), 2500);
  }

  private visible = () => {
    if (document.visibilityState === 'visible') this.load();
  };
}

export const statuses = new Statuses();
