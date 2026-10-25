export type User = {
  id: number;
  username: string;
  isAdmin: boolean;
  mustChangePassword: boolean;
};

/** A colour theme. `colors` has the ten tokens: bg, surface, border, text, muted, accent, onAccent, ok, bad, iconBg. */
export type Theme = {
  id: string;
  name: string;
  mode: 'light' | 'dark';
  colors: Record<string, string>;
};

export type Mode = 'light' | 'dark' | 'system';
export type LogoKind = 'app' | 'arches' | 'spectrum';
export type BackgroundKind = 'plain' | 'glow' | 'bridge' | 'wash' | 'grid';

/** How a user has set up the look of their dashboard. */
export type Preferences = {
  mode: Mode;
  lightTheme: string;
  darkTheme: string;
  logo: LogoKind;
  background: BackgroundKind;
};

/** A row on the Users screen. */
export type UserSummary = {
  id: number;
  username: string;
  isAdmin: boolean;
  mustChangePassword: boolean;
  /** Unix seconds, or null if the user has never signed in. */
  lastSeenAt: number | null;
};

/** A preset app in the built-in catalogue. */
export type CatalogEntry = {
  id: string;
  name: string;
  category: string;
  /** The port the app usually listens on. 80 means the plain web port. */
  port: number;
};

/** The address of a catalogue icon. */
export const iconUrl = (id: string) => `/api/catalog/icons/${encodeURIComponent(id)}.svg`;

/** An app tile on a user's dashboard. */
export type AppItem = {
  id: number;
  name: string;
  url: string;
  category: string;
  favourite: boolean;
  newTab: boolean;
  /** The catalogue id of the icon to show, or "" for the first letter of the name. */
  icon: string;
  position: number;
  /** Status check: look at the app now and then and show a badge on its icon. */
  checkEnabled: boolean;
  /** The address to look at. Empty means the app's own address. */
  checkUrl: string;
  /** Only an HTTP 2xx answer counts as online. Otherwise any answer does. */
  checkStrict: boolean;
  /** Accept a self-signed certificate. */
  checkInsecure: boolean;
};

/** What the user fills in to add or change an app. */
export type AppInput = Omit<AppItem, 'id' | 'position'>;

/** What the latest look at an app found. Times are in seconds since 1970. */
export type StatusResult = {
  state: 'up' | 'down' | 'checking';
  since: number;
  lastChecked: number;
  latencyMs: number;
  reason: string;
  address: string;
};

/** How often apps are checked and how long to wait, both in seconds. */
export type StatusSettings = { interval: number; timeout: number };

/** Checks made in one hour: when it began (Unix seconds), how many, and how many failed. */
export type HourBar = { start: number; checks: number; failures: number };
/** The last 24 hours of an app, oldest first, and the share of checks that passed (null before any). */
export type StatusHistory = { hours: HourBar[]; uptime: number | null };

export type TestResult = { up: boolean; latencyMs: number; code: number; reason: string; address: string };

/** An error response from the server. `message` is safe to show to the user. */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
    readonly code?: string,
  ) {
    super(message);
  }
}

let onUnauthorized: () => void = () => {};

/** Called when a request finds that the session is gone (expired or signed out elsewhere). */
export function setUnauthorizedHandler(fn: () => void) {
  onUnauthorized = fn;
}

// Sign-in, setup and the first "who am I" check expect to be told "not signed in".
const expectsUnauthorized = (path: string) =>
  path === '/api/me' || path === '/api/login' || path === '/api/setup';

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  let res: Response;
  try {
    res = await fetch(path, {
      method,
      credentials: 'same-origin',
      headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new ApiError(0, 'Cannot reach the server.');
  }

  if (res.status === 204) return undefined as T;

  const data = await res.json().catch(() => ({}));
  if (res.status === 401 && !expectsUnauthorized(path)) onUnauthorized();
  if (!res.ok) {
    throw new ApiError(res.status, data.error ?? 'Something went wrong.', data.code);
  }
  return data as T;
}

/** A Bifrost export file, as the server writes it and reads it back. */
export type ExportFile = {
  format: string;
  version: number;
  apps: unknown[];
  preferences?: Preferences;
};

export type ImportResult = { added: number; skipped: number; removed: number; appearanceApplied: boolean };

export const api = {
  health: () => request<{ status: string; version: string }>('GET', '/healthz'),
  setupStatus: () => request<{ needsSetup: boolean }>('GET', '/api/setup/status'),
  setup: (username: string, password: string) =>
    request<{ user: User }>('POST', '/api/setup', { username, password }),
  login: (username: string, password: string, remember: boolean) =>
    request<{ user: User }>('POST', '/api/login', { username, password, remember }),
  logout: () => request<void>('POST', '/api/logout'),
  me: () => request<{ user: User }>('GET', '/api/me'),
  catalog: () => request<{ categories: string[]; apps: CatalogEntry[] }>('GET', '/api/catalog'),
  themes: () => request<{ themes: Theme[] }>('GET', '/api/themes'),
  preferences: () => request<{ preferences: Preferences }>('GET', '/api/preferences'),
  savePreferences: (p: Preferences) => request<{ preferences: Preferences }>('PUT', '/api/preferences', p),
  users: () => request<{ users: UserSummary[] }>('GET', '/api/users'),
  createUser: (username: string, isAdmin: boolean) =>
    request<{ user: UserSummary; temporaryPassword: string }>('POST', '/api/users', { username, isAdmin }),
  resetUserPassword: (id: number) =>
    request<{ temporaryPassword: string }>('POST', `/api/users/${id}/reset-password`),
  deleteUser: (id: number) => request<void>('DELETE', `/api/users/${id}`),
  apps: () => request<{ apps: AppItem[] }>('GET', '/api/apps'),
  createApp: (input: AppInput) => request<{ app: AppItem }>('POST', '/api/apps', input),
  updateApp: (id: number, input: AppInput) =>
    request<{ app: AppItem }>('PUT', `/api/apps/${id}`, input),
  deleteApp: (id: number) => request<void>('DELETE', `/api/apps/${id}`),
  /** Saves the dashboard order. `ids` must list every one of the user's apps. */
  reorderApps: (ids: number[]) => request<void>('PUT', '/api/apps/order', { ids }),
  status: () => request<{ statuses: Record<string, StatusResult> }>('GET', '/api/status'),
  statusHistory: (id: number) => request<StatusHistory>('GET', `/api/status/${id}/history`),
  statusSettings: () => request<{ settings: StatusSettings }>('GET', '/api/status/settings'),
  saveStatusSettings: (v: StatusSettings) => request<{ settings: StatusSettings }>('PUT', '/api/status/settings', v),
  testStatus: (url: string, strict: boolean, insecure: boolean) =>
    request<TestResult>('POST', '/api/status/test', { url, strict, insecure }),
  importData: (file: ExportFile, mode: 'merge' | 'replace', appearance: boolean) =>
    request<ImportResult>('POST', '/api/import', { ...file, mode, appearance }),
  changePassword: (current: string, next: string) =>
    request<void>('POST', '/api/me/password', { current, new: next }),
};
