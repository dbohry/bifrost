export type View = 'grouped' | 'flat';

const key = 'bifrost.view';

function read(): View {
  try {
    return localStorage.getItem(key) === 'flat' ? 'flat' : 'grouped';
  } catch {
    return 'grouped';
  }
}

/**
 * How the dashboard lays out the apps: split by category, or one list in the
 * order the user left it. It is kept on this device, like a sort order, and
 * the page works without it.
 */
class ViewChoice {
  current = $state<View>(read());

  set(view: View) {
    this.current = view;
    try {
      localStorage.setItem(key, view);
    } catch {
      // private window or blocked storage: the choice just lasts until the page closes
    }
  }
}

export const view = new ViewChoice();
