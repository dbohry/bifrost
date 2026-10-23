import { api, type UserSummary } from './api';

/** The accounts shown on the Users screen. Only administrators can load them. */
class Users {
  items = $state<UserSummary[]>([]);
  loaded = $state(false);
  error = $state('');

  async load() {
    this.error = '';
    try {
      this.items = (await api.users()).users;
      this.loaded = true;
    } catch (e) {
      this.error = e instanceof Error ? e.message : 'Cannot load the users.';
    }
  }

  /** Adds a user and returns the temporary password, which is shown once. */
  async create(username: string, isAdmin: boolean): Promise<string> {
    const { user, temporaryPassword } = await api.createUser(username, isAdmin);
    this.items = [...this.items, user];
    return temporaryPassword;
  }

  /** Resets a user's password and returns the new temporary one. */
  async resetPassword(id: number): Promise<string> {
    const { temporaryPassword } = await api.resetUserPassword(id);
    this.items = this.items.map((u) => (u.id === id ? { ...u, mustChangePassword: true } : u));
    return temporaryPassword;
  }

  async remove(id: number) {
    await api.deleteUser(id);
    this.items = this.items.filter((u) => u.id !== id);
  }

  clear() {
    this.items = [];
    this.loaded = false;
    this.error = '';
  }
}

export const users = new Users();
