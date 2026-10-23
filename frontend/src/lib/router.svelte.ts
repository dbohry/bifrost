/** A very small history-based router. Screens read `router.path`. */
class Router {
  path = $state(location.pathname);

  constructor() {
    addEventListener('popstate', () => (this.path = location.pathname));
  }

  navigate(to: string, { replace = false } = {}) {
    if (to === this.path) return;
    if (replace) history.replaceState(null, '', to);
    else history.pushState(null, '', to);
    this.path = to;
  }
}

export const router = new Router();
