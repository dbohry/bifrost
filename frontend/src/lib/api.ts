export type User = {
  id: number;
  username: string;
  isAdmin: boolean;
  mustChangePassword: boolean;
};

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
  if (!res.ok) {
    throw new ApiError(res.status, data.error ?? 'Something went wrong.', data.code);
  }
  return data as T;
}

export const api = {
  health: () => request<{ status: string; version: string }>('GET', '/healthz'),
  setupStatus: () => request<{ needsSetup: boolean }>('GET', '/api/setup/status'),
  setup: (username: string, password: string) =>
    request<{ user: User }>('POST', '/api/setup', { username, password }),
  login: (username: string, password: string, remember: boolean) =>
    request<{ user: User }>('POST', '/api/login', { username, password, remember }),
  logout: () => request<void>('POST', '/api/logout'),
  me: () => request<{ user: User }>('GET', '/api/me'),
  changePassword: (current: string, next: string) =>
    request<void>('POST', '/api/me/password', { current, new: next }),
};
